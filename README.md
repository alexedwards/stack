# Stack

Stack makes it easy to chain your HTTP middleware and handlers together, and to pass request-scoped context between them. It's basically a context-aware version of [Alice](https://github.com/justinas/alice).

*Update 2015-06-22: Stack is now in version 1.0 (and has many improvements). The previous version remains available via [gopkg.in](http://gopkg.in/alexedwards/stack.v0).*

## Features

- An inbuilt request-scoped and concurrency-safe map for sharing data (or *context*) between your HTTP middleware and handlers.
- Make reusable and append-able handler chains in the Alice style.
- Lets you use middleware with the common `func(http.Handler) http.Handler`  pattern.
- Nicely compatible with the standard `net/http` package.
- Straightforward and non-magic code.

## Quickstart

This example illustrates the basic syntax.

```go
package main

import (
  "fmt"
  "github.com/alexedwards/stack"
  "net/http"
)

func main() {
  st := stack.New(tokenMiddleware, stack.Adapt(languageMiddleware)).Then(final)

  http.Handle("/", st)
  http.ListenAndServe(":3000", nil)
}

func tokenMiddleware(ctx *stack.Context, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx.Put("token", "c9e452805dee5044ba520198628abcaa")
    next.ServeHTTP(w, r)
  })
}

func languageMiddleware(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Language", "en-gb")
    next.ServeHTTP(w, r)
  })
}

func final(ctx *stack.Context, w http.ResponseWriter, r *http.Request) {
  token, _ := ctx.Get("token")
  fmt.Fprintf(w, "Token is: %s", token)
}
```

Making a request should return a response like:

```
$ curl -i localhost:3000
HTTP/1.1 200 OK
Content-Language: en-gb
Content-Length: 42
Content-Type: text/plain; charset=utf-8

Token is: c9e452805dee5044ba520198628abcaa
```

## Usage

### Creating a chain

New middleware chains are created using `stack.New`.

```go
stack.New(middlewareOne, middlewareTwo)
```

The `stack.New()` function is variadic, which means you can pass in as many pieces of middleware as you need. The order of the parameters should match the *order that you want them to be executed* during a request (reading left to right).

It accepts any middleware with the signature `func(*stack.Context, http.Handler) http.Handler`. In practice that means your middleware handlers should use the pattern:

```go
func middlewareOne(ctx *stack.Context, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // do something middleware-ish, accessing ctx
    next.ServeHTTP(w, r)
  })
}
```

Alternatively you can use the  `stack.Init()` function to create a new chain with an initial (or *base*) context. [See here for more information](#using-a-base-context).

You can add more middleware to an existing chain using the `Append()` method. This returns a new copy of the chain. For example:

`stack.New(middlewareOne, middlewareTwo).Append(middlewareThree, middlewareFour)`

### Adding a final handler

Handlers are added to the end of a chain using the `Then()` method. This  returns a 'closed' copy of the chain which satisfies the `http.Handler` interface (and means it can be used with Go's `http.DefaultServeMux` and other routers).

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(finalHandler))
```

The `Then()` method accepts handlers that use the following pattern:

```go
func finalHandler(ctx *stack.Context, w http.ResponseWriter, r *http.Request) {
   // do something handler-ish, accessing ctx
}
```

For convenience, `ThenHandler()` and `ThenHandlerFunc()` methods are also provided. These allow you to finish a chain with a standard `http.Handler` or `http.HandlerFunc` respectively. 

For example, you could use a standard `http.FileServer` as the handler like so:

```go
fs :=  http.FileServer(http.Dir("./static/"))
http.Handle("/", stack.New(middlewareOne, middlewareTwo).ThenHandler(fs))
```

Note that once a chain is 'closed' you can no longer append to it.

When a request executes a chain, control will flow through the chain and *then back up again* in the reverse direction.

### Re-using chains

Chains are designed to be reused. Typically you will setup a chain once, assign it to a variable, then use it across many different routes.

```go
stdStack := stack.New(middlewareOne, middlewareTwo)
extStack := stdStack.Append(middlewareThree, middlewareFour)

http.Handle("/foo", stdStack.Then(fooHandler))
http.Handle("/bar", stdStack.Then(barHandler))
http.Handle("/baz", extStack.Then(bazHandler))
```

As mentioned above, using `Append()` returns a new copy of the chain. In the example above, it means that you can happily alter `extStack` without it mutating `stdStack`.

### Using third-party middleware

There's a good chance that you'll want to make use of existing or third-party middleware in your chain. Any middleware with the signature `func(http.Handler) http.Handler` can be easily adapted for use in your chain with the `stack.Adapt()` function..

For example, if you wanted to use the New York Times' [gzip middleware](https://github.com/NYTimes/gziphandler) you would use:

```go
stack.New(middlewareOne, stack.Adapt(gziphandler.GzipHandler))
```

Or for Goji's [httpauth](https://github.com/goji/httpauth):

```go
authenticate := stack.Adapt(httpauth.SimpleBasicAuth("user", "pass"))
stack.New(middlewareOne, authenticate)
```

For middleware that takes extra parameters, like  [`http.TimeoutHandler`](http://golang.org/pkg/net/http/#TimeoutHandler), the easiest approach is to wrap it in your own constructor. For example:

```go
func timeout(ctx *stack.Context, next http.Handler) http.Handler {
  return http.TimeoutHandler(next, 1*time.Second, "timed out")
}
...
stack.New(timeout, middlewareTwo)
```

### Managing context

You should be aware that `stack.Context` is implemented as a pointer to a `map[string]interface{}`, scoped to the goroutine executing the current HTTP request.

Operations on `stack.Context` are protected by a mutex, so if you need to pass the context pointer to another goroutine (say for logging or completing background process) it is safe for concurrent use.

Data can be added to the context with `Context.Put()`. The first parameter should be a unique string (which acts as the key), and the second parameter the value you want to store. This value can be of any type. Calling `Put()` on an pre-existing key will update its value.

Data can be retrieved from the context with `Context.Get()`. This returns the value stored for the key, or an error if the key does not exist. Data can be also be removed using `Context.Delete()`.

Here's a simple example:

```go
func tokenMiddleware(ctx *stack.Context, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    ctx.Put("token", "c9e452805dee5044ba520198628abcaa")
    next.ServeHTTP(w, r)
  })
}

func tokenHandler(ctx *stack.Context, w http.ResponseWriter, r *http.Request) {
  token, err := ctx.Get("token")
  if err != nil {
    http.Error(w, err.Error(), 500)
    return
  }
  fmt.Fprintf(w, "Token is: %s", token)
}
```

**Using getters and setters**

Because the context is a `map[string]interface{}`, to keep your code type-safe at compile time it's a good idea to restrict yourself to using getter and setter functions which enforce the type when accessing context. This can also be a nice opportunity to return a default value if a key doesn't exist (rather than returning an error). The above example can be re-written as:

```go
func tokenMiddleware(ctx *stack.Context, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    SetToken(ctx, "c9e452805dee5044ba520198628abcab")
    next.ServeHTTP(w, r)
  })
}

func tokenHandler(ctx *stack.Context, w http.ResponseWriter, r *http.Request) {
  token := Token(ctx)
  fmt.Fprintf(w, "Token is: %s", token)
}

func SetToken(ctx *stack.Context, token string) {
    ctx.Put("token") = token
}

func Token(ctx *stack.Context) string {
    token, err := ctx.Get("token")
    if err != nil {
      return ""
    }
    if token, ok := token.(string); ok {
      return token
    }
    return ""
}
```

### Using a base context

It's possible to set up an initial (or *base*) context when creating your chain with the `stack.Init()` function instead of `stack.New()`. 

```go
baseCtx := stack.NewContext()
baseCtx.Put("foo", "bar")

ch := stack.Init(baseCtx).Append(middlewareOne, middlewareTwo)
```

This base context will be available to all requests which are handled by chains derived (only) from the `Init`'d one.

### Changing a base context

Occasionally it might be necessary to amend context *during the course of an HTTP request, but before it starts being handled by your middleware chain*.

An real-life example is needing to inject [httprouter.Params](https://github.com/julienschmidt/httprouter) into the context before it hands over control to a chain. 

This can be achieved with a combination of the `ClosedChain.BaseCtx()` method and the `stack.ReInit()` function.

The `ClosedChain.BaseCtx()` method returns a new pointer to a **copy** of the current base context. You can manipulate this new copy safely without affecting the original chain or any other requests.

You should then use the `stack.ReInit()` function to create a **new copy of the chain with the new base context**.

```go
newCtx := originalChain.BaseCtx()
newCtx.Put("params", ps)

newChain := stack.ReInit(newCtx, originalChain)
```

### TODO 

- Improve README clarity (and add badges)
- Make a `chain.Merge()` method
- Make `*Context.Put()` return itself (so `ctx := NewContext().Put("foo", "bar")` works)
- Mirror master in v1 branch (and mention gopkg.in in README)
- Add benchmarks
- Add usage examples section to documentation

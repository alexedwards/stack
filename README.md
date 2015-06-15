# Stack

Stack makes it simple to create context-aware middleware chains for Go web applications. It is heavily inspired by [Alice](https://github.com/justinas/alice).

[Skip to the example &rsaquo;](#example)

### Why?

- Stack provides a inbuilt request-scoped map for sharing data (or *context*) between HTTP handlers.
- It also lets you create stackable, reusable, handler chains in the Alice style.
- It's compatible with the common `func(http.Handler) http.Handler` middleware pattern.
- Chains satisfy the `http.Handler` interface, so they can be used with the `http.DefaultServeMux`.
- Is compile-time type-safe (so long as you use appropriate [getters and setters](#getting-and-setting-context))
- The package is designed to be simple, non-magic, and get out of your way.

### Installation

```bash
go get github.com/alexedwards/stack
```

## Usage

### Creating a Chain

New middleware chains are created using `stack.New`

```go
stack.New(middlewareOne, middlewareTwo, middlewareThree)
```

You should pass middleware as parameters in the same order that you want them to be executed (reading left to right).

The `stack.New` function accepts middleware that uses the following pattern:

```go
func middlewareOne(ctx stack.Context, h http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     // do something middleware-ish, accessing ctx
     h.ServeHTTP(w, r)
  })
}
```

Middleware with the signature `func(http.Handler) http.Handler` can also be added to a chain, using the `stack.Middleware` adapter. This makes it easy to register third-party middleware with Stack.

For example, if you want to use Goji's [httpauth](https://github.com/goji/httpauth) middleware you would do the following:

```go
authenticate := stack.Middleware(httpauth.SimpleBasicAuth("user", "pass"))

stack.New(authenticate, middlewareOne, middlewareTwo)
```

### Adding Handlers

Application handlers are added to a chain using the `Then()` method. When this method has been called the chain becomes a `http.Handler` and can be registered with Go's `http.DefaultServeMux`. 

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(appHandler))
```

The `Then()` method accepts handlers that use the following pattern:

```go
func appHandler(ctx stack.Context) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     // do something applicaton-ish, accessing ctx
  })
}
```

If this pattern feels to verbose, you can setup your handler as a simple function with the signature `func(stack.Context, http.ResponseWriter, *http.Request)` instead, then adapt it with the `stack.ContextHandlerFunc()` adapter. For example:

```go
func appHandler(ctx stack.Context, w http.ResponseWriter, r *http.Request) {
   // do something applicaton-ish, accessing ctx
}
```

Can be used like:

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(stack.ContextHandlerFunc(appHandler)))
```

In addtion any object which satisfies the `http.Handler` interface and doesn't need access to `stack.Context` can be used in `Then()` &ndash; via the `stack.Handler()` adapter. For example:

```go
fs :=  http.FileServer(http.Dir("./static/"))

http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(stack.Handler(fs)))
```

Similarly any function with the signature `func(http.ResponseWriter, *http.Request)` can be used via the `stack.HandlerFunc()` adapter. For example the function:

```go
func foo(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("foo"))
}
```

Can be used like:

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(stack.HandlerFunc(foo)))
```

### Example

This example chains together a third-party middleware package (`httpAuth`) and a custom `tokenMiddleware` middleware handler. This middleware sets a `token` value in the context which is later accessed by the `tokenHandler` application handler.

```go
package main

import (
    "fmt"
    "github.com/alexedwards/stack"
    "github.com/goji/httpauth"
    "net/http"
)

func main() {
    authenticate := stack.Middleware(httpauth.SimpleBasicAuth("user", "pass"))

    http.Handle("/", stack.New(authenticate, tokenMiddleware).Then(tokenHandler))
    http.ListenAndServe(":3000", nil)
}

func tokenMiddleware(ctx stack.Context, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Add a value to Context with the key 'token'
        ctx["token"] = "c9e452805dee5044ba520198628abcaa"
        next.ServeHTTP(w, r)
    })
}

func tokenHandler(ctx stack.Context) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Retrieve the token from Context and print it
        fmt.Fprintf(w, "Token is: %s", ctx["token"])
    })
}
```

Requesting the resource should give you a response like:

```bash
$ curl -i user:pass@localhost:3000/
HTTP/1.1 200 OK
Content-Length: 41
Content-Type: text/plain; charset=utf-8

Token is: c9e452805dee5044ba520198628abcaa

$ curl -i user:wrongpass@localhost:3000
HTTP/1.1 401 Unauthorized
Content-Length: 13
Content-Type: text/plain; charset=utf-8
Www-Authenticate: Basic realm="Restricted"

Unauthorized
```

### Getting and Setting Context

You should be aware that `stack.Context` is implemented as a `map[string]interface{}`, scoped to the goroutine executing the current HTTP request.

To keep your code type-safe at compile time it's a good idea to restrict yourself to using getter and setter functions when accessing Context. The above example is better written as:

```go
...

func tokenMiddleware(ctx stack.Context, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        SetToken(ctx, "c9e452805dee5044ba520198628abcab")
        next.ServeHTTP(w, r)
    })
}

func tokenHandler(ctx stack.Context) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Token is: %s", Token(ctx))
    })
}

func SetToken(ctx stack.Context, token string) {
    ctx["token"] = token
}

func Token(ctx stack.Context) string {
    if token, ok := ctx["token"].(string); ok {
        return token
    }
    return ""
}
```

You must make sure that your getter and setter signatures both use exactly the same type (in the above example, a `string`). If you mess them up and use different types in the signatures, any type issues may not be picked up until runtime.

As a side note: If you're planning to pass Context to a secondary goroutine for processing you'll need to make sure that it's safe for concurrent use, probably by implementing a [mutex lock](http://www.alexedwards.net/blog/understanding-mutexes) around potentially racy code.

### Reusing Stacks

Like Alice, Stack has the huge benefit of reusable middleware chains:

```go
stdStack := stack.New(middlewareOne, middlewareTwo)

extStack := stdStack.Append(middlewareThree, middlewareFour)

http.Handle("/foo", stdStack.Then(fooHandler))

http.Handle("/bar", extStack.Append(middlewareFive).Then(barHandler))
```

### Ideas

* ~~Add adapter for application handlers with the signature func(ctx, w, r) (done)~~
* Use x/net/context instead of stack.Context.
* Add tests.
* Handle adaptations automatically in New(), Append() and Then().
* Add an AppendStack() function for concatenating stacks.
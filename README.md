# Stack

Stack provides two main features:

1. A convenient interface for chaining middleware handlers and application handlers together, [Alice style](https://github.com/justinas/alice).
2. The ability to pass request-scoped data (or *context*) between middleware and application handlers.

[Skip to the example &rsaquo;](#example)

## Installation

```bash
go get github.com/alexedwards/stack
```

## Usage

### Creating a Chain

New middleware chains are created with `stack.New`, passing in your middleware as parameters. You should specify your middleware in the same order that you want them to be executed.

```go
stack.New(middlewareOne, middlewareTwo, middlewareThree)
```

The `stack.New` function accepts middleware with the signature `func(stack.Context, http.Handler) http.Handler`. In practice that means it accepts middleware with the following pattern:

```go
func middlewareOne(ctx stack.Context, next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     // do something middleware-ish, accessing ctx
     next.ServeHTTP(w, r)
  })
}
```

External middleware with the common signature `func(http.Handler) http.Handler` can also be added to a chain &ndash; you should convert it with the `stack.Middleware` adapter first. For example, if you want to make use of Goji's [httpauth](https://github.com/goji/httpauth) middleware you would do something similar to the following:

```go
authenticate := stack.Middleware(httpauth.SimpleBasicAuth("user", "pass"))

stack.New(authenticate, middlewareOne, middlewareTwo)
```

### Adding Handlers

Application handlers are added to a chain using the `Then` method. When the `Then` method has been called the chain becomes a `http.Handler` and compatible with Go's standard `http.DefaultServeMux`.

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(appHandler))
```

The `Then` method accepts handlers with the signature `func(stack.Context) http.Handler`. In practice that means it accepts handlers with the following pattern:

```go
func appHandler(ctx stack.Context) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
     // do something applicaton-ish, accessing ctx
  })
}
```

Any object which satisfies the `http.Handler` interface can also be used thanks to the `stack.Handler` adapter. For example:

```go
fs :=  stack.Handler(http.FileServer(http.Dir("./static/")))

http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(fs))
```

Similarly the `stack.HandlerFunc` adapter is provided so a function with the signature `func(http.ResponseWriter, *http.Request)` can also be used. For example:

```go
http.Handle("/", stack.New(middlewareOne, middlewareTwo).Then(stack.HandlerFunc(foo)))
```

Where `foo` is a function like the following...

```go
func foo(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("foo"))
}
```

### Example

Here's an example of a middleware chain using third-party `httpAuth` middleware and custom `tokenMiddleware` (which sets a token that is later accessed by our application handler).

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

    http.Handle("/hello", stack.New(authenticate, tokenMiddleware).Then(tokenHandler))
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

```bash
$ curl -i user:pass@localhost:3000/hello
HTTP/1.1 200 OK
Content-Length: 41
Content-Type: text/plain; charset=utf-8

Token is: c9e452805dee5044ba520198628abcaa
```

### Getting and Setting Context

`stack.Context` is implemented as a `map[string]interface{}`. To keep your code type-safe at compile time it's a good idea to use getters and setters when accessing Context.

The above example is better written as:

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

func Token(ctx stack.Context) string {
    return ctx["token"].(string)
}

func SetToken(ctx stack.Context, token string) {
    ctx["token"] = token
}
```

As a side note: If you're going to pass Context as a variable to a function running in a new goroutine, you'll need to make sure that it is safe for concurrent access. One way of doing this is to implement a mutex lock in your getter and setter functions.

### Reusing Stacks

Like Alice, middleware chains can also be reused and appended to. Like so:

```go
stdStack := stack.New(middlewareOne, middlewareTwo)

extStack := stdStack.Append(middlewareThree, middlewareFour)

http.Handle("/foo", stdStack.Then(fooHandler))

http.Handle("/bar", extStack.Append(middlewareFive).Then(barHandler))
```

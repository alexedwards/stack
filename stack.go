package stack

import "net/http"

type Context map[string]interface{}
type ContextHandler func(Context) http.Handler
type ContextMiddleware func(Context, http.Handler) http.Handler

type Chain struct {
	m []ContextMiddleware
	h ContextHandler
}

func New(mw ...ContextMiddleware) Chain {
	return Chain{m: mw}
}

func (c Chain) Append(mw ...ContextMiddleware) Chain {
	c.m = append(c.m, mw...)
	return c
}

func (c Chain) Then(h ContextHandler) chainHandler {
	c.h = h
	return chainHandler(c)
}

type chainHandler Chain

func (ch chainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := make(map[string]interface{})

	final := ch.h(ctx)

	for i := len(ch.m) - 1; i >= 0; i-- {
		final = ch.m[i](ctx, final)
	}

	final.ServeHTTP(w, r)
}

// Adapt http.Handler into a ContextHandler
func Handler(h http.Handler) ContextHandler {
	return func(ctx Context) http.Handler {
		return h
	}
}

// Adapt a function with the signature
// func(http.ResponseWriter, *http.Request) into a ContextHandler
func HandlerFunc(fn func(w http.ResponseWriter, r *http.Request)) ContextHandler {
	return Handler(http.HandlerFunc(fn))
}

// Adapt a function with the signature
// func(Context, http.ResponseWriter, *http.Request) into a ContextHandler
func ContextHandlerFunc(fn func(ctx Context, w http.ResponseWriter, r *http.Request)) ContextHandler {
	return func(ctx Context) http.Handler {
	  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	  	fn(ctx, w, r)
	  })
	}
}

// Adapt third party middleware with the signature
// func(http.Handler) http.Handler into ContextMiddleware
func Middleware(fn func(http.Handler) http.Handler) ContextMiddleware {
	return func(ctx Context, h http.Handler) http.Handler {
		return fn(h)
	}
}

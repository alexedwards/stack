package stack

import "net/http"

type chainHandler func(*Context) http.Handler
type chainMiddleware func(*Context, http.Handler) http.Handler

type Chain struct {
	mws     []chainMiddleware
	h       chainHandler
	baseCtx *Context
}

func Init(ctx *Context) Chain {
	return Chain{baseCtx: ctx}
}

func New(mws ...chainMiddleware) Chain {
	return Init(NewContext()).Append(mws...)
}

func (c Chain) Append(mws ...chainMiddleware) Chain {
	c.mws = append(c.mws, mws...)
	return c
}

func (c Chain) Then(chf func(ctx *Context, w http.ResponseWriter, r *http.Request)) closedChain {
	c.h = adaptContextHandlerFunc(chf)
	return closedChain(c)
}

func (c Chain) ThenHandler(h http.Handler) closedChain {
	c.h = adaptHandler(h)
	return closedChain(c)
}

func (c Chain) ThenHandlerFunc(fn func(http.ResponseWriter, *http.Request)) closedChain {
	c.h = adaptHandlerFunc(fn)
	return closedChain(c)
}

func (c Chain) ThenChainHandler(ch chainHandler) closedChain {
	c.h = ch
	return closedChain(c)
}

type closedChain Chain

func (cc closedChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Always take a copy of baseCtx (i.e. with the pointer at a brand new memory location)
	ctx := cc.baseCtx.copy()

	final := cc.h(ctx)
	for i := len(cc.mws) - 1; i >= 0; i-- {
		final = cc.mws[i](ctx, final)
	}
	final.ServeHTTP(w, r)
}

// Adapt third party middleware with the signature
// func(http.Handler) http.Handler into chainMiddleware
func AdaptMiddleware(fn func(http.Handler) http.Handler) chainMiddleware {
	return func(ctx *Context, h http.Handler) http.Handler {
		return fn(h)
	}
}

// Adapt http.Handler into a chainHandler
func adaptHandler(h http.Handler) chainHandler {
	return func(ctx *Context) http.Handler {
		return h
	}
}

// Adapt a function with the signature
// func(http.ResponseWriter, *http.Request) into a chainHandler
func adaptHandlerFunc(fn func(w http.ResponseWriter, r *http.Request)) chainHandler {
	return adaptHandler(http.HandlerFunc(fn))
}

// Adapt a function with the signature
// func(Context, http.ResponseWriter, *http.Request) into a chainHandler
func adaptContextHandlerFunc(fn func(ctx *Context, w http.ResponseWriter, r *http.Request)) chainHandler {
	return func(ctx *Context) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(ctx, w, r)
		})
	}
}

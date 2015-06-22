package stack

import "net/http"

type chainHandler func(*Context) http.Handler
type chainMiddleware func(*Context, http.Handler) http.Handler

type chain struct {
	mws     []chainMiddleware
	h       chainHandler
	baseCtx *Context
}

func Init(ctx *Context) chain {
	return chain{baseCtx: ctx}
}

func New(mws ...chainMiddleware) chain {
	return Init(NewContext()).Append(mws...)
}

func (c chain) Append(mws ...chainMiddleware) chain {
	newMws := make([]chainMiddleware, len(c.mws)+len(mws))
	copy(newMws[:len(c.mws)], c.mws)
	copy(newMws[len(c.mws):], mws)
	c.mws = newMws
	return c
}

func (c chain) Then(chf func(ctx *Context, w http.ResponseWriter, r *http.Request)) ClosedChain {
	c.h = adaptContextHandlerFunc(chf)
	return ClosedChain(c)
}

func (c chain) ThenHandler(h http.Handler) ClosedChain {
	c.h = adaptHandler(h)
	return ClosedChain(c)
}

func (c chain) ThenHandlerFunc(fn func(http.ResponseWriter, *http.Request)) ClosedChain {
	c.h = adaptHandlerFunc(fn)
	return ClosedChain(c)
}

func (c chain) ThenChainHandler(ch chainHandler) ClosedChain {
	c.h = ch
	return ClosedChain(c)
}

type ClosedChain chain

func (cc ClosedChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Always take a copy of baseCtx (i.e. with the pointer at a brand new memory location)
	ctx := cc.baseCtx.copy()

	final := cc.h(ctx)
	for i := len(cc.mws) - 1; i >= 0; i-- {
		final = cc.mws[i](ctx, final)
	}
	final.ServeHTTP(w, r)
}

func (cc ClosedChain) BaseCtx() *Context {
	return cc.baseCtx.copy()
}

func ReInit(ctx *Context, cc ClosedChain) ClosedChain {
	// Return a new copy of a ClosedChain with baseCtx pointing to a new location
	// Everything else stays the same
	cc.baseCtx = ctx
	return cc
}

// Adapt third party middleware with the signature
// func(http.Handler) http.Handler into chainMiddleware
func Adapt(fn func(http.Handler) http.Handler) chainMiddleware {
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

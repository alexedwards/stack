package stack

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func assertEquals(t *testing.T, e interface{}, o interface{}) {
	if e != o {
		t.Errorf("\n...expected = %v\n...obtained = %v", e, o)
	}
}

func serveAndRequest(h http.Handler) string {
	ts := httptest.NewServer(h)
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(resBody)
}

func bishMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx.Put("bish", "bash")
		fmt.Fprintf(w, "bishMiddleware>")
		next.ServeHTTP(w, r)
	})
}

func flipMiddleware(ctx *Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "flipMiddleware>")
		next.ServeHTTP(w, r)
	})
}

func wobbleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "wobbleMiddleware>")
		next.ServeHTTP(w, r)
	})
}

func bishHandler(ctx *Context, w http.ResponseWriter, r *http.Request) {
	val, _ := ctx.Get("bish")
	fmt.Fprintf(w, "bishHandler [bish=%v]", val)
}

func flipHandler(ctx *Context, w http.ResponseWriter, r *http.Request) {
	valb, _ := ctx.Get("bish")
	valf, _ := ctx.Get("flip")
	fmt.Fprintf(w, "flipHandler [bish=%v,flip=%v]", valb, valf)
}

func bishChainHandler(ctx *Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val, _ := ctx.Get("bish")
		fmt.Fprintf(w, "bishChainHandler [bish=%v]", val)
	})
}

func TestNew(t *testing.T) {
	st := New(bishMiddleware, flipMiddleware).Then(bishHandler)
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestAppend(t *testing.T) {
	st := New(bishMiddleware).Append(flipMiddleware, flipMiddleware).Then(bishHandler)
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>flipMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestAppendDoesNotMutate(t *testing.T) {
	st1 := New(bishMiddleware, flipMiddleware)
	st2 := st1.Append(flipMiddleware, flipMiddleware)
	res := serveAndRequest(st1.Then(bishHandler))
	assertEquals(t, "bishMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
	res = serveAndRequest(st2.Then(bishHandler))
	assertEquals(t, "bishMiddleware>flipMiddleware>flipMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestThen(t *testing.T) {
	chf := func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "An anonymous ContextHandlerFunc")
	}
	st := New().Then(chf)
	res := serveAndRequest(st)
	assertEquals(t, "An anonymous ContextHandlerFunc", res)
}

func TestThenHandler(t *testing.T) {
	st := New().ThenHandler(http.NotFoundHandler())
	res := serveAndRequest(st)
	assertEquals(t, "404 page not found\n", res)
}

func TestThenHandlerFunc(t *testing.T) {
	hf := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "An anonymous HandlerFunc")
	}
	st := New().ThenHandlerFunc(hf)
	res := serveAndRequest(st)
	assertEquals(t, "An anonymous HandlerFunc", res)
}

func TestThenChainHandler(t *testing.T) {
	st := New(bishMiddleware).ThenChainHandler(bishChainHandler)
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>bishChainHandler [bish=bash]", res)
}

func TestMixedMiddleware(t *testing.T) {
	st := New(bishMiddleware, AdaptMiddleware(wobbleMiddleware), flipMiddleware).Then(bishHandler)
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>wobbleMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestInit(t *testing.T) {
	ctx := NewContext()
	ctx.Put("flip", "flop")
	st := Init(ctx).Append(bishMiddleware).Then(flipHandler)
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>flipHandler [bish=bash,flip=flop]", res)
}

func TestBaseCtx(t *testing.T) {
	ctx := NewContext()
	ctx.Put("flip", "flop")
	st := Init(ctx).Append(bishMiddleware).Then(flipHandler)
	bc := st.BaseCtx()
	assertEquals(t, "map[flip:flop]", fmt.Sprintf("%v", bc.m))

	// Test that mutating the returned *Context doesn't mutate the original
	// i.e. it's a new copy
	bc.Put("bish", "bash")
	assertEquals(t, "map[flip:flop]", fmt.Sprintf("%v", st.baseCtx.m))
}

func TestReInit(t *testing.T) {
	ctx := NewContext()
	ctx.Put("flip", "flop")
	st := Init(ctx).Append(flipMiddleware).Then(flipHandler)
	res := serveAndRequest(st)
	assertEquals(t, "flipMiddleware>flipHandler [bish=<nil>,flip=flop]", res)

	newCtx := st.BaseCtx()
	newCtx.Put("bish", "bash")
	st2 := ReInit(newCtx, st)
	res = serveAndRequest(st2)
	assertEquals(t, "flipMiddleware>flipHandler [bish=bash,flip=flop]", res)

	// And the initial one shouldn't be mutated
	res = serveAndRequest(st)
	assertEquals(t, "flipMiddleware>flipHandler [bish=<nil>,flip=flop]", res)
}

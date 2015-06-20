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

func TestNew(t *testing.T) {
	st := New(bishMiddleware, flipMiddleware).Then(ContextHandlerFunc(bishHandler))
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestAppend(t *testing.T) {
	st := New(bishMiddleware).Append(flipMiddleware).Then(ContextHandlerFunc(bishHandler))
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

func TestThen(t *testing.T) {
	st := New().Then(Handler(http.NotFoundHandler()))
	res := serveAndRequest(st)
	assertEquals(t, "404 page not found\n", res)

	hf := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "An anonymous HandlerFunc")
	})
	st = New().Then(HandlerFunc(hf))
	res = serveAndRequest(st)
	assertEquals(t, "An anonymous HandlerFunc", res)

	chf := func(ctx *Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "An anonymous ContextHandlerFunc")
	}
	st = New().Then(ContextHandlerFunc(chf))
	res = serveAndRequest(st)
	assertEquals(t, "An anonymous ContextHandlerFunc", res)
}

func TestMixedMiddleware(t *testing.T) {
	st := New(bishMiddleware, Middleware(wobbleMiddleware), flipMiddleware).Then(ContextHandlerFunc(bishHandler))
	res := serveAndRequest(st)
	assertEquals(t, "bishMiddleware>wobbleMiddleware>flipMiddleware>bishHandler [bish=bash]", res)
}

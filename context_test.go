package stack

import "testing"

func SetUpContextSuite() *Context {
	ctx := NewContext()
	ctx.m["flip"] = "flop"
	return ctx
}

func TestGet(t *testing.T) {
	ctx := SetUpContextSuite()

	val, err := ctx.Get("flip")
	AssertEquals(t, nil, err)
	AssertEquals(t, "flop", val)

	_, err = ctx.Get("wibble")
	AssertEquals(t, "stack.Context: key \"wibble\" does not exist", err.Error())
}

func TestPut(t *testing.T) {
	ctx := SetUpContextSuite()

	ctx.Put("bish", "bash")
	AssertEquals(t, "bash", ctx.m["bish"])
}

func TestDelete(t *testing.T) {
	ctx := SetUpContextSuite()

	ctx.Delete("flip")
	AssertEquals(t, nil, ctx.m["flip"])
}

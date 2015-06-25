package stack

import "testing"

func TestGet(t *testing.T) {
	ctx := NewContext()
	ctx.m["flip"] = "flop"
	ctx.m["bish"] = nil

	val := ctx.Get("flip")
	assertEquals(t, "flop", val)

	val = ctx.Get("bish")
	assertEquals(t, nil, val)
}

func TestPut(t *testing.T) {
	ctx := NewContext()

	ctx.Put("bish", "bash")
	assertEquals(t, "bash", ctx.m["bish"])
}

func TestDelete(t *testing.T) {
	ctx := NewContext()
	ctx.m["flip"] = "flop"

	ctx.Delete("flip")
	assertEquals(t, nil, ctx.m["flip"])
}

func TestCopy(t *testing.T) {
	ctx := NewContext()
	ctx.m["flip"] = "flop"

	ctx2 := ctx.copy()
	ctx2.m["bish"] = "bash"
	assertEquals(t, nil, ctx.m["bish"])
	assertEquals(t, "bash", ctx2.m["bish"])
}

func TestExists(t *testing.T) {
	ctx := NewContext()
	ctx.m["flip"] = "flop"

	assertEquals(t, true, ctx.Exists("flip"))
	assertEquals(t, false, ctx.Exists("bash"))
}

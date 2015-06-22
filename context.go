package stack

import (
	"fmt"
	"sync"
)

type Context struct {
	mu sync.RWMutex
	m  map[string]interface{}
}

func NewContext() *Context {
	m := make(map[string]interface{})
	return &Context{m: m}
}

func (c *Context) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok := c.m[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("stack.Context: key %q does not exist", key)
}

func (c *Context) Put(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = val
}

func (c *Context) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

func (c *Context) copy() *Context {
	nc := NewContext()
	c.mu.RLock()
	c.mu.RUnlock()
	for k, v := range c.m {
		nc.m[k] = v
	}
	return nc
}

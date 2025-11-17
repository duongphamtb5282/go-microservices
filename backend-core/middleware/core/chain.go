package core

import (
	"context"
	"sort"
	"sync"
)

// Chain represents a middleware chain
type Chain struct {
	middlewares []Middleware
	mutex       sync.RWMutex
}

// NewChain creates a new middleware chain
func NewChain() *Chain {
	return &Chain{
		middlewares: make([]Middleware, 0),
	}
}

// Add adds a middleware to the chain
func (c *Chain) Add(middleware Middleware) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.middlewares = append(c.middlewares, middleware)
	c.sort()
}

// Remove removes a middleware from the chain
func (c *Chain) Remove(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, middleware := range c.middlewares {
		if middleware.Name() == name {
			c.middlewares = append(c.middlewares[:i], c.middlewares[i+1:]...)
			break
		}
	}
}

// Execute executes the middleware chain
func (c *Chain) Execute(ctx context.Context, req *Request) (*Response, error) {
	c.mutex.RLock()
	middlewares := make([]Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)
	c.mutex.RUnlock()

	if len(middlewares) == 0 {
		return nil, ErrChainEmpty
	}

	return c.executeChain(ctx, req, middlewares, 0)
}

// executeChain executes the middleware chain recursively
func (c *Chain) executeChain(ctx context.Context, req *Request, middlewares []Middleware, index int) (*Response, error) {
	if index >= len(middlewares) {
		return nil, ErrNoHandler
	}

	middleware := middlewares[index]
	next := func(ctx context.Context, req *Request) (*Response, error) {
		return c.executeChain(ctx, req, middlewares, index+1)
	}

	return middleware.Execute(ctx, req, next)
}

// List returns all middlewares in the chain
func (c *Chain) List() []Middleware {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	middlewares := make([]Middleware, len(c.middlewares))
	copy(middlewares, c.middlewares)
	return middlewares
}

// Clear clears all middlewares from the chain
func (c *Chain) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.middlewares = make([]Middleware, 0)
}

// Size returns the number of middlewares in the chain
func (c *Chain) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.middlewares)
}

// Has checks if a middleware exists in the chain
func (c *Chain) Has(name string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, middleware := range c.middlewares {
		if middleware.Name() == name {
			return true
		}
	}
	return false
}

// Get gets a middleware by name
func (c *Chain) Get(name string) (Middleware, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for _, middleware := range c.middlewares {
		if middleware.Name() == name {
			return middleware, nil
		}
	}
	return nil, ErrMiddlewareNotFound
}

// sort sorts middlewares by priority
func (c *Chain) sort() {
	sort.Slice(c.middlewares, func(i, j int) bool {
		return c.middlewares[i].Priority() < c.middlewares[j].Priority()
	})
}

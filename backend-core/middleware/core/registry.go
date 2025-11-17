package core

import (
	"context"
	"sync"
)

// MiddlewareChain represents a middleware chain
type MiddlewareChain interface {
	// Add adds a middleware to the chain
	Add(middleware Middleware)

	// Remove removes a middleware from the chain
	Remove(name string)

	// Execute executes the middleware chain
	Execute(ctx context.Context, req *Request) (*Response, error)

	// List returns all middlewares in the chain
	List() []Middleware
}

// MiddlewareRegistry represents a middleware registry
type MiddlewareRegistry interface {
	// Register registers a middleware
	Register(middleware Middleware) error

	// Get gets a middleware by name
	Get(name string) (Middleware, error)

	// List lists all registered middlewares
	List() []Middleware

	// Unregister unregisters a middleware
	Unregister(name string) error
}

// Registry is a concrete implementation of MiddlewareRegistry
type Registry struct {
	middlewares map[string]Middleware
	mu          sync.RWMutex
}

// NewRegistry creates a new middleware registry
func NewRegistry() *Registry {
	return &Registry{
		middlewares: make(map[string]Middleware),
	}
}

// Register registers a middleware
func (r *Registry) Register(middleware Middleware) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := middleware.Name()
	if _, exists := r.middlewares[name]; exists {
		return ErrMiddlewareExists
	}

	r.middlewares[name] = middleware
	return nil
}

// Get gets a middleware by name
func (r *Registry) Get(name string) (Middleware, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	middleware, exists := r.middlewares[name]
	if !exists {
		return nil, ErrMiddlewareNotFound
	}

	return middleware, nil
}

// List lists all registered middlewares
func (r *Registry) List() []Middleware {
	r.mu.RLock()
	defer r.mu.RUnlock()

	middlewares := make([]Middleware, 0, len(r.middlewares))
	for _, middleware := range r.middlewares {
		middlewares = append(middlewares, middleware)
	}

	return middlewares
}

// Unregister unregisters a middleware
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.middlewares[name]; !exists {
		return ErrMiddlewareNotFound
	}

	delete(r.middlewares, name)
	return nil
}

package core

import (
	"context"
)

// Middleware represents a middleware function
type Middleware interface {
	// Name returns the middleware name
	Name() string

	// Priority returns the middleware priority (lower = higher priority)
	Priority() int

	// Execute executes the middleware
	Execute(ctx context.Context, req *Request, next NextHandler) (*Response, error)
}

// NextHandler represents the next handler in the chain
type NextHandler func(ctx context.Context, req *Request) (*Response, error)

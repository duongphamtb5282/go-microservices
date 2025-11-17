package core

import "errors"

// Error types
var (
	ErrNoHandler          = errors.New("no handler found")
	ErrMiddlewareExists   = errors.New("middleware already exists")
	ErrMiddlewareNotFound = errors.New("middleware not found")
	ErrInvalidPriority    = errors.New("invalid priority")
	ErrChainEmpty         = errors.New("middleware chain is empty")
)

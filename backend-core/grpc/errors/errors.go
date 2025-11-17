package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Common error types
var (
	// ErrUnauthenticated represents an authentication error
	ErrUnauthenticated = errors.New("unauthenticated")
	// ErrUnauthorized represents an authorization error
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInvalidArgument represents an invalid argument error
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrNotFound represents a not found error
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists represents an already exists error
	ErrAlreadyExists = errors.New("already exists")
	// ErrInternal represents an internal error
	ErrInternal = errors.New("internal error")
	// ErrRateLimitExceeded represents a rate limit exceeded error
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// MapError maps an error to appropriate gRPC status code
func MapError(err error) error {
	if err == nil {
		return nil
	}

	// If already a gRPC status error, return as is
	if _, ok := status.FromError(err); ok {
		return err
	}

	// Map common errors to gRPC codes
	switch {
	case errors.Is(err, ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, ErrUnauthorized):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, ErrRateLimitExceeded):
		return status.Error(codes.ResourceExhausted, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// NewUnauthenticatedError creates a new unauthenticated error
func NewUnauthenticatedError(message string) error {
	return status.Error(codes.Unauthenticated, message)
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) error {
	return status.Error(codes.PermissionDenied, message)
}

// NewInvalidArgumentError creates a new invalid argument error
func NewInvalidArgumentError(message string) error {
	return status.Error(codes.InvalidArgument, message)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) error {
	return status.Error(codes.NotFound, message)
}

// NewInternalError creates a new internal error
func NewInternalError(message string) error {
	return status.Error(codes.Internal, message)
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(message string) error {
	return status.Error(codes.ResourceExhausted, message)
}

package recovery

import (
	"context"
	"fmt"
	"runtime/debug"

	"backend-core/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryHandlerFunc is a function that recovers from a panic
type RecoveryHandlerFunc func(ctx context.Context, p interface{}) error

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery
func UnaryServerInterceptor(logger *logging.Logger, handler RecoveryHandlerFunc) grpc.UnaryServerInterceptor {
	if handler == nil {
		handler = DefaultRecoveryHandler(logger)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handler(ctx, r)
			}
		}()

		return h(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor for panic recovery
func StreamServerInterceptor(logger *logging.Logger, handler RecoveryHandlerFunc) grpc.StreamServerInterceptor {
	if handler == nil {
		handler = DefaultRecoveryHandler(logger)
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, h grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = handler(ss.Context(), r)
			}
		}()

		return h(srv, ss)
	}
}

// DefaultRecoveryHandler is the default recovery handler
func DefaultRecoveryHandler(logger *logging.Logger) RecoveryHandlerFunc {
	return func(ctx context.Context, p interface{}) error {
		// Log the panic with stack trace
		stackTrace := string(debug.Stack())

		logger.Error("gRPC panic recovered",
			logging.Any("panic", p),
			logging.String("stack_trace", stackTrace))

		// Return internal error
		return status.Errorf(codes.Internal, "internal server error: %v", p)
	}
}

// CustomRecoveryHandler allows custom recovery logic
func CustomRecoveryHandler(logger *logging.Logger, customHandler func(interface{}) error) RecoveryHandlerFunc {
	return func(ctx context.Context, p interface{}) error {
		// Log the panic
		stackTrace := string(debug.Stack())

		logger.Error("gRPC panic recovered",
			logging.Any("panic", p),
			logging.String("stack_trace", stackTrace))

		// Use custom handler if provided
		if customHandler != nil {
			return customHandler(p)
		}

		// Default behavior
		return status.Errorf(codes.Internal, "internal server error: %v", p)
	}
}

// RecoveryHandlerFuncWithContext creates a recovery handler with access to context
func RecoveryHandlerFuncWithContext(logger *logging.Logger, fn func(ctx context.Context, p interface{}) error) RecoveryHandlerFunc {
	return func(ctx context.Context, p interface{}) error {
		// Log the panic
		stackTrace := string(debug.Stack())

		logger.Error("gRPC panic recovered",
			logging.Any("panic", p),
			logging.String("stack_trace", stackTrace))

		if fn != nil {
			return fn(ctx, p)
		}

		return status.Errorf(codes.Internal, "internal server error: %v", p)
	}
}

// NewRecoveryInterceptor creates a new recovery interceptor with options
func NewRecoveryInterceptor(logger *logging.Logger, opts ...RecoveryOption) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	config := &recoveryConfig{
		handler: DefaultRecoveryHandler(logger),
	}

	for _, opt := range opts {
		opt(config)
	}

	return UnaryServerInterceptor(logger, config.handler),
		StreamServerInterceptor(logger, config.handler)
}

type recoveryConfig struct {
	handler RecoveryHandlerFunc
}

// RecoveryOption is an option for recovery interceptor
type RecoveryOption func(*recoveryConfig)

// WithRecoveryHandler sets a custom recovery handler
func WithRecoveryHandler(handler RecoveryHandlerFunc) RecoveryOption {
	return func(c *recoveryConfig) {
		c.handler = handler
	}
}

// WithRecoveryHandlerFunc sets a simple recovery handler function
func WithRecoveryHandlerFunc(fn func(interface{}) error) RecoveryOption {
	return func(c *recoveryConfig) {
		c.handler = func(ctx context.Context, p interface{}) error {
			return fn(p)
		}
	}
}

// panicError wraps a panic value as an error
type panicError struct {
	value interface{}
}

func (p *panicError) Error() string {
	return fmt.Sprintf("panic: %v", p.value)
}

// WrapPanic wraps a panic value as an error
func WrapPanic(p interface{}) error {
	return &panicError{value: p}
}

package retry

import (
	"context"
	"math"
	"math/rand"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including initial request)
	MaxAttempts int

	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration

	// BackoffMultiplier is the backoff multiplier (typically 2.0)
	BackoffMultiplier float64

	// Jitter adds randomization to backoff to prevent thundering herd
	// Value between 0.0 (no jitter) and 1.0 (full jitter)
	Jitter float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            0.2, // 20% jitter
	}
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, don't retry
		return false
	}

	return IsRetryableCode(st.Code())
}

// IsRetryableCode determines if a gRPC status code is retryable
func IsRetryableCode(code codes.Code) bool {
	switch code {
	// Definitely retryable - transient errors
	case codes.Unavailable, // Service unavailable
		codes.DeadlineExceeded,  // Request timeout
		codes.ResourceExhausted, // Rate limited or overloaded
		codes.Aborted:           // Transaction conflict
		return true

	// Maybe retryable (network/server issues)
	case codes.Unknown, // Unknown error (network issues)
		codes.Internal: // Internal server error
		return true

	// Never retryable - client errors
	case codes.InvalidArgument, // Bad request
		codes.NotFound,           // Resource not found
		codes.AlreadyExists,      // Duplicate
		codes.PermissionDenied,   // Authorization failed
		codes.Unauthenticated,    // Authentication failed
		codes.FailedPrecondition, // Precondition not met
		codes.OutOfRange,         // Value out of range
		codes.Unimplemented,      // Method not implemented
		codes.DataLoss:           // Data corruption
		return false

	// Success or cancellation
	case codes.OK,
		codes.Canceled: // Client cancelled
		return false

	default:
		return false
	}
}

// CalculateBackoff calculates the next backoff duration with jitter
func CalculateBackoff(attempt int, config *RetryConfig) time.Duration {
	// Calculate exponential backoff
	backoff := float64(config.InitialBackoff) *
		math.Pow(config.BackoffMultiplier, float64(attempt))

	// Cap at max backoff
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	// Add jitter
	if config.Jitter > 0 {
		jitterRange := backoff * config.Jitter
		jitter := rand.Float64()*jitterRange*2 - jitterRange // +/- jitter
		backoff += jitter

		// Ensure non-negative
		if backoff < 0 {
			backoff = 0
		}
	}

	return time.Duration(backoff)
}

// UnaryClientInterceptor returns a new unary client interceptor with enhanced retry logic
func UnaryClientInterceptor(config *RetryConfig) grpc.UnaryClientInterceptor {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error

		for attempt := 0; attempt < config.MaxAttempts; attempt++ {
			// Try the request
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			lastErr = err

			// Don't retry on last attempt
			if attempt == config.MaxAttempts-1 {
				break
			}

			// Check if error is retryable
			if !IsRetryable(err) {
				return err
			}

			// Calculate backoff with jitter
			backoff := CalculateBackoff(attempt, config)

			// Wait before retry
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		return lastErr
	}
}

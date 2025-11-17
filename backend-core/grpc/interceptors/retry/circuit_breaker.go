package retry

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CircuitBreakerState represents circuit breaker states
type CircuitBreakerState int

const (
	StateClosed   CircuitBreakerState = iota // Normal operation
	StateOpen                                // Failing, reject requests
	StateHalfOpen                            // Testing if recovered
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           CircuitBreakerState
	failureCount    int
	successCount    int
	lastFailureTime time.Time

	// Configuration
	maxFailures     int           // Failures before opening
	resetTimeout    time.Duration // Time before trying again
	halfOpenSuccess int           // Successes needed to close
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, halfOpenSuccess int) *CircuitBreaker {
	return &CircuitBreaker{
		state:           StateClosed,
		maxFailures:     maxFailures,
		resetTimeout:    resetTimeout,
		halfOpenSuccess: halfOpenSuccess,
	}
}

// CanAttempt checks if a request can be attempted
func (cb *CircuitBreaker) CanAttempt() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess records a successful attempt
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenSuccess {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}
}

// RecordFailure records a failed attempt
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.failureCount = 0
		cb.successCount = 0
	} else if cb.state == StateClosed {
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// UnaryClientInterceptorWithCircuitBreaker returns a retry interceptor with circuit breaker
func UnaryClientInterceptorWithCircuitBreaker(config *RetryConfig, cb *CircuitBreaker) grpc.UnaryClientInterceptor {
	if config == nil {
		config = DefaultRetryConfig()
	}

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Check circuit breaker before attempting
		if !cb.CanAttempt() {
			return status.Errorf(codes.Unavailable, "circuit breaker is open for %s", method)
		}

		var lastErr error

		for attempt := 0; attempt < config.MaxAttempts; attempt++ {
			err := invoker(ctx, method, req, reply, cc, opts...)

			if err == nil {
				cb.RecordSuccess()
				return nil
			}

			lastErr = err
			cb.RecordFailure()

			// Don't retry on last attempt or if circuit is open
			if attempt == config.MaxAttempts-1 || !cb.CanAttempt() {
				break
			}

			// Check if retryable
			if !IsRetryable(err) {
				return err
			}

			// Calculate backoff
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

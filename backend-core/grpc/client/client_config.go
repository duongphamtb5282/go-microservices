package client

import "time"

// ClientConfig holds gRPC client configuration
type ClientConfig struct {
	// Address is the server address
	Address string
	// Timeout is the request timeout
	Timeout time.Duration
	// MaxMessageSize is the maximum message size in bytes
	MaxMessageSize int
	// Insecure determines if TLS should be disabled
	Insecure bool
	// KeepAliveTime is the keepalive time
	KeepAliveTime time.Duration
	// KeepAliveTimeout is the keepalive timeout
	KeepAliveTimeout time.Duration
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int
	// BackoffMultiplier is the backoff multiplier
	BackoffMultiplier float64
	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Address:          "localhost:50051",
		Timeout:          30 * time.Second,
		MaxMessageSize:   10 * 1024 * 1024, // 10MB
		Insecure:         true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:       3,
		BackoffMultiplier: 2.0,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
	}
}

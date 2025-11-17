package client

import (
	"context"
	"time"

	"backend-core/grpc/interceptors/auth"
	grpclogging "backend-core/grpc/interceptors/logging"
	"backend-core/grpc/interceptors/metrics"
	"backend-core/grpc/interceptors/retry"
	"backend-core/grpc/interceptors/tracing"
	"backend-core/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientBuilder builds a gRPC client with interceptors
type ClientBuilder struct {
	config             *ClientConfig
	retryConfig        *RetryConfig
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	dialOptions        []grpc.DialOption
}

// NewClientBuilder creates a new client builder
func NewClientBuilder(config *ClientConfig) *ClientBuilder {
	if config == nil {
		config = DefaultClientConfig()
	}

	return &ClientBuilder{
		config:             config,
		unaryInterceptors:  []grpc.UnaryClientInterceptor{},
		streamInterceptors: []grpc.StreamClientInterceptor{},
		dialOptions:        []grpc.DialOption{},
	}
}

// WithTimeout sets the request timeout
func (b *ClientBuilder) WithTimeout(timeout time.Duration) *ClientBuilder {
	b.config.Timeout = timeout
	return b
}

// WithRetry adds retry interceptor with enhanced logic
func (b *ClientBuilder) WithRetry(config *RetryConfig) *ClientBuilder {
	if config == nil {
		config = DefaultRetryConfig()
	}
	b.retryConfig = config

	// Use enhanced retry interceptor from retry package
	retryConfig := &retry.RetryConfig{
		MaxAttempts:       config.MaxAttempts,
		InitialBackoff:    config.InitialBackoff,
		MaxBackoff:        config.MaxBackoff,
		BackoffMultiplier: config.BackoffMultiplier,
		Jitter:            0.2, // 20% jitter by default
	}

	b.unaryInterceptors = append(b.unaryInterceptors, retry.UnaryClientInterceptor(retryConfig))
	return b
}

// WithLogging adds logging interceptor
func (b *ClientBuilder) WithLogging(logger *logging.Logger) *ClientBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		grpclogging.UnaryClientInterceptor(logger))
	return b
}

// WithAuth adds authentication interceptor
func (b *ClientBuilder) WithAuth(token string) *ClientBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		auth.UnaryClientInterceptor(token))
	b.streamInterceptors = append(b.streamInterceptors,
		auth.StreamClientInterceptor(token))
	return b
}

// WithMetrics adds metrics interceptor
func (b *ClientBuilder) WithMetrics(reporter metrics.MetricsReporter) *ClientBuilder {
	if reporter != nil {
		b.unaryInterceptors = append(b.unaryInterceptors,
			metrics.UnaryClientInterceptor(reporter))
	}
	return b
}

// WithTracing adds tracing interceptor
func (b *ClientBuilder) WithTracing() *ClientBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		tracing.UnaryClientInterceptor())
	b.streamInterceptors = append(b.streamInterceptors,
		tracing.StreamClientInterceptor())
	return b
}

// WithUnaryInterceptor adds a custom unary interceptor
func (b *ClientBuilder) WithUnaryInterceptor(interceptor grpc.UnaryClientInterceptor) *ClientBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors, interceptor)
	return b
}

// WithStreamInterceptor adds a custom stream interceptor
func (b *ClientBuilder) WithStreamInterceptor(interceptor grpc.StreamClientInterceptor) *ClientBuilder {
	b.streamInterceptors = append(b.streamInterceptors, interceptor)
	return b
}

// WithDialOption adds a custom dial option
func (b *ClientBuilder) WithDialOption(opt grpc.DialOption) *ClientBuilder {
	b.dialOptions = append(b.dialOptions, opt)
	return b
}

// Build creates the gRPC client connection
func (b *ClientBuilder) Build() (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}

	// Add interceptors
	if len(b.unaryInterceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(b.unaryInterceptors...))
	}
	if len(b.streamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(b.streamInterceptors...))
	}

	// Add credentials
	if b.config.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add message size limits
	if b.config.MaxMessageSize > 0 {
		opts = append(opts,
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(b.config.MaxMessageSize),
				grpc.MaxCallSendMsgSize(b.config.MaxMessageSize),
			))
	}

	// Add keepalive parameters
	if b.config.KeepAliveTime > 0 {
		kaParams := keepalive.ClientParameters{
			Time:                b.config.KeepAliveTime,
			Timeout:             b.config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}
		opts = append(opts, grpc.WithKeepaliveParams(kaParams))
	}

	// Add custom dial options
	opts = append(opts, b.dialOptions...)

	// Dial with timeout
	ctx, cancel := context.WithTimeout(context.Background(), b.config.Timeout)
	defer cancel()

	return grpc.DialContext(ctx, b.config.Address, opts...)
}

// BuildWithDefaults creates a gRPC client connection with default interceptors
func BuildWithDefaults(address string, logger *logging.Logger) (*grpc.ClientConn, error) {
	config := &ClientConfig{
		Address:          address,
		Timeout:          30 * time.Second,
		MaxMessageSize:   10 * 1024 * 1024,
		Insecure:         true,
		KeepAliveTime:    30 * time.Second,
		KeepAliveTimeout: 10 * time.Second,
	}

	return NewClientBuilder(config).
		WithRetry(DefaultRetryConfig()).
		WithLogging(logger).
		WithTracing().
		Build()
}

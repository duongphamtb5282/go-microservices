package server

import (
	"backend-core/cache"
	"backend-core/grpc/interceptors/auth"
	grpclogging "backend-core/grpc/interceptors/logging"
	"backend-core/grpc/interceptors/metrics"
	"backend-core/grpc/interceptors/ratelimit"
	"backend-core/grpc/interceptors/recovery"
	"backend-core/grpc/interceptors/tracing"
	"backend-core/grpc/interceptors/validation"
	"backend-core/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// ServerBuilder builds a gRPC server with interceptors
type ServerBuilder struct {
	config             *ServerConfig
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	serverOptions      []grpc.ServerOption
}

// NewServerBuilder creates a new server builder
func NewServerBuilder(config *ServerConfig) *ServerBuilder {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &ServerBuilder{
		config:             config,
		unaryInterceptors:  []grpc.UnaryServerInterceptor{},
		streamInterceptors: []grpc.StreamServerInterceptor{},
		serverOptions:      []grpc.ServerOption{},
	}
}

// WithRecovery adds panic recovery interceptor
func (b *ServerBuilder) WithRecovery(logger *logging.Logger) *ServerBuilder {
	unary, stream := recovery.NewRecoveryInterceptor(logger)
	b.unaryInterceptors = append(b.unaryInterceptors, unary)
	b.streamInterceptors = append(b.streamInterceptors, stream)
	return b
}

// WithLogging adds logging interceptor
func (b *ServerBuilder) WithLogging(logger *logging.Logger, config *grpclogging.LoggingConfig) *ServerBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		grpclogging.UnaryServerInterceptor(logger, config))
	b.streamInterceptors = append(b.streamInterceptors,
		grpclogging.StreamServerInterceptor(logger, config))
	return b
}

// WithAuth adds authentication interceptor
func (b *ServerBuilder) WithAuth(logger *logging.Logger, config *auth.AuthConfig) *ServerBuilder {
	if config != nil && config.Enabled {
		b.unaryInterceptors = append(b.unaryInterceptors,
			auth.UnaryServerInterceptor(logger, config))
		b.streamInterceptors = append(b.streamInterceptors,
			auth.StreamServerInterceptor(logger, config))
	}
	return b
}

// WithValidation adds validation interceptor
func (b *ServerBuilder) WithValidation() *ServerBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		validation.UnaryServerInterceptor())
	b.streamInterceptors = append(b.streamInterceptors,
		validation.StreamServerInterceptor())
	return b
}

// WithRateLimit adds rate limiting interceptor
func (b *ServerBuilder) WithRateLimit(cache cache.Cache, config *ratelimit.RateLimitConfig, logger *logging.Logger) *ServerBuilder {
	if config != nil && config.Enabled {
		limiter := ratelimit.NewRedisLimiter(cache, config, logger)
		b.unaryInterceptors = append(b.unaryInterceptors,
			ratelimit.UnaryServerInterceptor(limiter, config))
		b.streamInterceptors = append(b.streamInterceptors,
			ratelimit.StreamServerInterceptor(limiter, config))
	}
	return b
}

// WithMetrics adds metrics interceptor
func (b *ServerBuilder) WithMetrics(reporter metrics.MetricsReporter) *ServerBuilder {
	if reporter != nil {
		b.unaryInterceptors = append(b.unaryInterceptors,
			metrics.UnaryServerInterceptor(reporter))
		b.streamInterceptors = append(b.streamInterceptors,
			metrics.StreamServerInterceptor(reporter))
	}
	return b
}

// WithTracing adds tracing interceptor
func (b *ServerBuilder) WithTracing() *ServerBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors,
		tracing.UnaryServerInterceptor())
	b.streamInterceptors = append(b.streamInterceptors,
		tracing.StreamServerInterceptor())
	return b
}

// WithUnaryInterceptor adds a custom unary interceptor
func (b *ServerBuilder) WithUnaryInterceptor(interceptor grpc.UnaryServerInterceptor) *ServerBuilder {
	b.unaryInterceptors = append(b.unaryInterceptors, interceptor)
	return b
}

// WithStreamInterceptor adds a custom stream interceptor
func (b *ServerBuilder) WithStreamInterceptor(interceptor grpc.StreamServerInterceptor) *ServerBuilder {
	b.streamInterceptors = append(b.streamInterceptors, interceptor)
	return b
}

// WithServerOption adds a custom server option
func (b *ServerBuilder) WithServerOption(opt grpc.ServerOption) *ServerBuilder {
	b.serverOptions = append(b.serverOptions, opt)
	return b
}

// Build creates the gRPC server with all configured interceptors
func (b *ServerBuilder) Build() *grpc.Server {
	opts := []grpc.ServerOption{}

	// Add interceptors
	if len(b.unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(b.unaryInterceptors...))
	}
	if len(b.streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(b.streamInterceptors...))
	}

	// Add message size limits
	if b.config.MaxMessageSize > 0 {
		opts = append(opts,
			grpc.MaxRecvMsgSize(b.config.MaxMessageSize),
			grpc.MaxSendMsgSize(b.config.MaxMessageSize))
	}

	// Add connection timeout
	if b.config.ConnectionTimeout > 0 {
		opts = append(opts, grpc.ConnectionTimeout(b.config.ConnectionTimeout))
	}

	// Add keepalive parameters
	if b.config.KeepaliveTime > 0 {
		kaParams := keepalive.ServerParameters{
			Time:    b.config.KeepaliveTime,
			Timeout: b.config.KeepaliveTimeout,
		}
		opts = append(opts, grpc.KeepaliveParams(kaParams))
	}

	// Add max concurrent streams
	if b.config.MaxConcurrentStreams > 0 {
		opts = append(opts, grpc.MaxConcurrentStreams(b.config.MaxConcurrentStreams))
	}

	// Add custom server options
	opts = append(opts, b.serverOptions...)

	return grpc.NewServer(opts...)
}

// BuildWithDefaults creates a gRPC server with default interceptors
func BuildWithDefaults(config *ServerConfig, logger *logging.Logger, authConfig *auth.AuthConfig) *grpc.Server {
	builder := NewServerBuilder(config).
		WithRecovery(logger).
		WithLogging(logger, &grpclogging.LoggingConfig{
			LogPayload:        false,
			LogPayloadOnError: true,
		}).
		WithTracing().
		WithValidation()

	if authConfig != nil && authConfig.Enabled {
		builder = builder.WithAuth(logger, authConfig)
	}

	return builder.Build()
}

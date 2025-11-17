package ratelimit

import (
	"context"
	"fmt"
	"time"

	"backend-core/cache"
	grpcerrors "backend-core/grpc/errors"
	grpcmiddleware "backend-core/grpc/middleware"
	"backend-core/logging"

	"google.golang.org/grpc"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Enabled determines if rate limiting is enabled
	Enabled bool
	// RequestsPerMinute is the default rate limit
	RequestsPerMinute int
	// Burst is the burst size
	Burst int
	// MethodLimits are per-method rate limits
	MethodLimits map[string]int
	// KeyPrefix is the Redis key prefix
	KeyPrefix string
}

// Limiter interface for rate limiting
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// RedisLimiter implements rate limiting using Redis
type RedisLimiter struct {
	cache  cache.Cache
	config *RateLimitConfig
	logger *logging.Logger
}

// NewRedisLimiter creates a new Redis-based rate limiter
func NewRedisLimiter(cache cache.Cache, config *RateLimitConfig, logger *logging.Logger) *RedisLimiter {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "ratelimit:"
	}

	return &RedisLimiter{
		cache:  cache,
		config: config,
		logger: logger,
	}
}

// Allow checks if a request is allowed
func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := l.config.KeyPrefix + key

	// Get current count
	var countStr string
	err := l.cache.Get(ctx, fullKey, &countStr)
	if err != nil && err != cache.ErrCacheMiss {
		l.logger.Error("Failed to get rate limit count", "error", err)
		// On error, allow the request (fail open)
		return true, nil
	}

	count := 0
	if countStr != "" {
		fmt.Sscanf(countStr, "%d", &count)
	}

	// Check if limit exceeded
	limit := l.config.RequestsPerMinute
	if count >= limit {
		return false, nil
	}

	// Increment counter
	count++
	ttl := time.Minute

	err = l.cache.Set(ctx, fullKey, fmt.Sprintf("%d", count), ttl)
	if err != nil {
		l.logger.Error("Failed to set rate limit count", "error", err)
		// On error, allow the request (fail open)
		return true, nil
	}

	return true, nil
}

// UnaryServerInterceptor returns a new unary server interceptor for rate limiting
func UnaryServerInterceptor(limiter Limiter, config *RateLimitConfig) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !config.Enabled {
			return handler(ctx, req)
		}

		// Determine rate limit key (by user ID or IP)
		key := getRateLimitKey(ctx, info.FullMethod)

		// Check rate limit
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			// On error, allow the request (fail open)
			return handler(ctx, req)
		}

		if !allowed {
			return nil, grpcerrors.NewRateLimitError("rate limit exceeded")
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor for rate limiting
func StreamServerInterceptor(limiter Limiter, config *RateLimitConfig) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !config.Enabled {
			return handler(srv, ss)
		}

		ctx := ss.Context()

		// Determine rate limit key
		key := getRateLimitKey(ctx, info.FullMethod)

		// Check rate limit
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			// On error, allow the request (fail open)
			return handler(srv, ss)
		}

		if !allowed {
			return grpcerrors.NewRateLimitError("rate limit exceeded")
		}

		return handler(srv, ss)
	}
}

// getRateLimitKey generates a rate limit key based on user or IP
func getRateLimitKey(ctx context.Context, method string) string {
	// Try to get user ID
	if userID, ok := grpcmiddleware.GetUserID(ctx); ok {
		return fmt.Sprintf("user:%s:%s", userID, method)
	}

	// Fall back to IP address
	if clientIP, ok := grpcmiddleware.GetClientIP(ctx); ok {
		return fmt.Sprintf("ip:%s:%s", clientIP, method)
	}

	// Fall back to peer address
	peerAddr := grpcmiddleware.ExtractPeerInfo(ctx)
	return fmt.Sprintf("peer:%s:%s", peerAddr, method)
}

// MemoryLimiter implements a simple in-memory rate limiter (for testing)
type MemoryLimiter struct {
	limits map[string]*limiterState
	config *RateLimitConfig
}

type limiterState struct {
	count             int
	lastReset         time.Time
	requestsPerMinute int
}

// NewMemoryLimiter creates a new in-memory rate limiter
func NewMemoryLimiter(config *RateLimitConfig) *MemoryLimiter {
	return &MemoryLimiter{
		limits: make(map[string]*limiterState),
		config: config,
	}
}

// Allow checks if a request is allowed
func (l *MemoryLimiter) Allow(ctx context.Context, key string) (bool, error) {
	state, ok := l.limits[key]
	if !ok {
		state = &limiterState{
			count:             0,
			lastReset:         time.Now(),
			requestsPerMinute: l.config.RequestsPerMinute,
		}
		l.limits[key] = state
	}

	// Reset if window expired
	if time.Since(state.lastReset) > time.Minute {
		state.count = 0
		state.lastReset = time.Now()
	}

	// Check limit
	if state.count >= state.requestsPerMinute {
		return false, nil
	}

	state.count++
	return true, nil
}

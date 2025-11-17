package operations

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// RedisCounters handles Redis counter operations
type RedisCounters struct {
	client redis.UniversalClient
}

// NewRedisCounters creates a new RedisCounters instance
func NewRedisCounters(client redis.UniversalClient) *RedisCounters {
	return &RedisCounters{client: client}
}

// Increment increments a counter
func (r *RedisCounters) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrementBy increments a counter by a specific amount
func (r *RedisCounters) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Decrement decrements a counter
func (r *RedisCounters) Decrement(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// DecrementBy decrements a counter by a specific amount
func (r *RedisCounters) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.DecrBy(ctx, key, value).Result()
}

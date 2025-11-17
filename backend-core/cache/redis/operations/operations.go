package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisOperations handles basic Redis key-value operations
type RedisOperations struct {
	client redis.UniversalClient
}

// NewRedisOperations creates a new RedisOperations instance
func NewRedisOperations(client redis.UniversalClient) *RedisOperations {
	return &RedisOperations{client: client}
}

// Set stores a value in the cache
func (r *RedisOperations) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.Set(ctx, key, jsonValue, expiration).Err()
}

// Get retrieves a value from the cache
func (r *RedisOperations) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a value from the cache
func (r *RedisOperations) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (r *RedisOperations) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

// Exists checks if a key exists in the cache
func (r *RedisOperations) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire sets the expiration time for a key
func (r *RedisOperations) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key
func (r *RedisOperations) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Ping tests the Redis connection
func (r *RedisOperations) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// ErrCacheMiss is returned when a key is not found in the cache
var ErrCacheMiss = fmt.Errorf("cache miss")

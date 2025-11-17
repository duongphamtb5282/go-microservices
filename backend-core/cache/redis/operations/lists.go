package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisLists handles Redis list operations
type RedisLists struct {
	client redis.UniversalClient
}

// NewRedisLists creates a new RedisLists instance
func NewRedisLists(client redis.UniversalClient) *RedisLists {
	return &RedisLists{client: client}
}

// ListPush adds a value to the end of a list
func (r *RedisLists) ListPush(ctx context.Context, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.RPush(ctx, key, jsonValue).Err()
}

// ListPop removes and returns the last element of a list
func (r *RedisLists) ListPop(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// ListLength returns the length of a list
func (r *RedisLists) ListLength(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

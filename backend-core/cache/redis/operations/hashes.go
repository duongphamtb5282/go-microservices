package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisHashes handles Redis hash operations
type RedisHashes struct {
	client redis.UniversalClient
}

// NewRedisHashes creates a new RedisHashes instance
func NewRedisHashes(client redis.UniversalClient) *RedisHashes {
	return &RedisHashes{client: client}
}

// HashSet sets a field in a hash
func (r *RedisHashes) HashSet(ctx context.Context, key, field string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.HSet(ctx, key, field, jsonValue).Err()
}

// HashGet gets a field from a hash
func (r *RedisHashes) HashGet(ctx context.Context, key, field string, dest interface{}) error {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// HashGetAll gets all fields from a hash
func (r *RedisHashes) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

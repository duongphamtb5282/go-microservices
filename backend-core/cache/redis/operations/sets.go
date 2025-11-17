package operations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisSets handles Redis set operations
type RedisSets struct {
	client redis.UniversalClient
}

// NewRedisSets creates a new RedisSets instance
func NewRedisSets(client redis.UniversalClient) *RedisSets {
	return &RedisSets{client: client}
}

// SetAdd adds a member to a set
func (r *RedisSets) SetAdd(ctx context.Context, key string, member interface{}) error {
	jsonValue, err := json.Marshal(member)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.SAdd(ctx, key, jsonValue).Err()
}

// SetMembers returns all members of a set
func (r *RedisSets) SetMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// SetIsMember checks if a member exists in a set
func (r *RedisSets) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	jsonValue, err := json.Marshal(member)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.client.SIsMember(ctx, key, jsonValue).Result()
}

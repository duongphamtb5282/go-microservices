package reload

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCacheInvalidator implements cache invalidation for Redis
type RedisCacheInvalidator struct {
	client   redis.UniversalClient
	reloader CacheReloader
}

// NewRedisCacheInvalidator creates a new Redis cache invalidator
func NewRedisCacheInvalidator(client redis.UniversalClient, reloader CacheReloader) *RedisCacheInvalidator {
	return &RedisCacheInvalidator{
		client:   client,
		reloader: reloader,
	}
}

// Invalidate invalidates a single cache entry
func (i *RedisCacheInvalidator) Invalidate(ctx context.Context, key string) error {
	// Delete the key from cache
	if err := i.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate key %s: %w", key, err)
	}

	return nil
}

// InvalidatePattern invalidates cache entries matching a pattern
func (i *RedisCacheInvalidator) InvalidatePattern(ctx context.Context, pattern string) error {
	// Get all keys matching the pattern
	keys, err := i.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys matching pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all matching keys
	if err := i.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to invalidate keys matching pattern %s: %w", pattern, err)
	}

	return nil
}

// InvalidateAll invalidates all cache entries
func (i *RedisCacheInvalidator) InvalidateAll(ctx context.Context) error {
	// Flush the entire database
	if err := i.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to invalidate all cache: %w", err)
	}

	return nil
}

// InvalidateAndReload invalidates and reloads a cache entry
func (i *RedisCacheInvalidator) InvalidateAndReload(ctx context.Context, key string) error {
	// Invalidate the key
	if err := i.Invalidate(ctx, key); err != nil {
		return fmt.Errorf("failed to invalidate key %s: %w", key, err)
	}

	// Reload the key if reloader is available
	if i.reloader != nil {
		if err := i.reloader.Reload(ctx, key); err != nil {
			return fmt.Errorf("failed to reload key %s: %w", key, err)
		}
	}

	return nil
}

// InvalidatePatternAndReload invalidates and reloads cache entries matching a pattern
func (i *RedisCacheInvalidator) InvalidatePatternAndReload(ctx context.Context, pattern string) error {
	// Get all keys matching the pattern
	keys, err := i.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys matching pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Invalidate all matching keys
	if err := i.InvalidatePattern(ctx, pattern); err != nil {
		return fmt.Errorf("failed to invalidate keys matching pattern %s: %w", pattern, err)
	}

	// Reload all keys if reloader is available
	if i.reloader != nil {
		if err := i.reloader.ReloadBatch(ctx, keys); err != nil {
			return fmt.Errorf("failed to reload keys matching pattern %s: %w", pattern, err)
		}
	}

	return nil
}

// InvalidateWithTTL invalidates a key and sets a new TTL
func (i *RedisCacheInvalidator) InvalidateWithTTL(ctx context.Context, key string, ttl time.Duration) error {
	// Set the key with a very short TTL instead of deleting
	if err := i.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set TTL for key %s: %w", key, err)
	}

	return nil
}

// InvalidateAndSet invalidates a key and sets a new value
func (i *RedisCacheInvalidator) InvalidateAndSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Set the new value directly
	if err := i.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set new value for key %s: %w", key, err)
	}

	return nil
}

// GetInvalidationStats returns statistics about cache invalidation
func (i *RedisCacheInvalidator) GetInvalidationStats(ctx context.Context) (map[string]interface{}, error) {
	// Get database info
	info, err := i.client.Info(ctx, "keyspace").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database info: %w", err)
	}

	// Get total number of keys
	keys, err := i.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	stats := map[string]interface{}{
		"total_keys":    keys,
		"database_info": info,
		"timestamp":     time.Now(),
	}

	return stats, nil
}

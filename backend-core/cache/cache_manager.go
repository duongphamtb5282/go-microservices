package cache

import (
	"context"
	"time"
)

// CacheManager manages cache operations
type CacheManager struct {
	cache Cache
}

// NewCacheManager creates a new cache manager
func NewCacheManager(cache Cache) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

// GetCache returns the underlying cache instance
func (cm *CacheManager) GetCache() Cache {
	return cm.cache
}

// SetCache sets the cache instance
func (cm *CacheManager) SetCache(cache Cache) {
	cm.cache = cache
}

// GetOrSet gets a value from cache or sets it if not found
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, setter func() (interface{}, error), expiration time.Duration) error {
	err := cm.cache.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != ErrCacheMiss {
		return err
	}

	value, err := setter()
	if err != nil {
		return err
	}

	if err := cm.cache.Set(ctx, key, value, expiration); err != nil {
		return err
	}

	// Copy the value to dest
	return cm.cache.Get(ctx, key, dest)
}

// Remember caches the result of a function call
func (cm *CacheManager) Remember(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error), expiration time.Duration) error {
	return cm.GetOrSet(ctx, key, dest, fn, expiration)
}

// Forget removes a key from cache
func (cm *CacheManager) Forget(ctx context.Context, key string) error {
	return cm.cache.Delete(ctx, key)
}

// Flush removes all keys from cache
func (cm *CacheManager) Flush(ctx context.Context) error {
	return cm.cache.DeletePattern(ctx, "*")
}

// RememberForever caches a value forever (until manually removed)
func (cm *CacheManager) RememberForever(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error)) error {
	return cm.GetOrSet(ctx, key, dest, fn, 0)
}

// Increment increments a counter
func (cm *CacheManager) Increment(ctx context.Context, key string) (int64, error) {
	return cm.cache.Increment(ctx, key)
}

// Decrement decrements a counter
func (cm *CacheManager) Decrement(ctx context.Context, key string) (int64, error) {
	return cm.cache.Decrement(ctx, key)
}

// Has checks if a key exists in cache
func (cm *CacheManager) Has(ctx context.Context, key string) (bool, error) {
	return cm.cache.Exists(ctx, key)
}

// Put stores a value in cache
func (cm *CacheManager) Put(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cm.cache.Set(ctx, key, value, expiration)
}

// Get retrieves a value from cache
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	return cm.cache.Get(ctx, key, dest)
}

// Pull retrieves and removes a value from cache
func (cm *CacheManager) Pull(ctx context.Context, key string, dest interface{}) error {
	err := cm.cache.Get(ctx, key, dest)
	if err != nil {
		return err
	}

	return cm.cache.Delete(ctx, key)
}

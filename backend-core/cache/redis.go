package cache

import (
	"context"
	"fmt"
	"time"

	"backend-core/cache/redis/client"
	"backend-core/cache/redis/operations"
	"backend-core/cache/redis/reload"
	"backend-core/cache/redis/transactions"
	"backend-core/cache/redis/transformers"
	"backend-core/config"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements the Cache interface using Redis
// This is a facade that composes specialized Redis handlers
type RedisCache struct {
	client       redis.UniversalClient
	config       *config.RedisConfig
	ops          *operations.RedisOperations
	counters     *operations.RedisCounters
	lists        *operations.RedisLists
	sets         *operations.RedisSets
	hashes       *operations.RedisHashes
	transactions *transactions.RedisTransactions
	reloader     reload.CacheReloader
	invalidator  reload.CacheInvalidator
	warmer       reload.CacheWarmer
	// Custom reload logic components
	hookManager        *reload.ReloadHookManager
	strategyManager    *reload.CustomReloadStrategyManager
	transformerManager *transformers.DataTransformerManager
}

// NewRedisCache creates a new Redis cache instance from config
func NewRedisCache(cfg *config.RedisConfig) *RedisCache {
	factory := client.NewRedisClientFactory()
	redisClient := factory.CreateClient(cfg)

	return &RedisCache{
		client:             redisClient,
		config:             cfg,
		ops:                operations.NewRedisOperations(redisClient),
		counters:           operations.NewRedisCounters(redisClient),
		lists:              operations.NewRedisLists(redisClient),
		sets:               operations.NewRedisSets(redisClient),
		hashes:             operations.NewRedisHashes(redisClient),
		transactions:       transactions.NewRedisTransactions(redisClient),
		hookManager:        reload.NewReloadHookManager(),
		strategyManager:    reload.NewCustomReloadStrategyManager(),
		transformerManager: transformers.NewDataTransformerManager(),
	}
}

// NewStandaloneRedisCache creates a simple standalone Redis cache instance
func NewStandaloneRedisCache(addr, password string, db int) *RedisCache {
	cfg := &config.RedisConfig{
		Name:       "standalone-redis",
		Addr:       addr,
		Password:   password,
		DB:         db,
		UseCluster: false,
		PoolSize:   10,
	}
	return NewRedisCache(cfg)
}

// ============================================================================
// Basic Operations (delegated to RedisOperations)
// ============================================================================

// Set stores a value in the cache
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.ops.Set(ctx, key, value, expiration)
}

// Get retrieves a value from the cache
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	return r.ops.Get(ctx, key, dest)
}

// Delete removes a value from the cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.ops.Delete(ctx, key)
}

// DeletePattern removes all keys matching a pattern
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	return r.ops.DeletePattern(ctx, pattern)
}

// Exists checks if a key exists in the cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	return r.ops.Exists(ctx, key)
}

// Expire sets the expiration time for a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.ops.Expire(ctx, key, expiration)
}

// TTL returns the time to live for a key
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.ops.TTL(ctx, key)
}

// Ping tests the Redis connection
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.ops.Ping(ctx)
}

// ============================================================================
// Counter Operations (delegated to RedisCounters)
// ============================================================================

// Increment increments a counter
func (r *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	return r.counters.Increment(ctx, key)
}

// IncrementBy increments a counter by a specific amount
func (r *RedisCache) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.counters.IncrementBy(ctx, key, value)
}

// Decrement decrements a counter
func (r *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	return r.counters.Decrement(ctx, key)
}

// DecrementBy decrements a counter by a specific amount
func (r *RedisCache) DecrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.counters.DecrementBy(ctx, key, value)
}

// ============================================================================
// List Operations (delegated to RedisLists)
// ============================================================================

// ListPush adds a value to the end of a list
func (r *RedisCache) ListPush(ctx context.Context, key string, value interface{}) error {
	return r.lists.ListPush(ctx, key, value)
}

// ListPop removes and returns the last element of a list
func (r *RedisCache) ListPop(ctx context.Context, key string, dest interface{}) error {
	return r.lists.ListPop(ctx, key, dest)
}

// ListLength returns the length of a list
func (r *RedisCache) ListLength(ctx context.Context, key string) (int64, error) {
	return r.lists.ListLength(ctx, key)
}

// ============================================================================
// Set Operations (delegated to RedisSets)
// ============================================================================

// SetAdd adds a member to a set
func (r *RedisCache) SetAdd(ctx context.Context, key string, member interface{}) error {
	return r.sets.SetAdd(ctx, key, member)
}

// SetMembers returns all members of a set
func (r *RedisCache) SetMembers(ctx context.Context, key string) ([]string, error) {
	return r.sets.SetMembers(ctx, key)
}

// SetIsMember checks if a member exists in a set
func (r *RedisCache) SetIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return r.sets.SetIsMember(ctx, key, member)
}

// ============================================================================
// Hash Operations (delegated to RedisHashes)
// ============================================================================

// HashSet sets a field in a hash
func (r *RedisCache) HashSet(ctx context.Context, key, field string, value interface{}) error {
	return r.hashes.HashSet(ctx, key, field, value)
}

// HashGet gets a field from a hash
func (r *RedisCache) HashGet(ctx context.Context, key, field string, dest interface{}) error {
	return r.hashes.HashGet(ctx, key, field, dest)
}

// HashGetAll gets all fields from a hash
func (r *RedisCache) HashGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.hashes.HashGetAll(ctx, key)
}

// ============================================================================
// Connection Management
// ============================================================================

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetConfig returns the Redis configuration
func (r *RedisCache) GetConfig() *config.RedisConfig {
	return r.config
}

// IsCluster returns true if running in cluster mode
func (r *RedisCache) IsCluster() bool {
	return r.config != nil && r.config.UseCluster
}

// GetClient returns the underlying Redis client
func (r *RedisCache) GetClient() redis.UniversalClient {
	return r.client
}

// ============================================================================
// Transaction Operations (delegated to RedisTransactions)
// ============================================================================

// BeginTransaction starts a new Redis transaction
func (r *RedisCache) BeginTransaction(ctx context.Context) *transactions.RedisTransaction {
	return r.transactions.Begin(ctx)
}

// Watch watches one or more keys for modifications during a transaction
func (r *RedisCache) Watch(ctx context.Context, keys ...string) error {
	return r.transactions.Watch(ctx, keys...)
}

// WithTransaction executes a function within a transaction
func (r *RedisCache) WithTransaction(ctx context.Context, fn func(*transactions.RedisTransaction) error) error {
	tx := r.BeginTransaction(ctx)

	// Execute the function with the transaction
	if err := fn(tx); err != nil {
		// If function fails, discard the transaction
		tx.Discard()
		return err
	}

	// Execute the transaction
	_, err := tx.Exec()
	return err
}

// WithWatchTransaction executes a function within a watched transaction
func (r *RedisCache) WithWatchTransaction(ctx context.Context, keys []string, fn func(*transactions.RedisTransaction) error) error {
	// Watch the keys
	if err := r.Watch(ctx, keys...); err != nil {
		return err
	}

	// Execute the transaction
	return r.WithTransaction(ctx, fn)
}

// ============================================================================
// Cache Reloading Methods
// ============================================================================

// SetReloader sets the cache reloader
func (r *RedisCache) SetReloader(reloader reload.CacheReloader) {
	r.reloader = reloader
}

// SetInvalidator sets the cache invalidator
func (r *RedisCache) SetInvalidator(invalidator reload.CacheInvalidator) {
	r.invalidator = invalidator
}

// SetWarmer sets the cache warmer
func (r *RedisCache) SetWarmer(warmer reload.CacheWarmer) {
	r.warmer = warmer
}

// GetReloader returns the cache reloader
func (r *RedisCache) GetReloader() reload.CacheReloader {
	return r.reloader
}

// GetInvalidator returns the cache invalidator
func (r *RedisCache) GetInvalidator() reload.CacheInvalidator {
	return r.invalidator
}

// GetWarmer returns the cache warmer
func (r *RedisCache) GetWarmer() reload.CacheWarmer {
	return r.warmer
}

// Reload reloads a single cache entry
func (r *RedisCache) Reload(ctx context.Context, key string) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}
	return r.reloader.Reload(ctx, key)
}

// ReloadBatch reloads multiple cache entries
func (r *RedisCache) ReloadBatch(ctx context.Context, keys []string) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}
	return r.reloader.ReloadBatch(ctx, keys)
}

// ReloadAll reloads the entire cache
func (r *RedisCache) ReloadAll(ctx context.Context) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}
	return r.reloader.ReloadAll(ctx)
}

// Invalidate invalidates a cache entry
func (r *RedisCache) Invalidate(ctx context.Context, key string) error {
	if r.invalidator == nil {
		return fmt.Errorf("no invalidator configured")
	}
	return r.invalidator.Invalidate(ctx, key)
}

// InvalidatePattern invalidates cache entries matching a pattern
func (r *RedisCache) InvalidatePattern(ctx context.Context, pattern string) error {
	if r.invalidator == nil {
		return fmt.Errorf("no invalidator configured")
	}
	return r.invalidator.InvalidatePattern(ctx, pattern)
}

// InvalidateAll invalidates all cache entries
func (r *RedisCache) InvalidateAll(ctx context.Context) error {
	if r.invalidator == nil {
		return fmt.Errorf("no invalidator configured")
	}
	return r.invalidator.InvalidateAll(ctx)
}

// InvalidateAndReload invalidates and reloads a cache entry
func (r *RedisCache) InvalidateAndReload(ctx context.Context, key string) error {
	if r.invalidator == nil {
		return fmt.Errorf("no invalidator configured")
	}
	return r.invalidator.InvalidateAndReload(ctx, key)
}

// WarmUp warms up the cache
func (r *RedisCache) WarmUp(ctx context.Context) error {
	if r.warmer == nil {
		return fmt.Errorf("no warmer configured")
	}
	return r.warmer.WarmUp(ctx)
}

// WarmUpKeys warms up specific cache keys
func (r *RedisCache) WarmUpKeys(ctx context.Context, keys []string) error {
	if r.warmer == nil {
		return fmt.Errorf("no warmer configured")
	}
	return r.warmer.WarmUpKeys(ctx, keys)
}

// IsWarmedUp checks if cache is warmed up
func (r *RedisCache) IsWarmedUp(ctx context.Context) (bool, error) {
	if r.warmer == nil {
		return false, fmt.Errorf("no warmer configured")
	}
	return r.warmer.IsWarmedUp(ctx)
}

// ============================================================================
// Custom Reload Logic Methods
// ============================================================================

// GetHookManager returns the reload hook manager
func (r *RedisCache) GetHookManager() *reload.ReloadHookManager {
	return r.hookManager
}

// GetStrategyManager returns the custom reload strategy manager
func (r *RedisCache) GetStrategyManager() *reload.CustomReloadStrategyManager {
	return r.strategyManager
}

// GetTransformerManager returns the data transformer manager
func (r *RedisCache) GetTransformerManager() *transformers.DataTransformerManager {
	return r.transformerManager
}

// AddReloadHook adds a reload hook
func (r *RedisCache) AddReloadHook(hook reload.ReloadHook) {
	r.hookManager.AddHook(hook)
}

// RemoveReloadHook removes a reload hook by name
func (r *RedisCache) RemoveReloadHook(name string) {
	r.hookManager.RemoveHook(name)
}

// RegisterCustomStrategy registers a custom reload strategy
func (r *RedisCache) RegisterCustomStrategy(strategy reload.CustomReloadStrategy) {
	r.strategyManager.RegisterStrategy(strategy)
}

// GetCustomStrategy returns a custom strategy by name
func (r *RedisCache) GetCustomStrategy(name string) (reload.CustomReloadStrategy, bool) {
	return r.strategyManager.GetStrategy(name)
}

// SetDefaultCustomStrategy sets the default custom strategy
func (r *RedisCache) SetDefaultCustomStrategy(strategy reload.CustomReloadStrategy) {
	r.strategyManager.SetDefaultStrategy(strategy)
}

// AddDataTransformer adds a data transformer
func (r *RedisCache) AddDataTransformer(transformer transformers.DataTransformer) {
	r.transformerManager.AddTransformer(transformer)
}

// RemoveDataTransformer removes a data transformer by name
func (r *RedisCache) RemoveDataTransformer(name string) {
	r.transformerManager.RemoveTransformer(name)
}

// ReloadWithCustomLogic reloads data using custom logic
func (r *RedisCache) ReloadWithCustomLogic(ctx context.Context, key string, strategyName string) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}

	// Get custom strategy
	strategy, exists := r.strategyManager.GetStrategy(strategyName)
	if !exists {
		return fmt.Errorf("custom strategy %s not found", strategyName)
	}

	// Check if should reload
	lastReload := time.Now().Add(-time.Hour) // Default to 1 hour ago
	if strategy.ShouldReload(ctx, key, lastReload) {
		// Execute pre-reload hooks
		if err := r.hookManager.ExecuteBeforeReload(ctx, key, nil); err != nil {
			return fmt.Errorf("pre-reload hooks failed: %w", err)
		}

		// Execute strategy pre-reload
		if err := strategy.PreReload(ctx, key, nil); err != nil {
			return fmt.Errorf("strategy pre-reload failed: %w", err)
		}

		// Perform reload
		if err := r.reloader.Reload(ctx, key); err != nil {
			// Execute error hooks
			r.hookManager.ExecuteOnReloadError(ctx, key, err)
			strategy.OnReloadError(ctx, key, err)
			return fmt.Errorf("reload failed: %w", err)
		}

		// Execute strategy post-reload
		if err := strategy.PostReload(ctx, key, nil); err != nil {
			return fmt.Errorf("strategy post-reload failed: %w", err)
		}

		// Execute post-reload hooks
		if err := r.hookManager.ExecuteAfterReload(ctx, key, nil); err != nil {
			return fmt.Errorf("post-reload hooks failed: %w", err)
		}
	}

	return nil
}

// ReloadBatchWithCustomLogic reloads multiple keys using custom logic
func (r *RedisCache) ReloadBatchWithCustomLogic(ctx context.Context, keys []string, strategyName string) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}

	// Get custom strategy
	strategy, exists := r.strategyManager.GetStrategy(strategyName)
	if !exists {
		return fmt.Errorf("custom strategy %s not found", strategyName)
	}

	// Filter keys that should be reloaded
	var keysToReload []string
	lastReload := time.Now().Add(-time.Hour) // Default to 1 hour ago

	for _, key := range keys {
		if strategy.ShouldReload(ctx, key, lastReload) {
			keysToReload = append(keysToReload, key)
		}
	}

	if len(keysToReload) == 0 {
		return nil // No keys need reloading
	}

	// Sort keys by priority
	// This is a simplified implementation - in practice, you'd want more sophisticated sorting
	sortedKeys := make([]string, len(keysToReload))
	copy(sortedKeys, keysToReload)

	// Execute pre-reload hooks for all keys
	for _, key := range sortedKeys {
		if err := r.hookManager.ExecuteBeforeReload(ctx, key, nil); err != nil {
			return fmt.Errorf("pre-reload hooks failed for key %s: %w", key, err)
		}
		if err := strategy.PreReload(ctx, key, nil); err != nil {
			return fmt.Errorf("strategy pre-reload failed for key %s: %w", key, err)
		}
	}

	// Perform batch reload
	if err := r.reloader.ReloadBatch(ctx, sortedKeys); err != nil {
		// Execute error hooks for all keys
		for _, key := range sortedKeys {
			r.hookManager.ExecuteOnReloadError(ctx, key, err)
			strategy.OnReloadError(ctx, key, err)
		}
		return fmt.Errorf("batch reload failed: %w", err)
	}

	// Execute post-reload hooks for all keys
	for _, key := range sortedKeys {
		if err := strategy.PostReload(ctx, key, nil); err != nil {
			return fmt.Errorf("strategy post-reload failed for key %s: %w", key, err)
		}
		if err := r.hookManager.ExecuteAfterReload(ctx, key, nil); err != nil {
			return fmt.Errorf("post-reload hooks failed for key %s: %w", key, err)
		}
	}

	return nil
}

// Clear clears all data from the cache
func (r *RedisCache) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// TransformData transforms data using registered transformers
func (r *RedisCache) TransformData(ctx context.Context, key string, data interface{}) (interface{}, error) {
	return r.transformerManager.TransformData(ctx, key, data)
}

// ReloadWithTransformation reloads data and applies transformations
func (r *RedisCache) ReloadWithTransformation(ctx context.Context, key string) error {
	if r.reloader == nil {
		return fmt.Errorf("no reloader configured")
	}

	// Get data source
	source := r.reloader.GetDataSource()
	if source == nil {
		return fmt.Errorf("no data source configured")
	}

	// Load data from source
	data, err := source.LoadData(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	// Transform data
	transformedData, err := r.transformerManager.TransformData(ctx, key, data)
	if err != nil {
		return fmt.Errorf("failed to transform data: %w", err)
	}

	// Store transformed data in cache
	if err := r.client.Set(ctx, key, transformedData, time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store transformed data: %w", err)
	}

	return nil
}

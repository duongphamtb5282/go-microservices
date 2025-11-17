package decorators

import (
	"context"
	"fmt"
	"time"

	"backend-core/cache"
	"backend-core/cache/strategies"
	"backend-core/config"
	"backend-core/logging"
)

// CacheDecorator provides a decorator pattern for caching operations
type CacheDecorator struct {
	cache           cache.Cache
	strategyManager *strategies.StrategyManager
	logger          *logging.Logger
	config          *CacheDecoratorConfig
}

// CacheDecoratorConfig holds configuration for the cache decorator
type CacheDecoratorConfig struct {
	DefaultTTL       time.Duration `mapstructure:"default_ttl" json:"default_ttl" yaml:"default_ttl"`
	EntityTTL        time.Duration `mapstructure:"entity_ttl" json:"entity_ttl" yaml:"entity_ttl"`
	ListTTL          time.Duration `mapstructure:"list_ttl" json:"list_ttl" yaml:"list_ttl"`
	EnableStrategies bool          `mapstructure:"enable_strategies" json:"enable_strategies" yaml:"enable_strategies"`
	StrategyType     string        `mapstructure:"strategy_type" json:"strategy_type" yaml:"strategy_type"`
	EnableMetrics    bool          `mapstructure:"enable_metrics" json:"enable_metrics" yaml:"enable_metrics"`
	KeyPrefix        string        `mapstructure:"key_prefix" json:"key_prefix" yaml:"key_prefix"`
}

// DefaultCacheDecoratorConfig returns default configuration
func DefaultCacheDecoratorConfig() *CacheDecoratorConfig {
	return &CacheDecoratorConfig{
		DefaultTTL:       15 * time.Minute,
		EntityTTL:        30 * time.Minute,
		ListTTL:          5 * time.Minute,
		EnableStrategies: true,
		StrategyType:     "write_through",
		EnableMetrics:    true,
		KeyPrefix:        "app",
	}
}

// NewCacheDecorator creates a new cache decorator
func NewCacheDecorator(cfg config.RedisConfig, logger *logging.Logger, decoratorConfig *CacheDecoratorConfig) (*CacheDecorator, error) {
	// Create Redis cache instance
	redisCache := cache.NewRedisCache(&cfg)

	// Set default config if not provided
	if decoratorConfig == nil {
		decoratorConfig = DefaultCacheDecoratorConfig()
	}

	// Create strategy manager
	strategyManager := strategies.NewStrategyManager()

	// Add strategies if enabled
	if decoratorConfig.EnableStrategies {
		// Add read-through strategy
		readThroughStrategy := strategies.NewReadThroughStrategy(
			"read-through",
			redisCache,
			&GenericDataSource{cache: redisCache, logger: logger},
		)
		strategyManager.RegisterStrategy(readThroughStrategy)

		// Add write-through strategy
		writeThroughStrategy := strategies.NewWriteThroughStrategy(
			"write-through",
			redisCache,
			&GenericDataSource{cache: redisCache, logger: logger},
		)
		strategyManager.RegisterStrategy(writeThroughStrategy)

		// Add cache-aside strategy
		cacheAsideStrategy := strategies.NewCacheAsideStrategy(
			"cache-aside",
			redisCache,
			&GenericDataSource{cache: redisCache, logger: logger},
		)
		strategyManager.RegisterStrategy(cacheAsideStrategy)

		// Add write-behind strategy
		writeBehindStrategy := strategies.NewWriteBehindStrategy(
			"write-behind",
			redisCache,
			&GenericDataSource{cache: redisCache, logger: logger},
		)
		strategyManager.RegisterStrategy(writeBehindStrategy)
	}

	return &CacheDecorator{
		cache:           redisCache,
		strategyManager: strategyManager,
		logger:          logger,
		config:          decoratorConfig,
	}, nil
}

// GenericDataSource implements the DataSource interface for generic operations
type GenericDataSource struct {
	cache  cache.Cache
	logger *logging.Logger
}

// LoadData loads data from the source (database simulation)
func (ds *GenericDataSource) LoadData(ctx context.Context, key string, dest interface{}) error {
	// In a real implementation, this would load from database
	// For now, we'll simulate by checking if data exists in cache
	ds.logger.Info("Loading data from source", logging.String("key", key))

	// Simulate database load
	time.Sleep(10 * time.Millisecond)

	// Return cache miss to simulate database load
	return cache.ErrCacheMiss
}

// StoreData stores data in the source (database simulation)
func (ds *GenericDataSource) StoreData(ctx context.Context, key string, value interface{}) error {
	// In a real implementation, this would store in database
	ds.logger.Info("Storing data in source", logging.String("key", key))

	// Simulate database store
	time.Sleep(5 * time.Millisecond)

	return nil
}

// DeleteData deletes data from the source (database simulation)
func (ds *GenericDataSource) DeleteData(ctx context.Context, key string) error {
	// In a real implementation, this would delete from database
	ds.logger.Info("Deleting data from source", logging.String("key", key))

	// Simulate database delete
	time.Sleep(5 * time.Millisecond)

	return nil
}

// Generic Operations with Caching

// Get retrieves data by key with caching
func (cd *CacheDecorator) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := cd.buildKey(key)

	cd.logger.Info("Getting data from cache", logging.String("key", fullKey))

	// Try to get from cache first
	err := cd.cache.Get(ctx, fullKey, dest)
	if err == nil {
		cd.logger.Info("Data found in cache", logging.String("key", fullKey))
		return nil
	}

	if err != cache.ErrCacheMiss {
		cd.logger.Error("Cache error", logging.String("key", fullKey), logging.Error(err))
		return fmt.Errorf("cache error: %w", err)
	}

	cd.logger.Info("Data not found in cache", logging.String("key", fullKey))
	return cache.ErrCacheMiss
}

// Set stores data in cache
func (cd *CacheDecorator) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := cd.buildKey(key)

	cd.logger.Info("Setting data in cache", logging.String("key", fullKey))

	// Use write-through strategy if enabled
	if cd.config.EnableStrategies {
		strategy, exists := cd.strategyManager.GetStrategy(cd.config.StrategyType)
		if exists {
			return strategy.Write(ctx, fullKey, value, ttl)
		}
	}

	// Fallback to direct cache set
	return cd.cache.Set(ctx, fullKey, value, ttl)
}

// SetEntity stores an entity in cache with default entity TTL
func (cd *CacheDecorator) SetEntity(ctx context.Context, key string, value interface{}) error {
	return cd.Set(ctx, key, value, cd.config.EntityTTL)
}

// SetList stores a list in cache with default list TTL
func (cd *CacheDecorator) SetList(ctx context.Context, key string, value interface{}) error {
	return cd.Set(ctx, key, value, cd.config.ListTTL)
}

// Delete removes data from cache
func (cd *CacheDecorator) Delete(ctx context.Context, key string) error {
	fullKey := cd.buildKey(key)

	cd.logger.Info("Deleting data from cache", logging.String("key", fullKey))

	// Use write-through strategy if enabled
	if cd.config.EnableStrategies {
		strategy, exists := cd.strategyManager.GetStrategy(cd.config.StrategyType)
		if exists {
			return strategy.Delete(ctx, fullKey)
		}
	}

	// Fallback to direct cache delete
	return cd.cache.Delete(ctx, fullKey)
}

// DeletePattern removes all keys matching a pattern
func (cd *CacheDecorator) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := cd.buildKey(pattern)

	cd.logger.Info("Deleting data from cache by pattern", logging.String("pattern", fullPattern))

	// Use direct cache delete pattern (strategies don't support pattern deletion)
	return cd.cache.DeletePattern(ctx, fullPattern)
}

// Exists checks if a key exists in cache
func (cd *CacheDecorator) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := cd.buildKey(key)

	cd.logger.Info("Checking if data exists in cache", logging.String("key", fullKey))

	return cd.cache.Exists(ctx, fullKey)
}

// SetWithTTL stores data with custom TTL
func (cd *CacheDecorator) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return cd.Set(ctx, key, value, ttl)
}

// GetOrSet gets a value from cache or sets it if not found
func (cd *CacheDecorator) GetOrSet(ctx context.Context, key string, dest interface{}, setter func() (interface{}, error), ttl time.Duration) error {
	// Try to get from cache first
	err := cd.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != cache.ErrCacheMiss {
		return err
	}

	// Call setter to get the value
	value, err := setter()
	if err != nil {
		return err
	}

	// Set the value in cache
	if err := cd.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Copy the value to dest
	return cd.Get(ctx, key, dest)
}

// Remember caches the result of a function call
func (cd *CacheDecorator) Remember(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error), ttl time.Duration) error {
	return cd.GetOrSet(ctx, key, dest, fn, ttl)
}

// RememberForever caches a value forever (until manually removed)
func (cd *CacheDecorator) RememberForever(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error)) error {
	return cd.GetOrSet(ctx, key, dest, fn, 0)
}

// InvalidatePattern invalidates cache entries matching a pattern
func (cd *CacheDecorator) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := cd.buildKey(pattern)

	cd.logger.Info("Invalidating cache pattern", logging.String("pattern", fullPattern))

	return cd.cache.DeletePattern(ctx, fullPattern)
}

// ClearAll clears all cache entries
func (cd *CacheDecorator) ClearAll(ctx context.Context) error {
	cd.logger.Info("Clearing all cache entries")

	return cd.cache.Clear(ctx)
}

// GetCacheStats returns cache statistics
func (cd *CacheDecorator) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	cd.logger.Info("Getting cache statistics")

	stats := make(map[string]interface{})

	// Get strategy stats if enabled
	if cd.config.EnableStrategies {
		// For now, just add basic strategy info
		stats["strategies"] = map[string]interface{}{
			"enabled": true,
			"type":    cd.config.StrategyType,
		}
	}

	// Add basic cache info
	stats["config"] = map[string]interface{}{
		"default_ttl":       cd.config.DefaultTTL.String(),
		"entity_ttl":        cd.config.EntityTTL.String(),
		"list_ttl":          cd.config.ListTTL.String(),
		"enable_strategies": cd.config.EnableStrategies,
		"strategy_type":     cd.config.StrategyType,
		"enable_metrics":    cd.config.EnableMetrics,
		"key_prefix":        cd.config.KeyPrefix,
	}

	return stats, nil
}

// HealthCheck performs a cache health check
func (cd *CacheDecorator) HealthCheck(ctx context.Context) error {
	cd.logger.Info("Performing cache health check")

	// Test cache connection
	err := cd.cache.Ping(ctx)
	if err != nil {
		cd.logger.Error("Cache health check failed", logging.Error(err))
		return fmt.Errorf("cache health check failed: %w", err)
	}

	cd.logger.Info("Cache health check passed")
	return nil
}

// Close closes the cache decorator
func (cd *CacheDecorator) Close() error {
	cd.logger.Info("Closing cache decorator")

	return cd.cache.Close()
}

// GetCache returns the underlying cache instance
func (cd *CacheDecorator) GetCache() cache.Cache {
	return cd.cache
}

// GetStrategyManager returns the strategy manager
func (cd *CacheDecorator) GetStrategyManager() *strategies.StrategyManager {
	return cd.strategyManager
}

// GetConfig returns the decorator configuration
func (cd *CacheDecorator) GetConfig() *CacheDecoratorConfig {
	return cd.config
}

// buildKey builds a full cache key with prefix
func (cd *CacheDecorator) buildKey(key string) string {
	if cd.config.KeyPrefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", cd.config.KeyPrefix, key)
}

// SetKeyPrefix sets a new key prefix
func (cd *CacheDecorator) SetKeyPrefix(prefix string) {
	cd.config.KeyPrefix = prefix
}

// UpdateConfig updates the decorator configuration
func (cd *CacheDecorator) UpdateConfig(config *CacheDecoratorConfig) {
	if config != nil {
		cd.config = config
	}
}

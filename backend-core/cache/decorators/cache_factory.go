package decorators

import (
	"fmt"
	"time"

	"backend-core/cache"
	"backend-core/config"
	"backend-core/logging"
)

// CacheDecoratorFactory creates cache decorators with different configurations
type CacheDecoratorFactory struct {
	logger *logging.Logger
}

// NewCacheDecoratorFactory creates a new cache decorator factory
func NewCacheDecoratorFactory(logger *logging.Logger) *CacheDecoratorFactory {
	return &CacheDecoratorFactory{
		logger: logger,
	}
}

// CreateDefaultDecorator creates a cache decorator with default configuration
func (f *CacheDecoratorFactory) CreateDefaultDecorator(cfg config.RedisConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating default cache decorator")

	return NewCacheDecorator(cfg, f.logger, DefaultCacheDecoratorConfig())
}

// CreateHighPerformanceDecorator creates a cache decorator optimized for high performance
func (f *CacheDecoratorFactory) CreateHighPerformanceDecorator(cfg config.RedisConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating high-performance cache decorator")

	config := &CacheDecoratorConfig{
		DefaultTTL:       30 * time.Minute, // Longer TTL for better hit rate
		EntityTTL:        60 * time.Minute, // Longer entity TTL
		ListTTL:          10 * time.Minute, // Longer list TTL
		EnableStrategies: true,
		StrategyType:     cache.StrategyCacheAside, // Cache-aside for better performance
		EnableMetrics:    true,
		KeyPrefix:        cache.KeyPrefixPerformance,
	}

	return NewCacheDecorator(cfg, f.logger, config)
}

// CreateConsistencyDecorator creates a cache decorator optimized for consistency
func (f *CacheDecoratorFactory) CreateConsistencyDecorator(cfg config.RedisConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating consistency-optimized cache decorator")

	config := &CacheDecoratorConfig{
		DefaultTTL:       5 * time.Minute,  // Shorter TTL for better consistency
		EntityTTL:        10 * time.Minute, // Shorter entity TTL
		ListTTL:          2 * time.Minute,  // Shorter list TTL
		EnableStrategies: true,
		StrategyType:     cache.StrategyWriteThrough, // Write-through for strong consistency
		EnableMetrics:    true,
		KeyPrefix:        cache.KeyPrefixConsistency,
	}

	return NewCacheDecorator(cfg, f.logger, config)
}

// CreateReadHeavyDecorator creates a cache decorator optimized for read-heavy workloads
func (f *CacheDecoratorFactory) CreateReadHeavyDecorator(cfg config.RedisConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating read-heavy cache decorator")

	config := &CacheDecoratorConfig{
		DefaultTTL:       45 * time.Minute, // Very long TTL for read-heavy
		EntityTTL:        90 * time.Minute, // Very long entity TTL
		ListTTL:          15 * time.Minute, // Long list TTL
		EnableStrategies: true,
		StrategyType:     cache.StrategyReadThrough, // Read-through for read optimization
		EnableMetrics:    true,
		KeyPrefix:        cache.KeyPrefixRead,
	}

	return NewCacheDecorator(cfg, f.logger, config)
}

// CreateWriteHeavyDecorator creates a cache decorator optimized for write-heavy workloads
func (f *CacheDecoratorFactory) CreateWriteHeavyDecorator(cfg config.RedisConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating write-heavy cache decorator")

	config := &CacheDecoratorConfig{
		DefaultTTL:       2 * time.Minute, // Short TTL for write-heavy
		EntityTTL:        5 * time.Minute, // Short entity TTL
		ListTTL:          1 * time.Minute, // Very short list TTL
		EnableStrategies: true,
		StrategyType:     "write_behind", // Write-behind for write optimization
		EnableMetrics:    true,
		KeyPrefix:        "write",
	}

	return NewCacheDecorator(cfg, f.logger, config)
}

// CreateCustomDecorator creates a cache decorator with custom configuration
func (f *CacheDecoratorFactory) CreateCustomDecorator(cfg config.RedisConfig, decoratorConfig *CacheDecoratorConfig) (*CacheDecorator, error) {
	f.logger.Info("Creating custom cache decorator")

	if decoratorConfig == nil {
		return nil, fmt.Errorf("decorator config cannot be nil")
	}

	return NewCacheDecorator(cfg, f.logger, decoratorConfig)
}

// CreateDecoratorFromProfile creates a cache decorator based on a profile name
func (f *CacheDecoratorFactory) CreateDecoratorFromProfile(cfg config.RedisConfig, profile string) (*CacheDecorator, error) {
	f.logger.Info("Creating cache decorator from profile", logging.String("profile", profile))

	switch profile {
	case cache.ProfileDefault:
		return f.CreateDefaultDecorator(cfg)
	case cache.ProfileHighPerformance:
		return f.CreateHighPerformanceDecorator(cfg)
	case cache.ProfileConsistency:
		return f.CreateConsistencyDecorator(cfg)
	case cache.ProfileReadHeavy:
		return f.CreateReadHeavyDecorator(cfg)
	case cache.ProfileWriteHeavy:
		return f.CreateWriteHeavyDecorator(cfg)
	default:
		return nil, fmt.Errorf("unknown profile: %s", profile)
	}
}

// AvailableProfiles returns a list of available cache decorator profiles
func (f *CacheDecoratorFactory) AvailableProfiles() []string {
	return cache.GetAvailableProfiles()
}

// ProfileDescription returns a description of a cache decorator profile
func (f *CacheDecoratorFactory) ProfileDescription(profile string) string {
	if desc, exists := cache.ProfileDescriptions[profile]; exists {
		return desc
	}
	return "Unknown profile"
}

// CreateDecoratorWithKeyPrefix creates a decorator with a specific key prefix
func (f *CacheDecoratorFactory) CreateDecoratorWithKeyPrefix(cfg config.RedisConfig, keyPrefix string) (*CacheDecorator, error) {
	f.logger.Info("Creating cache decorator with key prefix", logging.String("key_prefix", keyPrefix))

	config := DefaultCacheDecoratorConfig()
	config.KeyPrefix = keyPrefix

	return NewCacheDecorator(cfg, f.logger, config)
}

// CreateDecoratorForService creates a decorator optimized for a specific service
func (f *CacheDecoratorFactory) CreateDecoratorForService(cfg config.RedisConfig, serviceName string, workloadType string) (*CacheDecorator, error) {
	f.logger.Info("Creating cache decorator for service",
		logging.String("service", serviceName),
		logging.String("workload", workloadType))

	// Create base config
	config := DefaultCacheDecoratorConfig()
	config.KeyPrefix = serviceName

	// Adjust configuration based on workload type
	switch workloadType {
	case "read-heavy":
		config.EntityTTL = 60 * time.Minute
		config.ListTTL = 15 * time.Minute
		config.StrategyType = "read_through"
	case "write-heavy":
		config.EntityTTL = 5 * time.Minute
		config.ListTTL = 2 * time.Minute
		config.StrategyType = "write_behind"
	case "balanced":
		config.EntityTTL = 30 * time.Minute
		config.ListTTL = 5 * time.Minute
		config.StrategyType = "write_through"
	case "high-performance":
		config.EntityTTL = 45 * time.Minute
		config.ListTTL = 10 * time.Minute
		config.StrategyType = "cache_aside"
	default:
		// Use default configuration
	}

	return NewCacheDecorator(cfg, f.logger, config)
}

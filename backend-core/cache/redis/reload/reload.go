package reload

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// DataSource represents a source of data for cache reloading
type DataSource interface {
	// LoadData loads data from the source
	LoadData(ctx context.Context, key string) (interface{}, error)

	// LoadDataBatch loads multiple data items from the source
	LoadDataBatch(ctx context.Context, keys []string) (map[string]interface{}, error)

	// LoadAllData loads all data from the source
	LoadAllData(ctx context.Context) (map[string]interface{}, error)

	// GetDataKeys returns all available data keys
	GetDataKeys(ctx context.Context) ([]string, error)

	// ValidateData validates loaded data
	ValidateData(data interface{}) error
}

// CacheReloader defines the interface for cache reloading
type CacheReloader interface {
	// Reload reloads cache data
	Reload(ctx context.Context, key string) error

	// ReloadBatch reloads multiple cache entries
	ReloadBatch(ctx context.Context, keys []string) error

	// ReloadAll reloads entire cache
	ReloadAll(ctx context.Context) error

	// GetReloadStrategy returns the current reload strategy
	GetReloadStrategy() CacheReloadStrategy

	// SetReloadStrategy sets the reload strategy
	SetReloadStrategy(strategy CacheReloadStrategy)

	// GetDataSource returns the data source
	GetDataSource() DataSource

	// SetDataSource sets the data source
	SetDataSource(source DataSource)
}

// CacheInvalidator defines the interface for cache invalidation
type CacheInvalidator interface {
	// Invalidate invalidates a cache entry
	Invalidate(ctx context.Context, key string) error

	// InvalidatePattern invalidates cache entries matching a pattern
	InvalidatePattern(ctx context.Context, pattern string) error

	// InvalidateAll invalidates all cache entries
	InvalidateAll(ctx context.Context) error

	// InvalidateAndReload invalidates and reloads a cache entry
	InvalidateAndReload(ctx context.Context, key string) error
}

// CacheWarmer defines the interface for cache warming
type CacheWarmer interface {
	// WarmUp warms up the cache with data
	WarmUp(ctx context.Context) error

	// WarmUpKeys warms up specific cache keys
	WarmUpKeys(ctx context.Context, keys []string) error

	// IsWarmedUp checks if cache is warmed up
	IsWarmedUp(ctx context.Context) (bool, error)
}

// CacheReloadStrategy defines how cache should be reloaded
type CacheReloadStrategy string

const (
	// StrategyRefresh refreshes cache in place
	StrategyRefresh CacheReloadStrategy = "refresh"
	// StrategyReplace replaces entire cache
	StrategyReplace CacheReloadStrategy = "replace"
	// StrategyLazy loads data on demand
	StrategyLazy CacheReloadStrategy = "lazy"
	// StrategyScheduled reloads on schedule
	StrategyScheduled CacheReloadStrategy = "scheduled"
)

// CacheReloadTrigger defines what triggers a cache reload
type CacheReloadTrigger string

const (
	// TriggerManual manual reload trigger
	TriggerManual CacheReloadTrigger = "manual"
	// TriggerTTLExpiry TTL expiry trigger
	TriggerTTLExpiry CacheReloadTrigger = "ttl_expiry"
	// TriggerDataChange data change trigger
	TriggerDataChange CacheReloadTrigger = "data_change"
	// TriggerScheduled scheduled trigger
	TriggerScheduled CacheReloadTrigger = "scheduled"
	// TriggerCacheMiss cache miss trigger
	TriggerCacheMiss CacheReloadTrigger = "cache_miss"
)

// CacheReloadConfig holds configuration for cache reloading
type CacheReloadConfig struct {
	// Strategy defines the reload strategy
	Strategy CacheReloadStrategy `mapstructure:"strategy" json:"strategy" yaml:"strategy"`

	// TTL defines the time-to-live for reloaded data
	TTL time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl"`

	// BatchSize defines the batch size for batch operations
	BatchSize int `mapstructure:"batch_size" json:"batch_size" yaml:"batch_size"`

	// MaxRetries defines the maximum number of retries
	MaxRetries int `mapstructure:"max_retries" json:"max_retries" yaml:"max_retries"`

	// RetryDelay defines the delay between retries
	RetryDelay time.Duration `mapstructure:"retry_delay" json:"retry_delay" yaml:"retry_delay"`

	// EnableLazyLoading enables lazy loading
	EnableLazyLoading bool `mapstructure:"enable_lazy_loading" json:"enable_lazy_loading" yaml:"enable_lazy_loading"`

	// EnableScheduledReload enables scheduled reloading
	EnableScheduledReload bool `mapstructure:"enable_scheduled_reload" json:"enable_scheduled_reload" yaml:"enable_scheduled_reload"`

	// ReloadInterval defines the interval for scheduled reloading
	ReloadInterval time.Duration `mapstructure:"reload_interval" json:"reload_interval" yaml:"reload_interval"`

	// EnableWarmUp enables cache warming
	EnableWarmUp bool `mapstructure:"enable_warm_up" json:"enable_warm_up" yaml:"enable_warm_up"`

	// WarmUpKeys defines specific keys to warm up
	WarmUpKeys []string `mapstructure:"warm_up_keys" json:"warm_up_keys" yaml:"warm_up_keys"`
}

// ReloadResult represents the result of a cache reload operation
type ReloadResult struct {
	// Success indicates if the reload was successful
	Success bool `json:"success"`

	// KeysReloaded is the number of keys reloaded
	KeysReloaded int `json:"keys_reloaded"`

	// KeysFailed is the number of keys that failed to reload
	KeysFailed int `json:"keys_failed"`

	// Errors contains any errors that occurred
	Errors []error `json:"errors,omitempty"`

	// Duration is the time taken for the reload
	Duration time.Duration `json:"duration"`

	// Timestamp is when the reload occurred
	Timestamp time.Time `json:"timestamp"`
}

// ReloadMetrics holds metrics for cache reloading
type ReloadMetrics struct {
	// TotalReloads is the total number of reloads
	TotalReloads int64 `json:"total_reloads"`

	// SuccessfulReloads is the number of successful reloads
	SuccessfulReloads int64 `json:"successful_reloads"`

	// FailedReloads is the number of failed reloads
	FailedReloads int64 `json:"failed_reloads"`

	// AverageReloadTime is the average time for reloads
	AverageReloadTime time.Duration `json:"average_reload_time"`

	// LastReloadTime is when the last reload occurred
	LastReloadTime time.Time `json:"last_reload_time"`

	// CacheHitRate is the cache hit rate
	CacheHitRate float64 `json:"cache_hit_rate"`
}

// RedisCacheReloader implements cache reloading for Redis
type RedisCacheReloader struct {
	client   redis.UniversalClient
	config   *CacheReloadConfig
	source   DataSource
	metrics  *ReloadMetrics
	mu       sync.RWMutex
	stopChan chan struct{}
}

// NewRedisCacheReloader creates a new Redis cache reloader
func NewRedisCacheReloader(client redis.UniversalClient, config *CacheReloadConfig, source DataSource) *RedisCacheReloader {
	reloader := &RedisCacheReloader{
		client:   client,
		config:   config,
		source:   source,
		metrics:  &ReloadMetrics{},
		stopChan: make(chan struct{}),
	}

	// Start scheduled reloading if enabled
	if config.EnableScheduledReload && config.ReloadInterval > 0 {
		go reloader.startScheduledReload()
	}

	return reloader
}

// Reload reloads a single cache entry
func (r *RedisCacheReloader) Reload(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := time.Now()
	defer func() {
		r.updateMetrics(time.Since(start), true)
	}()

	// Load data from source
	data, err := r.source.LoadData(ctx, key)
	if err != nil {
		r.updateMetrics(time.Since(start), false)
		return fmt.Errorf("failed to load data for key %s: %w", key, err)
	}

	// Validate data
	if err := r.source.ValidateData(data); err != nil {
		r.updateMetrics(time.Since(start), false)
		return fmt.Errorf("data validation failed for key %s: %w", key, err)
	}

	// Store in cache based on strategy
	switch r.config.Strategy {
	case StrategyRefresh, StrategyReplace:
		err = r.client.Set(ctx, key, data, r.config.TTL).Err()
	case StrategyLazy:
		// For lazy loading, we don't pre-populate the cache
		return nil
	default:
		err = r.client.Set(ctx, key, data, r.config.TTL).Err()
	}

	if err != nil {
		r.updateMetrics(time.Since(start), false)
		return fmt.Errorf("failed to store data for key %s: %w", key, err)
	}

	return nil
}

// ReloadBatch reloads multiple cache entries
func (r *RedisCacheReloader) ReloadBatch(ctx context.Context, keys []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := time.Now()
	result := &ReloadResult{
		Timestamp: time.Now(),
	}

	defer func() {
		result.Duration = time.Since(start)
		r.updateMetrics(result.Duration, result.Success)
	}()

	// Load data in batches
	batchSize := r.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]
		if err := r.reloadBatch(ctx, batchKeys, result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	}

	result.Success = len(result.Errors) == 0
	return nil
}

// ReloadAll reloads the entire cache
func (r *RedisCacheReloader) ReloadAll(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	start := time.Now()
	result := &ReloadResult{
		Timestamp: time.Now(),
	}

	defer func() {
		result.Duration = time.Since(start)
		r.updateMetrics(result.Duration, result.Success)
	}()

	// Get all data from source
	allData, err := r.source.LoadAllData(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to load all data: %w", err))
		result.Success = false
		return err
	}

	// Clear existing cache if using replace strategy
	if r.config.Strategy == StrategyReplace {
		if err := r.client.FlushDB(ctx).Err(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to clear cache: %w", err))
		}
	}

	// Store all data
	for key, data := range allData {
		if err := r.client.Set(ctx, key, data, r.config.TTL).Err(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to store key %s: %w", key, err))
			result.KeysFailed++
		} else {
			result.KeysReloaded++
		}
	}

	result.Success = len(result.Errors) == 0
	return nil
}

// reloadBatch reloads a batch of keys
func (r *RedisCacheReloader) reloadBatch(ctx context.Context, keys []string, result *ReloadResult) error {
	// Load data from source
	dataMap, err := r.source.LoadDataBatch(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to load batch data: %w", err)
	}

	// Store data in cache
	for key, data := range dataMap {
		// Validate data
		if err := r.source.ValidateData(data); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("data validation failed for key %s: %w", key, err))
			result.KeysFailed++
			continue
		}

		// Store in cache
		if err := r.client.Set(ctx, key, data, r.config.TTL).Err(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to store key %s: %w", key, err))
			result.KeysFailed++
		} else {
			result.KeysReloaded++
		}
	}

	return nil
}

// GetReloadStrategy returns the current reload strategy
func (r *RedisCacheReloader) GetReloadStrategy() CacheReloadStrategy {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config.Strategy
}

// SetReloadStrategy sets the reload strategy
func (r *RedisCacheReloader) SetReloadStrategy(strategy CacheReloadStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config.Strategy = strategy
}

// GetDataSource returns the data source
func (r *RedisCacheReloader) GetDataSource() DataSource {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.source
}

// SetDataSource sets the data source
func (r *RedisCacheReloader) SetDataSource(source DataSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.source = source
}

// startScheduledReload starts the scheduled reload process
func (r *RedisCacheReloader) startScheduledReload() {
	ticker := time.NewTicker(r.config.ReloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), r.config.ReloadInterval)
			if err := r.ReloadAll(ctx); err != nil {
				// Log error but continue
				fmt.Printf("Scheduled reload failed: %v\n", err)
			}
			cancel()
		case <-r.stopChan:
			return
		}
	}
}

// Stop stops the scheduled reload process
func (r *RedisCacheReloader) Stop() {
	close(r.stopChan)
}

// updateMetrics updates the reload metrics
func (r *RedisCacheReloader) updateMetrics(duration time.Duration, success bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.metrics.TotalReloads++
	if success {
		r.metrics.SuccessfulReloads++
	} else {
		r.metrics.FailedReloads++
	}

	// Update average reload time
	if r.metrics.TotalReloads > 0 {
		totalTime := r.metrics.AverageReloadTime * time.Duration(r.metrics.TotalReloads-1)
		r.metrics.AverageReloadTime = (totalTime + duration) / time.Duration(r.metrics.TotalReloads)
	}

	r.metrics.LastReloadTime = time.Now()
}

// GetMetrics returns the current reload metrics
func (r *RedisCacheReloader) GetMetrics() *ReloadMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *r.metrics
	return &metrics
}

// ResetMetrics resets the reload metrics
func (r *RedisCacheReloader) ResetMetrics() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = &ReloadMetrics{}
}

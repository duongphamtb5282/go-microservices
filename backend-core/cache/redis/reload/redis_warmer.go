package reload

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCacheWarmer implements cache warming for Redis
type RedisCacheWarmer struct {
	client   redis.UniversalClient
	reloader CacheReloader
	config   *CacheReloadConfig
	warmedUp bool
	mu       sync.RWMutex
}

// NewRedisCacheWarmer creates a new Redis cache warmer
func NewRedisCacheWarmer(client redis.UniversalClient, reloader CacheReloader, config *CacheReloadConfig) *RedisCacheWarmer {
	return &RedisCacheWarmer{
		client:   client,
		reloader: reloader,
		config:   config,
		warmedUp: false,
	}
}

// WarmUp warms up the cache with data
func (w *RedisCacheWarmer) WarmUp(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	start := time.Now()
	defer func() {
		fmt.Printf("Cache warm-up completed in %v\n", time.Since(start))
	}()

	// Check if already warmed up
	if w.warmedUp {
		return nil
	}

	// Get data source
	source := w.reloader.GetDataSource()
	if source == nil {
		return fmt.Errorf("no data source available for warm-up")
	}

	// Get all data keys
	keys, err := source.GetDataKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get data keys: %w", err)
	}

	// Warm up specific keys if configured
	if len(w.config.WarmUpKeys) > 0 {
		keys = w.config.WarmUpKeys
	}

	// Warm up the cache
	if err := w.warmUpKeys(ctx, keys); err != nil {
		return fmt.Errorf("failed to warm up cache: %w", err)
	}

	w.warmedUp = true
	return nil
}

// WarmUpKeys warms up specific cache keys
func (w *RedisCacheWarmer) WarmUpKeys(ctx context.Context, keys []string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.warmUpKeys(ctx, keys)
}

// warmUpKeys internal method to warm up keys
func (w *RedisCacheWarmer) warmUpKeys(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Get data source
	source := w.reloader.GetDataSource()
	if source == nil {
		return fmt.Errorf("no data source available for warm-up")
	}

	// Load data in batches
	batchSize := w.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batchKeys := keys[i:end]
		if err := w.warmUpBatch(ctx, source, batchKeys); err != nil {
			return fmt.Errorf("failed to warm up batch: %w", err)
		}
	}

	return nil
}

// warmUpBatch warms up a batch of keys
func (w *RedisCacheWarmer) warmUpBatch(ctx context.Context, source DataSource, keys []string) error {
	// Load data from source
	dataMap, err := source.LoadDataBatch(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to load batch data: %w", err)
	}

	// Store data in cache
	for key, data := range dataMap {
		// Validate data
		if err := source.ValidateData(data); err != nil {
			return fmt.Errorf("data validation failed for key %s: %w", key, err)
		}

		// Store in cache
		if err := w.client.Set(ctx, key, data, w.config.TTL).Err(); err != nil {
			return fmt.Errorf("failed to store key %s: %w", key, err)
		}
	}

	return nil
}

// IsWarmedUp checks if cache is warmed up
func (w *RedisCacheWarmer) IsWarmedUp(ctx context.Context) (bool, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Check if we have a warm-up marker
	exists, err := w.client.Exists(ctx, "cache:warmed_up").Result()
	if err != nil {
		return false, fmt.Errorf("failed to check warm-up status: %w", err)
	}

	return exists > 0 || w.warmedUp, nil
}

// SetWarmedUp sets the warmed up status
func (w *RedisCacheWarmer) SetWarmedUp(ctx context.Context, warmedUp bool) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.warmedUp = warmedUp

	if warmedUp {
		// Set a marker in cache
		if err := w.client.Set(ctx, "cache:warmed_up", true, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to set warm-up marker: %w", err)
		}
	} else {
		// Remove the marker
		if err := w.client.Del(ctx, "cache:warmed_up").Err(); err != nil {
			return fmt.Errorf("failed to remove warm-up marker: %w", err)
		}
	}

	return nil
}

// WarmUpWithStrategy warms up cache using a specific strategy
func (w *RedisCacheWarmer) WarmUpWithStrategy(ctx context.Context, strategy CacheReloadStrategy) error {
	// Store original strategy
	originalStrategy := w.reloader.GetReloadStrategy()
	defer w.reloader.SetReloadStrategy(originalStrategy)

	// Set the strategy
	w.reloader.SetReloadStrategy(strategy)

	// Perform warm-up
	return w.WarmUp(ctx)
}

// WarmUpSelective warms up only specific keys based on a filter function
func (w *RedisCacheWarmer) WarmUpSelective(ctx context.Context, keyFilter func(string) bool) error {
	// Get data source
	source := w.reloader.GetDataSource()
	if source == nil {
		return fmt.Errorf("no data source available for warm-up")
	}

	// Get all data keys
	keys, err := source.GetDataKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get data keys: %w", err)
	}

	// Filter keys
	var filteredKeys []string
	for _, key := range keys {
		if keyFilter(key) {
			filteredKeys = append(filteredKeys, key)
		}
	}

	// Warm up filtered keys
	return w.WarmUpKeys(ctx, filteredKeys)
}

// GetWarmUpStats returns statistics about cache warm-up
func (w *RedisCacheWarmer) GetWarmUpStats(ctx context.Context) (map[string]interface{}, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Get total number of keys
	totalKeys, err := w.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	// Check if warmed up
	isWarmedUp, err := w.IsWarmedUp(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check warm-up status: %w", err)
	}

	stats := map[string]interface{}{
		"is_warmed_up": isWarmedUp,
		"total_keys":   totalKeys,
		"timestamp":    time.Now(),
	}

	return stats, nil
}

package operations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// CacheWarmer handles intelligent cache warming strategies
type CacheWarmer struct {
	client     redis.UniversalClient
	logger     *zap.Logger
	dataSource DataSource
	strategies map[string]WarmingStrategy
	mutex      sync.RWMutex
}

// DataSource interface for loading data for cache warming
type DataSource interface {
	LoadData(ctx context.Context, keys []string) (map[string]interface{}, error)
	LoadAllKeys(ctx context.Context, pattern string) ([]string, error)
}

// WarmingStrategy defines how to warm specific types of data
type WarmingStrategy interface {
	GetKeys(ctx context.Context) ([]string, error)
	GetTTL() time.Duration
	GetBatchSize() int
	ShouldWarm(ctx context.Context) bool
}

// PriorityWarmingStrategy warms cache based on priority and access patterns
type PriorityWarmingStrategy struct {
	pattern     string
	ttl         time.Duration
	batchSize   int
	priority    int
	lastAccess  time.Time
	accessCount int64
	dataSource  DataSource
}

// NewCacheWarmer creates a new cache warmer
func NewCacheWarmer(client redis.UniversalClient, dataSource DataSource, logger *zap.Logger) *CacheWarmer {
	return &CacheWarmer{
		client:     client,
		logger:     logger,
		dataSource: dataSource,
		strategies: make(map[string]WarmingStrategy),
	}
}

// NewPriorityWarmingStrategy creates a new priority-based warming strategy
func NewPriorityWarmingStrategy(pattern string, ttl time.Duration, batchSize, priority int, dataSource DataSource) *PriorityWarmingStrategy {
	return &PriorityWarmingStrategy{
		pattern:    pattern,
		ttl:        ttl,
		batchSize:  batchSize,
		priority:   priority,
		lastAccess: time.Now(),
		dataSource: dataSource,
	}
}

// RegisterStrategy registers a warming strategy
func (cw *CacheWarmer) RegisterStrategy(name string, strategy WarmingStrategy) {
	cw.mutex.Lock()
	defer cw.mutex.Unlock()
	cw.strategies[name] = strategy
	cw.logger.Info("Registered cache warming strategy",
		zap.String("name", name),
		zap.Duration("ttl", strategy.GetTTL()),
		zap.Int("batch_size", strategy.GetBatchSize()),
	)
}

// WarmUp warms up cache using registered strategies
func (cw *CacheWarmer) WarmUp(ctx context.Context) error {
	cw.mutex.RLock()
	strategies := make(map[string]WarmingStrategy)
	for name, strategy := range cw.strategies {
		strategies[name] = strategy
	}
	cw.mutex.RUnlock()

	var errors []error
	var errMutex sync.Mutex
	var wg sync.WaitGroup

	for name, strategy := range strategies {
		if !strategy.ShouldWarm(ctx) {
			continue
		}

		wg.Add(1)
		go func(name string, strategy WarmingStrategy) {
			defer wg.Done()

			start := time.Now()
			if err := cw.warmStrategy(ctx, name, strategy); err != nil {
				cw.logger.Error("Failed to warm cache strategy",
					zap.String("strategy", name),
					zap.Error(err),
					zap.Duration("duration", time.Since(start)),
				)
				errMutex.Lock()
				errors = append(errors, fmt.Errorf("strategy %s: %w", name, err))
				errMutex.Unlock()
			} else {
				cw.logger.Info("Successfully warmed cache strategy",
					zap.String("strategy", name),
					zap.Duration("duration", time.Since(start)),
				)
			}
		}(name, strategy)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("multiple cache warming errors: %v", errors)
	}

	return nil
}

// WarmUpStrategy warms up a specific strategy
func (cw *CacheWarmer) WarmUpStrategy(ctx context.Context, name string) error {
	cw.mutex.RLock()
	strategy, exists := cw.strategies[name]
	cw.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("strategy %s not found", name)
	}

	return cw.warmStrategy(ctx, name, strategy)
}

// warmStrategy implements the warming logic for a strategy
func (cw *CacheWarmer) warmStrategy(ctx context.Context, name string, strategy WarmingStrategy) error {
	// Get keys to warm
	keys, err := strategy.GetKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get keys for strategy %s: %w", name, err)
	}

	if len(keys) == 0 {
		cw.logger.Debug("No keys to warm for strategy", zap.String("strategy", name))
		return nil
	}

	// Load data in batches
	batchSize := strategy.GetBatchSize()
	ttl := strategy.GetTTL()

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}

		batch := keys[i:end]
		if err := cw.warmBatch(ctx, batch, ttl, name); err != nil {
			return fmt.Errorf("failed to warm batch for strategy %s: %w", name, err)
		}

		// Small delay between batches to avoid overwhelming the data source
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// warmBatch warms a batch of keys
func (cw *CacheWarmer) warmBatch(ctx context.Context, keys []string, ttl time.Duration, strategyName string) error {
	// Load data from source
	data, err := cw.dataSource.LoadData(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}

	// Use pipeline for efficient writes
	pipe := cw.client.Pipeline()
	for key, value := range data {
		pipe.Set(ctx, key, value, ttl)
	}

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}

	cw.logger.Debug("Warmed batch",
		zap.String("strategy", strategyName),
		zap.Int("keys", len(keys)),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// GetWarmingStats returns statistics about cache warming
func (cw *CacheWarmer) GetWarmingStats() map[string]interface{} {
	cw.mutex.RLock()
	defer cw.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["registered_strategies"] = len(cw.strategies)

	for name, strategy := range cw.strategies {
		stats[fmt.Sprintf("strategy_%s_batch_size", name)] = strategy.GetBatchSize()
		stats[fmt.Sprintf("strategy_%s_ttl", name)] = strategy.GetTTL().String()
	}

	return stats
}

// Implementation of PriorityWarmingStrategy methods
func (pws *PriorityWarmingStrategy) GetKeys(ctx context.Context) ([]string, error) {
	// This would typically query a database or external system
	// For now, return a pattern-based key list
	// In a real implementation, this would be more sophisticated
	keys, err := pws.dataSource.LoadAllKeys(ctx, pws.pattern)
	if err != nil {
		return nil, err
	}

	// Sort by priority and access patterns
	// This is a simplified implementation
	return keys, nil
}

func (pws *PriorityWarmingStrategy) GetTTL() time.Duration {
	return pws.ttl
}

func (pws *PriorityWarmingStrategy) GetBatchSize() int {
	return pws.batchSize
}

func (pws *PriorityWarmingStrategy) ShouldWarm(ctx context.Context) bool {
	// Warm up if enough time has passed or access count is high
	timeSinceLastWarm := time.Since(pws.lastAccess)
	shouldWarmByTime := timeSinceLastWarm > 1*time.Hour
	shouldWarmByAccess := pws.accessCount > 1000

	return shouldWarmByTime || shouldWarmByAccess
}

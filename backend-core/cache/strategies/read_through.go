package strategies

import (
	"context"
	"fmt"
	"time"

	"backend-core/cache"
)

// ReadThroughStrategy implements read-through caching
type ReadThroughStrategy struct {
	name       string
	cache      cache.Cache
	dataSource cache.DataSource
	stats      *StrategyStats
}

// NewReadThroughStrategy creates a new read-through strategy
func NewReadThroughStrategy(name string, cache cache.Cache, dataSource cache.DataSource) *ReadThroughStrategy {
	return &ReadThroughStrategy{
		name:       name,
		cache:      cache,
		dataSource: dataSource,
		stats:      &StrategyStats{Name: name},
	}
}

func (s *ReadThroughStrategy) GetName() string {
	return s.name
}

func (s *ReadThroughStrategy) GetDescription() string {
	return "Read-through strategy: always reads from cache first, falls back to data source"
}

func (s *ReadThroughStrategy) Read(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		s.stats.AverageReadTime = time.Since(start)
		s.stats.LastUsed = time.Now()
	}()

	// Try to read from cache first
	err := s.cache.Get(ctx, key, dest)
	if err == nil {
		s.stats.Hits++
		return nil
	}

	// Cache miss - read from data source
	if err := s.dataSource.LoadData(ctx, key, dest); err != nil {
		s.stats.Misses++
		return fmt.Errorf("failed to load data from source: %w", err)
	}

	// Store in cache for next time
	if err := s.cache.Set(ctx, key, dest, time.Hour); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to cache data for key %s: %v\n", key, err)
	}

	s.stats.Hits++
	return nil
}

func (s *ReadThroughStrategy) Write(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		s.stats.AverageWriteTime = time.Since(start)
		s.stats.LastUsed = time.Now()
	}()

	// Write to data source first
	if err := s.dataSource.StoreData(ctx, key, value); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to store data in source: %w", err)
	}

	// Then write to cache
	if err := s.cache.Set(ctx, key, value, ttl); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to cache data: %w", err)
	}

	s.stats.Writes++
	return nil
}

func (s *ReadThroughStrategy) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		s.stats.AverageWriteTime = time.Since(start)
		s.stats.LastUsed = time.Now()
	}()

	// Delete from data source first
	if err := s.dataSource.DeleteData(ctx, key); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to delete data from source: %w", err)
	}

	// Then delete from cache
	if err := s.cache.Delete(ctx, key); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to delete cached data: %w", err)
	}

	s.stats.Deletes++
	return nil
}

func (s *ReadThroughStrategy) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

func (s *ReadThroughStrategy) Clear(ctx context.Context) error {
	return s.cache.Clear(ctx)
}

func (s *ReadThroughStrategy) GetStats(ctx context.Context) (*StrategyStats, error) {
	return s.stats, nil
}

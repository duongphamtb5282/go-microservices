package strategies

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-core/cache"
)

// WriteBehindStrategy implements write-behind caching
type WriteBehindStrategy struct {
	name       string
	cache      cache.Cache
	dataSource cache.DataSource
	stats      *StrategyStats
	writeQueue chan WriteOperation
	wg         sync.WaitGroup
	stopChan   chan struct{}
}

// WriteOperation represents a write operation
type WriteOperation struct {
	Key   string
	Value interface{}
	TTL   time.Duration
	Ctx   context.Context
}

// NewWriteBehindStrategy creates a new write-behind strategy
func NewWriteBehindStrategy(name string, cache cache.Cache, dataSource cache.DataSource) *WriteBehindStrategy {
	strategy := &WriteBehindStrategy{
		name:       name,
		cache:      cache,
		dataSource: dataSource,
		stats:      &StrategyStats{Name: name},
		writeQueue: make(chan WriteOperation, 1000),
		stopChan:   make(chan struct{}),
	}

	// Start the background writer
	go strategy.backgroundWriter()

	return strategy
}

func (s *WriteBehindStrategy) GetName() string {
	return s.name
}

func (s *WriteBehindStrategy) GetDescription() string {
	return "Write-behind strategy: writes to cache immediately, data source asynchronously"
}

func (s *WriteBehindStrategy) Read(ctx context.Context, key string, dest interface{}) error {
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

func (s *WriteBehindStrategy) Write(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		s.stats.AverageWriteTime = time.Since(start)
		s.stats.LastUsed = time.Now()
	}()

	// Write to cache immediately
	if err := s.cache.Set(ctx, key, value, ttl); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to cache data: %w", err)
	}

	// Queue write to data source
	select {
	case s.writeQueue <- WriteOperation{Key: key, Value: value, TTL: ttl, Ctx: ctx}:
		// Successfully queued
	default:
		// Queue is full, write synchronously
		if err := s.dataSource.StoreData(ctx, key, value); err != nil {
			s.stats.Errors++
			return fmt.Errorf("failed to store data in source: %w", err)
		}
	}

	s.stats.Writes++
	return nil
}

func (s *WriteBehindStrategy) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		s.stats.AverageWriteTime = time.Since(start)
		s.stats.LastUsed = time.Now()
	}()

	// Delete from cache immediately
	if err := s.cache.Delete(ctx, key); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to delete cached data: %w", err)
	}

	// Queue delete to data source
	select {
	case s.writeQueue <- WriteOperation{Key: key, Value: nil, TTL: 0, Ctx: ctx}:
		// Successfully queued
	default:
		// Queue is full, delete synchronously
		if err := s.dataSource.DeleteData(ctx, key); err != nil {
			s.stats.Errors++
			return fmt.Errorf("failed to delete data from source: %w", err)
		}
	}

	s.stats.Deletes++
	return nil
}

func (s *WriteBehindStrategy) Exists(ctx context.Context, key string) (bool, error) {
	return s.cache.Exists(ctx, key)
}

func (s *WriteBehindStrategy) Clear(ctx context.Context) error {
	return s.cache.Clear(ctx)
}

func (s *WriteBehindStrategy) GetStats(ctx context.Context) (*StrategyStats, error) {
	return s.stats, nil
}

// backgroundWriter processes write operations in the background
func (s *WriteBehindStrategy) backgroundWriter() {
	for {
		select {
		case op := <-s.writeQueue:
			s.processWriteOperation(op)
		case <-s.stopChan:
			return
		}
	}
}

// processWriteOperation processes a single write operation
func (s *WriteBehindStrategy) processWriteOperation(op WriteOperation) {
	if op.Value != nil {
		// Store operation
		if err := s.dataSource.StoreData(op.Ctx, op.Key, op.Value); err != nil {
			s.stats.Errors++
			fmt.Printf("Error storing data for key %s: %v\n", op.Key, err)
		}
	} else {
		// Delete operation
		if err := s.dataSource.DeleteData(op.Ctx, op.Key); err != nil {
			s.stats.Errors++
			fmt.Printf("Error deleting data for key %s: %v\n", op.Key, err)
		}
	}
}

// Stop stops the background writer
func (s *WriteBehindStrategy) Stop() {
	close(s.stopChan)
	s.wg.Wait()
}

package operations

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"backend-core/logging"
)

// CacheInvalidator handles intelligent cache invalidation
type CacheInvalidator struct {
	client       redis.UniversalClient
	logger       *logging.Logger
	dependencies map[string][]string // key -> dependent keys
	mutex        sync.RWMutex
}

// InvalidationEvent represents a cache invalidation event
type InvalidationEvent struct {
	Key       string            `json:"key"`
	Pattern   string            `json:"pattern"`
	Strategy  string            `json:"strategy"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
}

// InvalidationStrategy defines how to invalidate cache
type InvalidationStrategy struct {
	Name       string
	Pattern    string
	Delay      time.Duration
	MaxRetries int
	BatchSize  int
}

// NewCacheInvalidator creates a new cache invalidator
func NewCacheInvalidator(client redis.UniversalClient, logger *logging.Logger) *CacheInvalidator {
	return &CacheInvalidator{
		client:       client,
		logger:       logger,
		dependencies: make(map[string][]string),
	}
}

// RegisterDependency registers a dependency between keys
func (ci *CacheInvalidator) RegisterDependency(key string, dependentKeys []string) {
	ci.mutex.Lock()
	defer ci.mutex.Unlock()

	ci.dependencies[key] = dependentKeys

	// Store dependency in Redis for persistence
	ci.storeDependency(key, dependentKeys)
}

// InvalidateKey invalidates a specific key and its dependencies
func (ci *CacheInvalidator) InvalidateKey(ctx context.Context, key string, strategy InvalidationStrategy) error {
	start := time.Now()

	// Get dependent keys
	ci.mutex.RLock()
	dependentKeys, exists := ci.dependencies[key]
	ci.mutex.RUnlock()

	if !exists {
		// Try to load from Redis
		dependentKeys = ci.loadDependency(key)
	}

	// Add the main key
	allKeys := append([]string{key}, dependentKeys...)

	// Invalidate using strategy
	if err := ci.invalidateKeys(ctx, allKeys, strategy); err != nil {
		return fmt.Errorf("failed to invalidate keys: %w", err)
	}

	// Record the invalidation event
	ci.recordInvalidationEvent(ctx, InvalidationEvent{
		Key:       key,
		Strategy:  strategy.Name,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"invalidated_keys": fmt.Sprintf("%d", len(allKeys)),
			"duration":         time.Since(start).String(),
		},
	})

	ci.logger.Info("Cache invalidation completed", "key", key, "strategy", strategy.Name, "keys_invalidated", len(allKeys), "duration", time.Since(start))

	return nil
}

// InvalidatePattern invalidates keys matching a pattern
func (ci *CacheInvalidator) InvalidatePattern(ctx context.Context, pattern string, strategy InvalidationStrategy) error {
	start := time.Now()

	// Find keys matching pattern
	keys, err := ci.findKeysByPattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to find keys by pattern: %w", err)
	}

	if len(keys) == 0 {
		ci.logger.Debug("No keys found for pattern", "pattern", pattern)
		return nil
	}

	// Invalidate keys
	if err := ci.invalidateKeys(ctx, keys, strategy); err != nil {
		return fmt.Errorf("failed to invalidate keys by pattern: %w", err)
	}

	// Record the invalidation event
	ci.recordInvalidationEvent(ctx, InvalidationEvent{
		Pattern:   pattern,
		Strategy:  strategy.Name,
		Timestamp: time.Now(),
		Metadata: map[string]string{
			"invalidated_keys": fmt.Sprintf("%d", len(keys)),
			"duration":         time.Since(start).String(),
		},
	})

	ci.logger.Info("Pattern-based cache invalidation completed", "pattern", pattern, "strategy", strategy.Name, "keys_invalidated", len(keys), "duration", time.Since(start))

	return nil
}

// InvalidateByTags invalidates cache entries by tags
func (ci *CacheInvalidator) InvalidateByTags(ctx context.Context, tags []string, strategy InvalidationStrategy) error {
	if len(tags) == 0 {
		return nil
	}

	start := time.Now()
	var allKeys []string

	// Get keys for each tag
	for _, tag := range tags {
		tagKeys, err := ci.client.SMembers(ctx, fmt.Sprintf("cache:tag:%s", tag)).Result()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get keys for tag %s: %w", tag, err)
		}
		allKeys = append(allKeys, tagKeys...)
	}

	if len(allKeys) == 0 {
		ci.logger.Debug("No keys found for tags", "tags", tags)
		return nil
	}

	// Invalidate keys
	if err := ci.invalidateKeys(ctx, allKeys, strategy); err != nil {
		return fmt.Errorf("failed to invalidate keys by tags: %w", err)
	}

	ci.logger.Info("Tag-based cache invalidation completed", "tags", tags, "strategy", strategy.Name, "keys_invalidated", len(allKeys), "duration", time.Since(start))

	return nil
}

// invalidateKeys performs the actual invalidation with retry logic
func (ci *CacheInvalidator) invalidateKeys(ctx context.Context, keys []string, strategy InvalidationStrategy) error {
	if len(keys) == 0 {
		return nil
	}

	var lastErr error
	batchSize := strategy.BatchSize
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}

	// Retry logic
	for attempt := 0; attempt <= strategy.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(strategy.Delay * time.Duration(attempt))
		}

		// Process in batches
		for i := 0; i < len(keys); i += batchSize {
			end := i + batchSize
			if end > len(keys) {
				end = len(keys)
			}

			batch := keys[i:end]
			if err := ci.invalidateBatch(ctx, batch); err != nil {
				lastErr = err
				ci.logger.Warn("Failed to invalidate batch", "error", err, "attempt", attempt+1, "batch_size", len(batch))
				continue
			}
		}

		// If we get here, all batches succeeded
		return nil
	}

	return fmt.Errorf("failed to invalidate keys after %d attempts: %w", strategy.MaxRetries+1, lastErr)
}

// invalidateBatch invalidates a batch of keys
func (ci *CacheInvalidator) invalidateBatch(ctx context.Context, keys []string) error {
	// Use pipeline for efficient deletion
	pipe := ci.client.Pipeline()

	for _, key := range keys {
		// Delete the key
		pipe.Del(ctx, key)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return ci.removeKeysFromTags(ctx, keys)
}

// findKeysByPattern finds keys matching a pattern
func (ci *CacheInvalidator) findKeysByPattern(ctx context.Context, pattern string) ([]string, error) {
	// Use SCAN for efficient pattern matching
	var keys []string
	var cursor uint64

	for {
		result, cursor, err := ci.client.Scan(ctx, cursor, pattern, 1000).Result()
		if err != nil {
			return nil, err
		}

		keys = append(keys, result...)
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// removeKeysFromTags removes the provided keys from all tag sets
func (ci *CacheInvalidator) removeKeysFromTags(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	var cursor uint64
	for {
		tagKeys, nextCursor, err := ci.client.Scan(ctx, cursor, "cache:tag:*", 1000).Result()
		if err != nil {
			return fmt.Errorf("failed to scan tag keys: %w", err)
		}

		if len(tagKeys) > 0 {
			pipe := ci.client.Pipeline()
			for _, tagKey := range tagKeys {
				for _, key := range keys {
					pipe.SRem(ctx, tagKey, key)
				}
			}
			if _, err := pipe.Exec(ctx); err != nil {
				return fmt.Errorf("failed to remove keys from tag sets: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// storeDependency stores dependency information in Redis
func (ci *CacheInvalidator) storeDependency(key string, dependentKeys []string) {
	ctx := context.Background()

	// Store dependency mapping
	dependencyKey := fmt.Sprintf("cache:dependency:%s", key)
	ci.client.SAdd(ctx, dependencyKey, dependentKeys)

	// Set expiration to clean up old dependencies
	ci.client.Expire(ctx, dependencyKey, 24*time.Hour)
}

// loadDependency loads dependency information from Redis
func (ci *CacheInvalidator) loadDependency(key string) []string {
	ctx := context.Background()
	dependencyKey := fmt.Sprintf("cache:dependency:%s", key)

	keys, err := ci.client.SMembers(ctx, dependencyKey).Result()
	if err != nil {
		ci.logger.Warn("Failed to load dependency from Redis", "key", key, "error", err)
		return nil
	}

	return keys
}

// recordInvalidationEvent records an invalidation event for analytics
func (ci *CacheInvalidator) recordInvalidationEvent(ctx context.Context, event InvalidationEvent) {
	eventJSON, _ := json.Marshal(event)
	ci.client.LPush(ctx, "cache:invalidation_events", eventJSON)
	ci.client.LTrim(ctx, "cache:invalidation_events", 0, 999)          // Keep last 1000 events
	ci.client.Expire(ctx, "cache:invalidation_events", 7*24*time.Hour) // Expire after 7 days
}

// GetInvalidationStats returns invalidation statistics
func (ci *CacheInvalidator) GetInvalidationStats() map[string]interface{} {
	ctx := context.Background()

	// Get recent events
	events, err := ci.client.LRange(ctx, "cache:invalidation_events", 0, -1).Result()
	if err != nil {
		ci.logger.Warn("Failed to get invalidation events", "error", err)
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	stats := map[string]interface{}{
		"total_events": len(events),
		"dependencies": len(ci.dependencies),
	}

	// Parse events for more detailed stats
	var recentEvents []InvalidationEvent
	for _, eventStr := range events {
		var event InvalidationEvent
		if json.Unmarshal([]byte(eventStr), &event) == nil {
			recentEvents = append(recentEvents, event)
		}
	}

	if len(recentEvents) > 0 {
		// Calculate time-based stats
		now := time.Now()
		var last24h, last7d int

		for _, event := range recentEvents {
			hoursAgo := now.Sub(event.Timestamp).Hours()
			if hoursAgo <= 24 {
				last24h++
			}
			if hoursAgo <= 168 { // 7 days
				last7d++
			}
		}

		stats["last_24h"] = last24h
		stats["last_7d"] = last7d
	}

	return stats
}

// Cleanup removes expired dependencies and events
func (ci *CacheInvalidator) Cleanup(ctx context.Context) error {
	ci.logger.Info("Starting cache invalidator cleanup")

	// Clean up old events
	eventsKey := "cache:invalidation_events"
	ci.client.LTrim(ctx, eventsKey, 0, 499) // Keep only 500 most recent
	ci.client.Expire(ctx, eventsKey, 7*24*time.Hour)

	// Clean up old dependencies (older than 24 hours)
	pattern := "cache:dependency:*"
	keys, err := ci.findKeysByPattern(ctx, pattern)
	if err != nil {
		return fmt.Errorf("failed to find dependency keys: %w", err)
	}

	for _, key := range keys {
		ci.client.Expire(ctx, key, 24*time.Hour)
	}

	ci.logger.Info("Cache invalidator cleanup completed", "dependency_keys", len(keys))
	return nil
}

package masking

import (
	"sync"
	"time"
)

// InMemoryCache implements an in-memory cache for masking rules
type InMemoryCache struct {
	items   map[string]cacheItem
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
	cleanup *time.Ticker
	stop    chan struct{}
}

// cacheItem represents a cached item with expiration
type cacheItem struct {
	rule       MaskingRule
	expiration time.Time
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache(maxSize int, ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		items:   make(map[string]cacheItem),
		ttl:     ttl,
		maxSize: maxSize,
		stop:    make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.cleanup = time.NewTicker(ttl / 2)
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a masking rule from cache
func (c *InMemoryCache) Get(field string) (MaskingRule, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[field]
	if !exists {
		return MaskingRule{}, false
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		return MaskingRule{}, false
	}

	return item.rule, true
}

// Set stores a masking rule in cache
func (c *InMemoryCache) Set(field string, rule MaskingRule) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if we need to evict items
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[field] = cacheItem{
		rule:       rule,
		expiration: time.Now().Add(c.ttl),
	}
}

// Clear clears the cache
func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]cacheItem)
}

// Size returns the current cache size
func (c *InMemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// evictOldest removes the oldest item from cache
func (c *InMemoryCache) evictOldest() {
	var oldestField string
	var oldestTime time.Time

	for field, item := range c.items {
		if oldestField == "" || item.expiration.Before(oldestTime) {
			oldestField = field
			oldestTime = item.expiration
		}
	}

	if oldestField != "" {
		delete(c.items, oldestField)
	}
}

// cleanupExpired removes expired items from cache
func (c *InMemoryCache) cleanupExpired() {
	for {
		select {
		case <-c.cleanup.C:
			c.mutex.Lock()
			now := time.Now()
			for field, item := range c.items {
				if now.After(item.expiration) {
					delete(c.items, field)
				}
			}
			c.mutex.Unlock()
		case <-c.stop:
			return
		}
	}
}

// Close stops the cache cleanup goroutine
func (c *InMemoryCache) Close() {
	if c.cleanup != nil {
		c.cleanup.Stop()
	}
	close(c.stop)
}

// NoOpCache implements a no-operation cache
type NoOpCache struct{}

// NewNoOpCache creates a new no-operation cache
func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

// Get always returns false
func (c *NoOpCache) Get(field string) (MaskingRule, bool) {
	return MaskingRule{}, false
}

// Set does nothing
func (c *NoOpCache) Set(field string, rule MaskingRule) {
	// No-op
}

// Clear does nothing
func (c *NoOpCache) Clear() {
	// No-op
}

// Size always returns 0
func (c *NoOpCache) Size() int {
	return 0
}

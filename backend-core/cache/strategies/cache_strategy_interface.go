package strategies

import (
	"context"
	"time"
)

// CacheStrategy defines the interface for cache strategies
type CacheStrategy interface {
	// GetName returns the strategy name
	GetName() string

	// GetDescription returns the strategy
	GetDescription() string

	// Read reads data using the strategy
	Read(ctx context.Context, key string, dest interface{}) error

	// Write writes data using the strategy
	Write(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete deletes data using the strategy
	Delete(ctx context.Context, key string) error

	// Exists checks if key exists using the strategy
	Exists(ctx context.Context, key string) (bool, error)

	// Clear clears all data using the strategy
	Clear(ctx context.Context) error

	// GetStats returns strategy-specific statistics
	GetStats(ctx context.Context) (*StrategyStats, error)
}

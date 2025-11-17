package reload

import (
	"context"
	"time"
)

// CustomReloadStrategy defines a custom reload strategy
type CustomReloadStrategy interface {
	// GetName returns the name of the strategy
	GetName() string

	// GetDescription returns the description of the strategy
	GetDescription() string

	// ShouldReload determines if a key should be reloaded
	ShouldReload(ctx context.Context, key string, lastReload time.Time) bool

	// GetReloadPriority returns the priority for reloading (lower = higher priority)
	GetReloadPriority(ctx context.Context, key string) int

	// PreReload is called before reloading a key
	PreReload(ctx context.Context, key string, data interface{}) error

	// PostReload is called after reloading a key
	PostReload(ctx context.Context, key string, data interface{}) error

	// OnReloadError is called when reloading fails
	OnReloadError(ctx context.Context, key string, err error) error

	// GetBatchSize returns the recommended batch size for this strategy
	GetBatchSize() int

	// GetRetryCount returns the number of retries for this strategy
	GetRetryCount() int

	// GetRetryDelay returns the delay between retries
	GetRetryDelay() time.Duration
}

package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned when a key is not found in the cache
var ErrCacheMiss = errors.New("cache miss")

// Cache defines the interface for cache operations
type Cache interface {
	// Basic operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Counter operations
	Increment(ctx context.Context, key string) (int64, error)
	IncrementBy(ctx context.Context, key string, value int64) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	DecrementBy(ctx context.Context, key string, value int64) (int64, error)

	// List operations
	ListPush(ctx context.Context, key string, value interface{}) error
	ListPop(ctx context.Context, key string, dest interface{}) error
	ListLength(ctx context.Context, key string) (int64, error)

	// Set operations
	SetAdd(ctx context.Context, key string, member interface{}) error
	SetMembers(ctx context.Context, key string) ([]string, error)
	SetIsMember(ctx context.Context, key string, member interface{}) (bool, error)

	// Hash operations
	HashSet(ctx context.Context, key, field string, value interface{}) error
	HashGet(ctx context.Context, key, field string, dest interface{}) error
	HashGetAll(ctx context.Context, key string) (map[string]string, error)

	// Connection operations
	Ping(ctx context.Context) error
	Close() error

	// Clear clears all data from the cache
	Clear(ctx context.Context) error
}

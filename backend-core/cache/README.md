# Cache Package

This package provides a comprehensive caching solution with support for multiple cache backends, cache strategies, and data source integration.

## Package Structure

The cache package has been separated into focused files for better maintainability:

### Core Files

- **`cache.go`** - Core cache interface and error definitions
- **`data_source.go`** - Data source interface for external data loading
- **`cache_manager.go`** - Cache manager implementation with high-level operations

### Implementation Files

- **`redis.go`** - Redis cache implementation
- **`memory.go`** - In-memory cache implementation
- **`decorators/`** - Cache decorators and strategies
- **`strategies/`** - Cache strategies (read-through, write-through, etc.)

## Core Components

### 1. Cache Interface (`cache.go`)

The `Cache` interface defines all cache operations:

```go
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
    Clear(ctx context.Context) error
}
```

### 2. Data Source Interface (`data_source.go`)

The `DataSource` interface defines operations for loading data from external sources:

```go
type DataSource interface {
    // LoadData loads data from the source
    LoadData(ctx context.Context, key string, dest interface{}) error

    // StoreData stores data in the source
    StoreData(ctx context.Context, key string, value interface{}) error

    // DeleteData deletes data from the source
    DeleteData(ctx context.Context, key string) error
}
```

### 3. Cache Manager (`cache_manager.go`)

The `CacheManager` provides high-level cache operations and patterns:

```go
type CacheManager struct {
    cache Cache
}

// High-level operations
func (cm *CacheManager) GetOrSet(ctx context.Context, key string, dest interface{}, setter func() (interface{}, error), expiration time.Duration) error
func (cm *CacheManager) Remember(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error), expiration time.Duration) error
func (cm *CacheManager) RememberForever(ctx context.Context, key string, dest interface{}, fn func() (interface{}, error)) error
func (cm *CacheManager) Forget(ctx context.Context, key string) error
func (cm *CacheManager) Flush(ctx context.Context) error
func (cm *CacheManager) Pull(ctx context.Context, key string, dest interface{}) error
```

## Usage Examples

### Basic Cache Operations

```go
// Create a cache manager
cacheManager := cache.NewCacheManager(redisCache)

// Remember pattern - cache function result
var result string
err := cacheManager.Remember(ctx, "user:123", &result, func() (interface{}, error) {
    // This function will be called if cache miss
    return "user data", nil
}, 10*time.Minute)

// Get or set pattern
err = cacheManager.GetOrSet(ctx, "key", &value, func() (interface{}, error) {
    return expensiveOperation(), nil
}, 5*time.Minute)

// Pull pattern - get and remove
err = cacheManager.Pull(ctx, "temp:key", &value)
```

### Data Source Integration

```go
// Implement DataSource interface
type DatabaseDataSource struct {
    db *sql.DB
}

func (ds *DatabaseDataSource) LoadData(ctx context.Context, key string, dest interface{}) error {
    // Load data from database
    return ds.db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", key).Scan(dest)
}

func (ds *DatabaseDataSource) StoreData(ctx context.Context, key string, value interface{}) error {
    // Store data in database
    _, err := ds.db.ExecContext(ctx, "INSERT INTO users (id, data) VALUES ($1, $2)", key, value)
    return err
}

func (ds *DatabaseDataSource) DeleteData(ctx context.Context, key string) error {
    // Delete data from database
    _, err := ds.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", key)
    return err
}
```

### Cache Strategies

```go
// Read-through strategy
readThroughStrategy := strategies.NewReadThroughStrategy(
    "read-through",
    redisCache,
    databaseDataSource,
)

// Write-through strategy
writeThroughStrategy := strategies.NewWriteThroughStrategy(
    "write-through",
    redisCache,
    databaseDataSource,
)

// Cache-aside strategy
cacheAsideStrategy := strategies.NewCacheAsideStrategy(
    "cache-aside",
    redisCache,
    databaseDataSource,
)
```

## Error Handling

The package defines a standard error for cache misses:

```go
var ErrCacheMiss = errors.New("cache miss")
```

This error should be returned by cache implementations when a key is not found.

## Thread Safety

All cache implementations should be thread-safe and can be used concurrently from multiple goroutines.

## Configuration

Cache implementations can be configured through their respective configuration structures:

```go
// Redis configuration
redisConfig := &config.RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "",
    DB:       0,
}

// Create Redis cache
redisCache := cache.NewRedisCache(redisConfig)
```

## Monitoring

Cache implementations should provide monitoring capabilities:

- **Hit/Miss Ratios**: Track cache effectiveness
- **Response Times**: Monitor cache performance
- **Memory Usage**: Track memory consumption
- **Connection Pool**: Monitor connection health

## Best Practices

### 1. Key Naming

- Use consistent key prefixes
- Include version information
- Use descriptive names

```go
// Good
"user:123:profile"
"session:abc123:data"
"config:app:version"

// Bad
"u123"
"data"
"config"
```

### 2. Expiration Times

- Set appropriate TTL values
- Use different TTLs for different data types
- Consider data freshness requirements

```go
// User data - 1 hour
cache.Set(ctx, "user:123", userData, 1*time.Hour)

// Session data - 30 minutes
cache.Set(ctx, "session:abc", sessionData, 30*time.Minute)

// Configuration - 1 day
cache.Set(ctx, "config:app", configData, 24*time.Hour)
```

### 3. Error Handling

- Always handle cache errors gracefully
- Fall back to data source on cache errors
- Log cache errors for monitoring

```go
err := cache.Get(ctx, key, &value)
if err != nil {
    if err == cache.ErrCacheMiss {
        // Load from data source
        return loadFromDataSource(ctx, key)
    }
    // Log error and fall back
    log.Error("Cache error", err)
    return loadFromDataSource(ctx, key)
}
```

### 4. Memory Management

- Monitor memory usage
- Use appropriate data structures
- Clean up expired keys

### 5. Testing

- Test cache behavior
- Mock cache for unit tests
- Test cache invalidation
- Test concurrent access

## Migration Guide

If you're migrating from the old single-file structure:

1. **Update imports**: No changes needed, all interfaces remain the same
2. **Update implementations**: Ensure they implement the separated interfaces
3. **Update tests**: Tests should continue to work without changes

The separation is purely organizational and doesn't change the public API.

## Future Enhancements

- **Distributed Caching**: Support for distributed cache backends
- **Cache Warming**: Proactive cache population
- **Cache Analytics**: Advanced cache analytics and insights
- **Auto-scaling**: Dynamic cache scaling based on load
- **Multi-tier Caching**: Support for multiple cache tiers

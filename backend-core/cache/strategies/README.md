# Cache Strategies

This package provides a comprehensive set of cache strategies for implementing different caching patterns in Go microservices.

## Available Strategies

### 1. Read-Through Strategy

- **Purpose**: Reads from cache first, falls back to data source on miss
- **Use Case**: When you want to ensure data consistency and reduce data source load
- **Behavior**:
  - On read: Cache → Data Source (if miss) → Update cache
  - On write: Cache + Data Source simultaneously

### 2. Write-Through Strategy

- **Purpose**: Writes to both cache and data source simultaneously
- **Use Case**: When you need strong consistency between cache and data source
- **Behavior**:
  - On read: Cache only
  - On write: Cache + Data Source (both must succeed)

### 3. Write-Behind Strategy

- **Purpose**: Writes to cache immediately, data source asynchronously
- **Use Case**: When you need high write performance and can tolerate eventual consistency
- **Behavior**:
  - On read: Cache only
  - On write: Cache immediately + Queue for data source

### 4. Cache-Aside Strategy

- **Purpose**: Application manages both cache and data source
- **Use Case**: When you need full control over caching logic
- **Behavior**:
  - On read: Check cache → Load from data source if miss → Update cache
  - On write: Update data source → Invalidate cache

## Usage

```go
package main

import (
    "context"
    "time"

    "backend-core/cache"
    "backend-core/cache/strategies"
)

func main() {
    // Create your cache and data source implementations
    var redisCache cache.Cache
    var dataSource cache.DataSource

    // Create strategy manager
    manager := strategies.NewStrategyManager()

    // Create and register strategies
    readThrough := strategies.NewReadThroughStrategy("read-through", redisCache, dataSource)
    writeThrough := strategies.NewWriteThroughStrategy("write-through", redisCache, dataSource)
    writeBehind := strategies.NewWriteBehindStrategy("write-behind", redisCache, dataSource)
    cacheAside := strategies.NewCacheAsideStrategy("cache-aside", redisCache, dataSource)

    manager.RegisterStrategy(readThrough)
    manager.RegisterStrategy(writeThrough)
    manager.RegisterStrategy(writeBehind)
    manager.RegisterStrategy(cacheAside)

    // Set default strategy
    manager.SetDefaultStrategy("read-through")

    // Use strategies
    ctx := context.Background()
    strategy, _ := manager.GetDefaultStrategy()

    // Read data
    var result interface{}
    err := strategy.Read(ctx, "user:123", &result)

    // Write data
    err = strategy.Write(ctx, "user:123", map[string]string{"name": "John"}, 5*time.Minute)

    // Get statistics
    stats, err := strategy.GetStats(ctx)
    fmt.Printf("Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
}
```

## Strategy Selection Guide

| Strategy      | Consistency  | Performance | Complexity | Best For                            |
| ------------- | ------------ | ----------- | ---------- | ----------------------------------- |
| Read-Through  | Strong       | Medium      | Low        | Read-heavy workloads                |
| Write-Through | Strong       | Low         | Low        | Write-heavy, consistency critical   |
| Write-Behind  | Eventual     | High        | Medium     | High-write, eventual consistency OK |
| Cache-Aside   | Configurable | High        | High       | Full control needed                 |

## Statistics

Each strategy tracks the following metrics:

- **Hits**: Number of successful cache reads
- **Misses**: Number of cache misses
- **Writes**: Number of write operations
- **Deletes**: Number of delete operations
- **Errors**: Number of errors encountered
- **Average Read Time**: Average time for read operations
- **Average Write Time**: Average time for write operations
- **Last Used**: Timestamp of last operation

## Error Handling

All strategies implement proper error handling and will return appropriate errors for:

- Cache connection failures
- Data source connection failures
- Serialization/deserialization errors
- Timeout errors
- Invalid parameters

## Thread Safety

All strategies are thread-safe and can be used concurrently from multiple goroutines.

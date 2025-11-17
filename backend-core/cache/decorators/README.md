# Cache Decorators

A comprehensive caching solution for Go microservices using the decorator pattern. This package provides flexible, reusable cache decorators that can be easily integrated into any Go service.

## Features

- **Decorator Pattern**: Clean separation of concerns with cache logic
- **Multiple Strategies**: Read-through, write-through, cache-aside, write-behind
- **HTTP Integration**: Ready-to-use middleware for HTTP applications
- **Configurable TTL**: Different TTL settings for entities, lists, and default values
- **Key Prefixing**: Namespace isolation for different services
- **Health Checks**: Built-in cache health monitoring
- **Statistics**: Detailed cache performance metrics
- **Factory Pattern**: Easy creation of pre-configured decorators

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"

    "backend-core/cache/decorators"
    "backend-core/config"
    "backend-core/logging"
)

func main() {
    // Initialize logger
    logger, _ := logging.NewLogger(&config.LoggingConfig{
        Level:  "info",
        Format: "json",
        Output: []string{"stdout"},
    })

    // Redis configuration
    redisConfig := config.RedisConfig{
        Name:     "my-service",
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    }

    // Create cache decorator
    factory := decorators.NewCacheDecoratorFactory(logger)
    decorator, err := factory.CreateDefaultDecorator(redisConfig)
    if err != nil {
        log.Fatal(err)
    }
    defer decorator.Close()

    // Use the decorator
    ctx := context.Background()

    // Set a value
    err = decorator.Set(ctx, "user:123", map[string]interface{}{
        "id":   "123",
        "name": "John Doe",
    }, 5*time.Minute)

    // Get a value
    var user map[string]interface{}
    err = decorator.Get(ctx, "user:123", &user)
}
```

### HTTP Integration

```go
package main

import (
    "backend-core/cache/decorators"
    "net/http"
    "time"
)

func main() {
    // Create cache decorator
    factory := decorators.NewCacheDecoratorFactory(logger)
    decorator, _ := factory.CreateDecoratorForService(redisConfig, "auth-service", "balanced")

    // Create HTTP middleware
    cacheMiddleware := decorators.NewHTTPCacheMiddleware(decorator, logger)

    mux := http.NewServeMux()

    // Cache individual entities
    mux.Handle("/api/v1/users/", cacheMiddleware.CacheEntity("user")(getUserHandler))

    // Cache lists
    mux.Handle("/api/v1/users", cacheMiddleware.CacheGet(5*time.Minute)(listUsersHandler))

    // Invalidate cache on mutations
    mux.Handle("/api/v1/users", cacheMiddleware.InvalidateCache("user")(createUserHandler))
}
```

## Configuration Profiles

### Available Profiles

| Profile            | Description               | TTL               | Strategy      | Use Case              |
| ------------------ | ------------------------- | ----------------- | ------------- | --------------------- |
| `default`          | Balanced configuration    | 15min/30min/5min  | write-through | General purpose       |
| `high-performance` | Optimized for performance | 30min/60min/10min | cache-aside   | High-traffic services |
| `consistency`      | Optimized for consistency | 5min/10min/2min   | write-through | Critical data         |
| `read-heavy`       | Read-optimized            | 45min/90min/15min | read-through  | Read-heavy workloads  |
| `write-heavy`      | Write-optimized           | 2min/5min/1min    | write-behind  | Write-heavy workloads |

### Creating Profile-Based Decorators

```go
// Create decorator from profile
decorator, err := factory.CreateDecoratorFromProfile(redisConfig, "high-performance")

// Create service-specific decorator
decorator, err := factory.CreateDecoratorForService(redisConfig, "auth-service", "read-heavy")

// Create decorator with custom key prefix
decorator, err := factory.CreateDecoratorWithKeyPrefix(redisConfig, "my-service")
```

## Advanced Configuration

### Custom Configuration

```go
customConfig := &decorators.CacheDecoratorConfig{
    DefaultTTL:       20 * time.Minute,
    EntityTTL:        45 * time.Minute,
    ListTTL:          8 * time.Minute,
    EnableStrategies: true,
    StrategyType:     "cache_aside",
    EnableMetrics:    true,
    KeyPrefix:        "myapp",
}

decorator, err := factory.CreateCustomDecorator(redisConfig, customConfig)
```

### Strategy Types

- **`read_through`**: Cache loads data from source on miss
- **`write_through`**: Cache writes to both cache and source
- **`cache_aside`**: Application manages cache and source separately
- **`write_behind`**: Cache writes to source asynchronously

## API Reference

### CacheDecorator Methods

#### Basic Operations

- `Get(ctx, key, dest)` - Retrieve data from cache
- `Set(ctx, key, value, ttl)` - Store data in cache
- `SetEntity(ctx, key, value)` - Store entity with default entity TTL
- `SetList(ctx, key, value)` - Store list with default list TTL
- `Delete(ctx, key)` - Remove data from cache
- `Exists(ctx, key)` - Check if key exists

#### Advanced Operations

- `GetOrSet(ctx, key, dest, setter, ttl)` - Get or set pattern
- `Remember(ctx, key, dest, fn, ttl)` - Cache function result
- `RememberForever(ctx, key, dest, fn)` - Cache function result forever
- `InvalidatePattern(ctx, pattern)` - Invalidate keys matching pattern
- `ClearAll(ctx)` - Clear all cache entries

#### Management

- `GetCacheStats(ctx)` - Get cache statistics
- `HealthCheck(ctx)` - Perform health check
- `Close()` - Close the decorator

### HTTP Middleware

#### Cache Middleware

- `CacheGet(ttl)` - Cache GET requests with custom TTL
- `CacheEntity(entityType)` - Cache individual entities
- `InvalidateCache(entityType)` - Invalidate cache on mutations

#### Management Handlers

- `CacheStatsHandler()` - Get cache statistics endpoint
- `CacheHealthHandler()` - Cache health check endpoint
- `CacheClearHandler()` - Clear cache endpoint (requires auth)

## Best Practices

### 1. Key Naming Convention

```go
// Use hierarchical keys
"user:123"           // Individual user
"user:list:page=1"   // User list
"user:search:term"   // Search results
```

### 2. TTL Strategy

```go
// Use appropriate TTLs for different data types
decorator.SetEntity(ctx, "user:123", user)     // 30min default
decorator.SetList(ctx, "users:list", users)    // 5min default
decorator.Set(ctx, "config:app", config, 1*time.Hour) // Custom TTL
```

### 3. Error Handling

```go
var user User
err := decorator.Get(ctx, "user:123", &user)
if err != nil {
    if err == cache.ErrCacheMiss {
        // Load from database
        user, err = loadUserFromDB("123")
        if err != nil {
            return err
        }
        // Cache the result
        decorator.SetEntity(ctx, "user:123", user)
    } else {
        // Handle cache error
        return err
    }
}
```

### 4. HTTP Middleware Usage

```go
// Apply caching to read operations
mux.Handle("/api/v1/users/",
    cacheMiddleware.CacheEntity("user")(getUserHandler))

// Apply invalidation to write operations
mux.Handle("/api/v1/users",
    cacheMiddleware.InvalidateCache("user")(createUserHandler))
```

## Monitoring and Debugging

### Cache Statistics

```go
stats, err := decorator.GetCacheStats(ctx)
if err == nil {
    fmt.Printf("Cache stats: %+v\n", stats)
}
```

### Health Check

```go
err := decorator.HealthCheck(ctx)
if err != nil {
    log.Printf("Cache is unhealthy: %v", err)
}
```

### HTTP Management Endpoints

```go
// Add management endpoints
mux.Handle("/cache/stats", cacheMiddleware.CacheStatsHandler())
mux.Handle("/cache/health", cacheMiddleware.CacheHealthHandler())
mux.Handle("/cache/clear", cacheMiddleware.CacheClearHandler())
```

## Examples

See `example_usage.go` for comprehensive examples including:

- Basic cache operations
- Gin integration
- Custom configuration
- Different decorator profiles
- Error handling patterns

## Dependencies

- `backend-core/cache` - Core cache interfaces
- `backend-core/cache/strategies` - Cache strategies
- `backend-core/config` - Configuration management
- `backend-core/logging` - Logging utilities
- `go.uber.org/zap` - Structured logging

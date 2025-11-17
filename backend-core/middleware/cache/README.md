# Cache Middleware Package

This package provides framework-specific cache middleware for different HTTP frameworks.

## Package Structure

```
backend-core/middleware/cache/
├── http/                    # Standard HTTP middleware
│   └── middleware.go        # HTTP cache middleware
├── gin/                     # Gin framework middleware (future)
├── echo/                    # Echo framework middleware (future)
└── README.md               # This file
```

## HTTP Cache Middleware

The HTTP middleware provides standard `net/http` middleware for caching operations.

### Features

- **Cache GET Requests**: Automatically caches GET request responses
- **Entity Caching**: Caches individual entities by type and ID
- **Cache Invalidation**: Automatically invalidates cache on mutations
- **Health Checks**: Provides cache health and statistics endpoints
- **Security**: Includes authentication for cache management operations

### Usage

```go
package main

import (
    "net/http"
    "time"

    "backend-core/cache/decorators"
    "backend-core/middleware/cache/http"
    "backend-core/logging"
)

func main() {
    // Create cache decorator
    cacheDecorator, err := decorators.NewCacheDecorator(redisConfig, logger, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Create HTTP cache middleware
    httpCache := http.NewHTTPCacheMiddleware(cacheDecorator, logger)

    // Create HTTP server
    mux := http.NewServeMux()

    // Add cache middleware to routes
    mux.Handle("/api/users/", httpCache.CacheEntity("user")(userHandler))
    mux.Handle("/api/users", httpCache.CacheGet(10*time.Minute)(userListHandler))

    // Add cache management endpoints
    mux.Handle("/cache/stats", httpCache.CacheStatsHandler())
    mux.Handle("/cache/health", httpCache.CacheHealthHandler())
    mux.Handle("/cache/clear", httpCache.CacheClearHandler())

    http.ListenAndServe(":8080", mux)
}
```

### Middleware Types

#### 1. CacheGet Middleware

Caches GET request responses with a specified TTL.

```go
// Cache GET requests for 10 minutes
middleware.CacheGet(10 * time.Minute)
```

#### 2. CacheEntity Middleware

Caches individual entities by type and ID.

```go
// Cache user entities
middleware.CacheEntity("user")
```

#### 3. InvalidateCache Middleware

Invalidates cache entries when mutations occur.

```go
// Invalidate user cache on mutations
middleware.InvalidateCache("user")
```

### Response Writers

#### CacheResponseWriter

Captures HTTP responses and stores them in cache.

- Only caches successful responses (2xx status codes)
- Asynchronous caching to avoid blocking responses
- Configurable TTL per request

#### InvalidationResponseWriter

Handles cache invalidation after mutations.

- Invalidates cache on successful mutations
- Pattern-based invalidation
- Asynchronous invalidation

### Management Endpoints

#### Cache Statistics

```http
GET /cache/stats
```

Returns cache statistics including hit rates, memory usage, etc.

#### Cache Health Check

```http
GET /cache/health
```

Returns cache health status.

#### Cache Clear

```http
POST /cache/clear
X-Cache-Clear-Token: admin-token
```

Clears all cache entries (requires authentication).

## Framework-Specific Middleware

### Gin Middleware (Future)

```go
// Planned structure
backend-core/middleware/cache/gin/
├── middleware.go           # Gin-specific cache middleware
└── handlers.go            # Gin cache management handlers
```

### Echo Middleware (Future)

```go
// Planned structure
backend-core/middleware/cache/echo/
├── middleware.go           # Echo-specific cache middleware
└── handlers.go            # Echo cache management handlers
```

## Best Practices

### 1. Cache Key Design

- Use consistent key prefixes
- Include version information
- Consider query parameters

```go
// Good cache key design
func buildCacheKey(r *http.Request) string {
    return fmt.Sprintf("%s:%s:%s",
        r.Method,
        r.URL.Path,
        r.URL.RawQuery)
}
```

### 2. TTL Configuration

- Set appropriate TTL values based on data freshness requirements
- Use different TTLs for different data types
- Consider cache invalidation strategies

```go
// Different TTLs for different data types
userCache := middleware.CacheEntity("user")        // 1 hour
configCache := middleware.CacheGet(24 * time.Hour) // 24 hours
sessionCache := middleware.CacheGet(30 * time.Minute) // 30 minutes
```

### 3. Error Handling

- Always handle cache errors gracefully
- Fall back to data source on cache errors
- Log cache errors for monitoring

```go
// Good error handling
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

### 4. Security

- Protect cache management endpoints
- Use authentication tokens
- Validate cache clear requests

```go
// Secure cache clear endpoint
if r.Header.Get("X-Cache-Clear-Token") != "admin-token" {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
```

## Migration Guide

### From cache/middleware/http/

1. **Update imports**:

   ```go
   // Old
   import "backend-core/cache/middleware/http"

   // New
   import "backend-core/middleware/cache/http"
   ```

2. **Update usage**:
   ```go
   // Usage remains the same
   middleware := http.NewHTTPCacheMiddleware(cacheDecorator, logger)
   router.Use(middleware.CacheGet(10*time.Minute))
   ```

## Future Enhancements

### 1. Framework Support

- **Gin Middleware**: Gin-specific cache middleware
- **Echo Middleware**: Echo-specific cache middleware
- **Fiber Middleware**: Fiber-specific cache middleware

### 2. Advanced Features

- **Conditional Caching**: Cache based on request conditions
- **Cache Warming**: Proactive cache population
- **Cache Analytics**: Advanced cache analytics

### 3. Performance

- **Connection Pooling**: Optimized connection management
- **Batch Operations**: Batch cache operations
- **Compression**: Response compression

### 4. Security

- **Rate Limiting**: Cache operation rate limiting
- **Access Control**: Fine-grained access control
- **Audit Logging**: Comprehensive audit logging

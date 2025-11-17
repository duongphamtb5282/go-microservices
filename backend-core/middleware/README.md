# Middleware Package

This package provides a comprehensive middleware system with support for multiple HTTP frameworks, middleware chaining, and various middleware types including caching, logging, security, and validation.

## Package Structure

```
middleware/
├── core/                    # Core middleware interfaces and utilities
│   ├── chain.go            # Middleware chaining
│   ├── interfaces.go       # Middleware interfaces
│   └── registry.go         # Middleware registry
├── cache/                   # Cache middleware
│   ├── http/               # HTTP cache middleware
│   └── README.md           # Cache middleware documentation
├── exception/              # Exception handling middleware
│   ├── error_response.go   # Error response handling
│   ├── exception_handler.go # Exception handler
│   ├── gin_middleware.go   # Gin exception middleware
│   ├── middleware_factory.go # Exception middleware factory
│   └── README.md           # Exception middleware documentation
├── factory/                # Middleware factory
│   ├── builder.go         # Middleware builder
│   ├── config.go          # Middleware configuration
│   └── middleware_factory.go # Middleware factory
├── framework/              # Framework-specific adapters
│   ├── echo/              # Echo framework adapter
│   ├── fiber/             # Fiber framework adapter
│   ├── gin/               # Gin framework adapter
│   └── net_http/          # Net/HTTP adapter
├── examples/               # Middleware examples
│   ├── advanced_usage.go  # Advanced usage examples
│   ├── basic_usage.go     # Basic usage examples
│   └── custom_middleware.go # Custom middleware examples
├── http/                   # HTTP middleware
├── integration/            # Integration middleware
├── logging/                # Logging middleware
├── monitoring/             # Monitoring middleware
├── security/               # Security middleware
├── testing/                # Testing middleware
├── transformation/         # Data transformation middleware
├── validation/             # Validation middleware
├── README.md              # Main middleware documentation
└── USAGE_GUIDE.md         # Usage guide
```

## Core Components

### 1. Middleware Interface (`core/interfaces.go`)

The core middleware interface defines the standard middleware contract:

```go
type Middleware interface {
    // Process request/response
    Process(ctx context.Context, req *Request, next Handler) (*Response, error)
}

type Handler interface {
    Handle(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
    Method  string
    URL     string
    Headers map[string]string
    Body    []byte
    Params  map[string]string
}

type Response struct {
    StatusCode int
    Headers    map[string]string
    Body       []byte
}
```

### 2. Middleware Chain (`core/chain.go`)

The middleware chain provides middleware composition and execution:

```go
type Chain struct {
    middlewares []Middleware
}

func NewChain() *Chain
func (c *Chain) Use(middleware Middleware) *Chain
func (c *Chain) Build(handler Handler) Handler
func (c *Chain) Execute(ctx context.Context, req *Request) (*Response, error)
```

### 3. Middleware Registry (`core/registry.go`)

The middleware registry provides middleware management:

```go
type Registry struct {
    middlewares map[string]Middleware
}

func NewRegistry() *Registry
func (r *Registry) Register(name string, middleware Middleware)
func (r *Registry) Get(name string) (Middleware, error)
func (r *Registry) List() []string
```

## Framework Adapters

### Gin Framework Adapter

```go
// Gin middleware adapter
type GinMiddleware struct {
    middleware Middleware
}

func (g *GinMiddleware) Handler() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Convert Gin context to middleware request
        req := &Request{
            Method:  c.Request.Method,
            URL:     c.Request.URL.String(),
            Headers: extractHeaders(c.Request),
            Body:    getRequestBody(c.Request),
            Params:  extractParams(c),
        }

        // Process middleware
        resp, err := g.middleware.Process(c.Request.Context(), req, nextHandler(c))
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        // Set response
        c.Status(resp.StatusCode)
        for key, value := range resp.Headers {
            c.Header(key, value)
        }
        c.Data(200, "application/json", resp.Body)
    }
}
```

### Echo Framework Adapter

```go
// Echo middleware adapter
type EchoMiddleware struct {
    middleware Middleware
}

func (e *EchoMiddleware) Handler() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Convert Echo context to middleware request
            req := &Request{
                Method:  c.Request().Method,
                URL:     c.Request().URL.String(),
                Headers: extractHeaders(c.Request()),
                Body:    getRequestBody(c.Request()),
                Params:  extractParams(c),
            }

            // Process middleware
            resp, err := e.middleware.Process(c.Request().Context(), req, nextHandler(c, next))
            if err != nil {
                return err
            }

            // Set response
            c.Response().Status = resp.StatusCode
            for key, value := range resp.Headers {
                c.Response().Header().Set(key, value)
            }
            return c.JSONBlob(resp.StatusCode, resp.Body)
        }
    }
}
```

### Fiber Framework Adapter

```go
// Fiber middleware adapter
type FiberMiddleware struct {
    middleware Middleware
}

func (f *FiberMiddleware) Handler() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Convert Fiber context to middleware request
        req := &Request{
            Method:  c.Method(),
            URL:     c.OriginalURL(),
            Headers: extractHeaders(c),
            Body:    c.Body(),
            Params:  extractParams(c),
        }

        // Process middleware
        resp, err := f.middleware.Process(c.Context(), req, nextHandler(c))
        if err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }

        // Set response
        c.Status(resp.StatusCode)
        for key, value := range resp.Headers {
            c.Set(key, value)
        }
        return c.Send(resp.Body)
    }
}
```

## Built-in Middleware

### 1. Logging Middleware

```go
// Create logging middleware
loggingMiddleware := NewLoggingMiddleware(&LoggingConfig{
    Level:  "info",
    Format: "json",
    Fields: []string{"method", "path", "status", "duration"},
})

// Use with middleware chain
chain := NewChain().
    Use(loggingMiddleware).
    Use(otherMiddleware)

// Or use directly with framework
gin.Use(NewGinLoggingMiddleware(loggingMiddleware))
```

### 2. Cache Middleware

```go
// Create cache middleware
cacheMiddleware := NewCacheMiddleware(&CacheConfig{
    TTL:        5 * time.Minute,
    KeyPrefix:  "api:",
    Strategies: []string{"read-through", "write-through"},
})

// Use with middleware chain
chain := NewChain().
    Use(cacheMiddleware).
    Use(handler)

// Or use directly with framework
gin.Use(NewGinCacheMiddleware(cacheMiddleware))
```

### 3. Security Middleware

```go
// Create security middleware
securityMiddleware := NewSecurityMiddleware(&SecurityConfig{
    CORS: &CORSConfig{
        AllowedOrigins: []string{"https://example.com"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type", "Authorization"},
    },
    RateLimit: &RateLimitConfig{
        Requests: 100,
        Window:   1 * time.Minute,
    },
    Helmet: &HelmetConfig{
        XSSProtection: true,
        NoSniff:       true,
        FrameOptions:  "DENY",
    },
})

// Use with middleware chain
chain := NewChain().
    Use(securityMiddleware).
    Use(handler)
```

### 4. Validation Middleware

```go
// Create validation middleware
validationMiddleware := NewValidationMiddleware(&ValidationConfig{
    SchemaPath: "schemas/",
    StrictMode: true,
    CustomValidators: map[string]Validator{
        "email": emailValidator,
        "phone": phoneValidator,
    },
})

// Use with middleware chain
chain := NewChain().
    Use(validationMiddleware).
    Use(handler)
```

### 5. Monitoring Middleware

```go
// Create monitoring middleware
monitoringMiddleware := NewMonitoringMiddleware(&MonitoringConfig{
    Metrics: &MetricsConfig{
        Enabled: true,
        Port:    9090,
    },
    Tracing: &TracingConfig{
        Enabled: true,
        Service: "my-service",
    },
    Health: &HealthConfig{
        Enabled: true,
        Path:    "/health",
    },
})

// Use with middleware chain
chain := NewChain().
    Use(monitoringMiddleware).
    Use(handler)
```

## Custom Middleware

### Creating Custom Middleware

```go
// Custom middleware implementation
type CustomMiddleware struct {
    config *CustomConfig
}

func NewCustomMiddleware(config *CustomConfig) *CustomMiddleware {
    return &CustomMiddleware{config: config}
}

func (m *CustomMiddleware) Process(ctx context.Context, req *Request, next Handler) (*Response, error) {
    // Pre-processing
    start := time.Now()

    // Add custom headers
    req.Headers["X-Custom-Header"] = "custom-value"

    // Process request
    resp, err := next.Handle(ctx, req)
    if err != nil {
        return nil, err
    }

    // Post-processing
    duration := time.Since(start)
    resp.Headers["X-Processing-Time"] = duration.String()

    return resp, nil
}
```

### Middleware with Dependencies

```go
// Middleware with dependencies
type DatabaseMiddleware struct {
    db     Database
    logger Logger
    cache  Cache
}

func NewDatabaseMiddleware(db Database, logger Logger, cache Cache) *DatabaseMiddleware {
    return &DatabaseMiddleware{
        db:     db,
        logger: logger,
        cache:  cache,
    }
}

func (m *DatabaseMiddleware) Process(ctx context.Context, req *Request, next Handler) (*Response, error) {
    // Use dependencies
    m.logger.Info("Processing request", String("path", req.URL))

    // Check cache
    cached, err := m.cache.Get(ctx, req.URL)
    if err == nil {
        return &Response{
            StatusCode: 200,
            Body:       cached,
        }, nil
    }

    // Process request
    resp, err := next.Handle(ctx, req)
    if err != nil {
        return nil, err
    }

    // Cache response
    m.cache.Set(ctx, req.URL, resp.Body, 5*time.Minute)

    return resp, nil
}
```

## Middleware Factory

### Using Middleware Factory

```go
// Create middleware factory
factory := NewMiddlewareFactory()

// Register middleware
factory.Register("logging", NewLoggingMiddleware)
factory.Register("cache", NewCacheMiddleware)
factory.Register("security", NewSecurityMiddleware)

// Create middleware chain
chain := factory.CreateChain("logging", "cache", "security")

// Build handler
handler := chain.Build(finalHandler)
```

### Configuration-Based Middleware

```go
// Middleware configuration
config := &MiddlewareConfig{
    Middlewares: []MiddlewareConfig{
        {
            Name: "logging",
            Config: map[string]interface{}{
                "level": "info",
                "format": "json",
            },
        },
        {
            Name: "cache",
            Config: map[string]interface{}{
                "ttl": "5m",
                "strategy": "read-through",
            },
        },
    },
}

// Create middleware from configuration
chain, err := factory.CreateChainFromConfig(config)
if err != nil {
    log.Fatal("Failed to create middleware chain", err)
}
```

## Middleware Testing

### Unit Testing

```go
func TestCustomMiddleware(t *testing.T) {
    // Create test middleware
    middleware := NewCustomMiddleware(&CustomConfig{
        Timeout: 5 * time.Second,
    })

    // Create test request
    req := &Request{
        Method: "GET",
        URL:    "/test",
        Headers: map[string]string{},
        Body:   []byte{},
    }

    // Create test handler
    handler := func(ctx context.Context, req *Request) (*Response, error) {
        return &Response{
            StatusCode: 200,
            Body:       []byte("test response"),
        }, nil
    }

    // Test middleware
    resp, err := middleware.Process(context.Background(), req, handler)
    require.NoError(t, err)
    require.Equal(t, 200, resp.StatusCode)
    require.Equal(t, "test response", string(resp.Body))
}
```

### Integration Testing

```go
func TestMiddlewareChain(t *testing.T) {
    // Create middleware chain
    chain := NewChain().
        Use(NewLoggingMiddleware(&LoggingConfig{})).
        Use(NewCacheMiddleware(&CacheConfig{})).
        Use(NewSecurityMiddleware(&SecurityConfig{}))

    // Create test server
    server := httptest.NewServer(chain.Build(testHandler))
    defer server.Close()

    // Test request
    resp, err := http.Get(server.URL + "/test")
    require.NoError(t, err)
    require.Equal(t, 200, resp.StatusCode)
}
```

## Best Practices

### 1. Middleware Order

- **Security** middleware should be first
- **Logging** middleware should be early
- **Cache** middleware should be before expensive operations
- **Validation** middleware should be before business logic
- **Error handling** middleware should be last

### 2. Error Handling

- Always handle errors gracefully
- Provide meaningful error messages
- Log errors with context
- Don't expose internal errors to clients

### 3. Performance

- Keep middleware lightweight
- Avoid expensive operations in middleware
- Use caching for expensive computations
- Monitor middleware performance

### 4. Testing

- Test middleware in isolation
- Test middleware chains
- Mock dependencies
- Test error scenarios

### 5. Configuration

- Use configuration for middleware settings
- Provide sensible defaults
- Validate configuration
- Document configuration options

## Examples

### Complete Application Setup

```go
func main() {
    // Create middleware factory
    factory := NewMiddlewareFactory()

    // Register middleware
    factory.Register("logging", NewLoggingMiddleware)
    factory.Register("cache", NewCacheMiddleware)
    factory.Register("security", NewSecurityMiddleware)
    factory.Register("validation", NewValidationMiddleware)

    // Create middleware chain
    chain := factory.CreateChain("security", "logging", "cache", "validation")

    // Create Gin router
    router := gin.New()

    // Apply middleware
    router.Use(NewGinMiddleware(chain))

    // Add routes
    router.GET("/users", getUsers)
    router.POST("/users", createUser)
    router.PUT("/users/:id", updateUser)
    router.DELETE("/users/:id", deleteUser)

    // Start server
    router.Run(":8080")
}
```

### Custom Middleware Example

```go
// Request ID middleware
type RequestIDMiddleware struct {
    generator IDGenerator
}

func NewRequestIDMiddleware() *RequestIDMiddleware {
    return &RequestIDMiddleware{
        generator: NewUUIDGenerator(),
    }
}

func (m *RequestIDMiddleware) Process(ctx context.Context, req *Request, next Handler) (*Response, error) {
    // Generate request ID
    requestID := m.generator.Generate()

    // Add to context
    ctx = context.WithValue(ctx, "request_id", requestID)

    // Add to request headers
    req.Headers["X-Request-ID"] = requestID

    // Process request
    resp, err := next.Handle(ctx, req)
    if err != nil {
        return nil, err
    }

    // Add to response headers
    resp.Headers["X-Request-ID"] = requestID

    return resp, nil
}
```

## Migration Guide

When upgrading middleware:

1. **Check breaking changes** in the changelog
2. **Update middleware configuration** if needed
3. **Test middleware chains** in all environments
4. **Update framework adapters** if using custom adapters
5. **Monitor performance** after upgrade

## Future Enhancements

- **GraphQL middleware** - Support for GraphQL middleware
- **gRPC middleware** - Support for gRPC middleware
- **WebSocket middleware** - Support for WebSocket middleware
- **Middleware composition** - Advanced middleware composition patterns
- **Dynamic middleware** - Runtime middleware configuration

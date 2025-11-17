# Exception Handling Middleware

A comprehensive exception handling middleware for Go microservices that provides centralized error handling, logging, and consistent error responses.

## 🎯 Features

- **Centralized Error Handling**: Single point for all exception handling
- **Standardized Error Responses**: Consistent error format across all services
- **Gin Framework Integration**: Ready-to-use Gin middleware
- **Panic Recovery**: Automatic panic recovery with detailed logging
- **Request Context**: Automatic request ID generation and context tracking
- **Stack Trace Support**: Optional stack trace inclusion for debugging
- **Environment-Specific Configuration**: Different settings for dev/prod/test
- **Comprehensive Logging**: Structured logging with request context

## 📁 Structure

```
backend-core/middleware/exception/
├── error_response.go        # Error response structures
├── exception_handler.go      # Core exception handling logic
├── gin_middleware.go        # Gin framework integration
├── middleware_factory.go   # Factory for creating middleware
└── README.md               # This file
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    "backend-core/middleware/exception"
    "go.uber.org/zap"
)

func main() {
    r := gin.New()

    // Create logger
    logger, _ := zap.NewProduction()

    // Create exception middleware
    exceptionMiddleware := exception.CreateProductionMiddleware(
        "auth-service",
        "1.0.0",
        logger,
    )

    // Apply middleware
    r.Use(exceptionMiddleware...)

    // Your routes
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    r.Run(":8080")
}
```

### Custom Configuration

```go
// Create custom configuration
config := &exception.MiddlewareConfig{
    ServiceName:   "auth-service",
    Version:       "1.0.0",
    IncludeStack:  true,  // Include stack traces
    HideInternal:  false, // Show internal error details
    Logger:        logger,
}

// Create factory
factory := exception.NewMiddlewareFactory(config)

// Create middleware
middleware := factory.CreateFullMiddlewareStack()

// Apply to router
r.Use(middleware...)
```

## 🔧 Configuration Options

### MiddlewareConfig

```go
type MiddlewareConfig struct {
    ServiceName   string     // Service name for logging
    Version       string     // Service version
    IncludeStack  bool       // Include stack traces in responses
    HideInternal  bool       // Hide internal error details
    Logger        *zap.Logger // Logger instance
}
```

### Environment-Specific Configurations

#### Production

```go
middleware := exception.CreateProductionMiddleware(
    "auth-service",
    "1.0.0",
    logger,
)
// - IncludeStack: false (security)
// - HideInternal: true (security)
```

#### Development

```go
middleware := exception.CreateDevelopmentMiddleware(
    "auth-service",
    "1.0.0",
    logger,
)
// - IncludeStack: true (debugging)
// - HideInternal: false (debugging)
```

#### Testing

```go
middleware := exception.CreateTestMiddleware(
    "auth-service",
    "1.0.0",
)
// - IncludeStack: true (debugging)
// - HideInternal: false (debugging)
```

## 📝 Error Response Format

### Standard Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "field": "email",
      "value": "invalid-email",
      "errors": [
        {
          "field": "email",
          "value": "invalid-email",
          "message": "Invalid email format"
        }
      ],
      "count": 1
    },
    "type": "ValidationError"
  },
  "request_id": "20231201120000-abc12345",
  "timestamp": "2023-12-01T12:00:00Z",
  "path": "/api/v1/users",
  "method": "POST"
}
```

### Error Types

- **ValidationError**: Input validation failures
- **NotFoundError**: Resource not found
- **UnauthorizedError**: Authentication failures
- **ForbiddenError**: Authorization failures
- **InternalError**: Internal server errors
- **ExternalError**: External service failures
- **DomainError**: Domain-specific errors

## 🛠️ Advanced Usage

### Custom Exception Handler

```go
type CustomExceptionHandler struct {
    *exception.DefaultExceptionHandler
    // Add custom fields
}

func (h *CustomExceptionHandler) HandleException(ctx context.Context, err error, requestInfo exception.RequestInfo) exception.ErrorResponse {
    // Custom logic before calling default handler
    if customErr, ok := err.(*CustomError); ok {
        return h.handleCustomError(customErr, requestInfo)
    }

    // Fall back to default handler
    return h.DefaultExceptionHandler.HandleException(ctx, err, requestInfo)
}

func (h *CustomExceptionHandler) handleCustomError(err *CustomError, requestInfo exception.RequestInfo) exception.ErrorResponse {
    // Custom error handling logic
    return exception.NewErrorResponse(
        "CUSTOM_ERROR",
        err.Message,
        "CustomError",
        err.Details,
    ).WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)
}
```

### Manual Error Handling

```go
func createUser(c *gin.Context) {
    // Your business logic
    user, err := userService.CreateUser(userData)
    if err != nil {
        // Use helper function
        exception.GinErrorResponse(c, err, http.StatusBadRequest)
        return
    }

    // Success response
    exception.GinSuccessResponse(c, user, http.StatusCreated)
}
```

### Request Context Integration

```go
func authMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract user ID from JWT
        userID := extractUserIDFromJWT(c)

        // Set in context for exception handling
        c.Set("user_id", userID)

        c.Next()
    }
}
```

## 📊 Logging

### Structured Logging

The middleware automatically logs:

- **Request Information**: Method, path, IP, user agent
- **Error Details**: Error type, message, stack trace
- **Context Information**: Request ID, user ID, service info
- **Performance Metrics**: Request duration, status codes

### Log Format

```json
{
  "level": "error",
  "ts": 1701432000.123,
  "msg": "Exception occurred",
  "error": "validation error for field 'email': Invalid email format",
  "request_id": "20231201120000-abc12345",
  "method": "POST",
  "path": "/api/v1/users",
  "user_agent": "Mozilla/5.0...",
  "ip": "192.168.1.100",
  "user_id": "user-123",
  "service": "auth-service",
  "version": "1.0.0",
  "stack": "goroutine 1 [running]:\n..."
}
```

## 🔒 Security Considerations

### Production Settings

- **HideInternal**: Set to `true` to hide internal error details
- **IncludeStack**: Set to `false` to prevent stack trace exposure
- **Sensitive Data**: Ensure sensitive data is not logged

### Error Sanitization

```go
// Custom sanitization
func sanitizeError(err error) error {
    // Remove sensitive information
    if strings.Contains(err.Error(), "password") {
        return errors.New("Authentication failed")
    }
    return err
}
```

## 🧪 Testing

### Unit Testing

```go
func TestExceptionHandler(t *testing.T) {
    logger, _ := zap.NewDevelopment()
    handler := exception.NewDefaultExceptionHandler(logger, "test-service", "1.0.0")

    requestInfo := exception.RequestInfo{
        RequestID: "test-123",
        Method:    "POST",
        Path:      "/api/v1/users",
        IP:        "127.0.0.1",
    }

    err := errors.NewValidationError("email", "invalid", "Invalid email format")
    response := handler.HandleException(context.Background(), err, requestInfo)

    assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
    assert.Equal(t, "Invalid email format", response.Error.Message)
}
```

### Integration Testing

```go
func TestGinMiddleware(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()

    logger, _ := zap.NewDevelopment()
    middleware := exception.CreateTestMiddleware("test-service", "1.0.0")
    r.Use(middleware...)

    r.GET("/test", func(c *gin.Context) {
        panic("test panic")
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/test", nil)
    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusInternalServerError, w.Code)

    var response exception.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "InternalError", response.Error.Type)
}
```

## 🚀 Performance

### Optimizations

- **Minimal Allocations**: Efficient memory usage
- **Fast Error Classification**: Quick error type detection
- **Async Logging**: Non-blocking log operations
- **Request Context Caching**: Cached request information

### Benchmarks

```go
func BenchmarkExceptionHandler(b *testing.B) {
    logger, _ := zap.NewProduction()
    handler := exception.NewDefaultExceptionHandler(logger, "bench-service", "1.0.0")

    requestInfo := exception.RequestInfo{
        RequestID: "bench-123",
        Method:    "GET",
        Path:      "/api/v1/test",
    }

    err := errors.New("test error")

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        handler.HandleException(context.Background(), err, requestInfo)
    }
}
```

## 🔄 Integration with Other Services

### Auth Service Example

```go
// auth-service/main.go
func main() {
    r := gin.New()

    // Load configuration
    config := loadConfig()
    logger := setupLogger(config)

    // Create exception middleware
    exceptionMiddleware := exception.CreateProductionMiddleware(
        "auth-service",
        config.Version,
        logger,
    )

    // Apply middleware
    r.Use(exceptionMiddleware...)

    // Add other middleware
    r.Use(corsMiddleware())
    r.Use(authMiddleware())

    // Setup routes
    setupRoutes(r)

    r.Run(":8080")
}
```

### User Service Example

```go
// user-service/main.go
func main() {
    r := gin.New()

    // Development middleware for local development
    if os.Getenv("APP_ENV") == "development" {
        exceptionMiddleware := exception.CreateDevelopmentMiddleware(
            "user-service",
            "1.0.0",
            logger,
        )
        r.Use(exceptionMiddleware...)
    } else {
        // Production middleware
        exceptionMiddleware := exception.CreateProductionMiddleware(
            "user-service",
            "1.0.0",
            logger,
        )
        r.Use(exceptionMiddleware...)
    }

    r.Run(":8081")
}
```

## 📚 Best Practices

### 1. Error Classification

```go
// Use specific error types
return errors.NewValidationError("email", email, "Invalid email format")
return errors.NewDomainError("USER_NOT_FOUND", "User not found")
```

### 2. Request Context

```go
// Always set request context
c.Set("request_id", generateRequestID())
c.Set("user_id", userID)
```

### 3. Logging Levels

```go
// Use appropriate log levels
logger.Error("Critical error", zap.Error(err))
logger.Warn("Warning", zap.Error(err))
logger.Info("Information", zap.String("message", "User created"))
```

### 4. Error Responses

```go
// Use helper functions for consistent responses
exception.GinErrorResponse(c, err, http.StatusBadRequest)
exception.GinSuccessResponse(c, data, http.StatusOK)
```

## 🐛 Troubleshooting

### Common Issues

1. **Stack Traces Not Showing**

   - Check `IncludeStack` configuration
   - Verify logger configuration

2. **Internal Errors Exposed**

   - Set `HideInternal` to `true`
   - Check error message sanitization

3. **Request ID Missing**
   - Ensure middleware is applied before routes
   - Check request ID generation logic

### Debug Mode

```go
// Enable debug logging
logger, _ := zap.NewDevelopment()
config := &exception.MiddlewareConfig{
    ServiceName:   "debug-service",
    Version:       "1.0.0",
    IncludeStack:  true,
    HideInternal:  false,
    Logger:        logger,
}
```

## 📄 License

This middleware is part of the backend-core package and follows the same license terms.

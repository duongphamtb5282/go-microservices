# Logging Package

This package provides a comprehensive logging solution with structured logging, sensitive data masking, and multiple output formats using Zap as the underlying logger.

## Package Structure

```
logging/
├── logger.go              # Core logger interface and implementation
└── masking/               # Sensitive data masking utilities
    ├── auditor.go         # Audit logging
    ├── cache.go           # Cache masking
    ├── encoder.go         # Log encoding
    ├── factory.go         # Masking factory
    ├── masker.go          # Data masking
    ├── metrics.go         # Logging metrics
    ├── strategies.go      # Masking strategies
    └── types.go           # Masking types
```

## Core Components

### 1. Logger Interface (`logger.go`)

The core logger interface provides structured logging capabilities:

```go
type Logger interface {
    // Basic logging methods
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
    Panic(msg string, fields ...Field)

    // Context-aware logging
    WithContext(ctx context.Context) Logger
    WithFields(fields ...Field) Logger
    WithError(err error) Logger

    // Structured logging
    Log(level Level, msg string, fields ...Field)

    // Flush and close
    Sync() error
    Close() error
}
```

### 2. Field Types

```go
// Field types for structured logging
type Field interface{}

// Common field constructors
func String(key, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Float64(key string, val float64) Field
func Bool(key string, val bool) Field
func Time(key string, val time.Time) Field
func Duration(key string, val time.Duration) Field
func Error(err error) Field
func Any(key string, val interface{}) Field
```

### 3. Log Levels

```go
type Level int

const (
    DebugLevel Level = iota
    InfoLevel
    WarnLevel
    ErrorLevel
    FatalLevel
    PanicLevel
)
```

## Basic Usage

### Creating a Logger

```go
// Create a new logger
logger, err := NewLogger(&LoggingConfig{
    Level:  "info",
    Format: "json",
    Output: "stdout",
})

if err != nil {
    log.Fatal("Failed to create logger", err)
}

// Use the logger
logger.Info("Application started",
    String("version", "1.0.0"),
    String("environment", "production"),
)
```

### Structured Logging

```go
// Log with structured fields
logger.Info("User login",
    String("user_id", "123"),
    String("email", "user@example.com"),
    Time("timestamp", time.Now()),
    Duration("response_time", 150*time.Millisecond),
)

// Log with error
err := someOperation()
if err != nil {
    logger.Error("Operation failed",
        Error(err),
        String("operation", "user_creation"),
        String("user_id", "123"),
    )
}
```

### Context-Aware Logging

```go
// Create logger with context
ctx := context.WithValue(context.Background(), "request_id", "req-123")
loggerWithCtx := logger.WithContext(ctx)

// All logs will include request_id
loggerWithCtx.Info("Processing request",
    String("endpoint", "/api/users"),
    String("method", "POST"),
)

// Create logger with fields
loggerWithFields := logger.WithFields(
    String("service", "user-service"),
    String("version", "1.0.0"),
)

loggerWithFields.Info("Service started")
```

## Configuration

### Logging Configuration

```go
type LoggingConfig struct {
    Level      string `yaml:"level" json:"level"`           // debug, info, warn, error
    Format     string `yaml:"format" json:"format"`         // json, console
    Output     string `yaml:"output" json:"output"`          // stdout, stderr, file
    File       string `yaml:"file" json:"file"`              // Log file path
    MaxSize    int    `yaml:"max_size" json:"max_size"`      // Max file size in MB
    MaxBackups int    `yaml:"max_backups" json:"max_backups"` // Max backup files
    MaxAge     int    `yaml:"max_age" json:"max_age"`        // Max age in days
    Compress   bool   `yaml:"compress" json:"compress"`      // Compress old files
}
```

### Environment-Specific Configuration

```go
// Development configuration
devConfig := &LoggingConfig{
    Level:  "debug",
    Format: "console",
    Output: "stdout",
}

// Production configuration
prodConfig := &LoggingConfig{
    Level:      "info",
    Format:     "json",
    Output:     "file",
    File:       "/var/log/app.log",
    MaxSize:    100,
    MaxBackups: 5,
    MaxAge:     30,
    Compress:   true,
}
```

## Sensitive Data Masking

### Basic Masking

```go
// Create logger with masking
logger, err := NewLoggerWithMasking(&LoggingConfig{
    Level: "info",
    Format: "json",
}, &MaskingConfig{
    MaskedFields: []string{"password", "secret", "token", "key"},
    MaskChar:     "*",
    MaskLength:    4,
})

// Sensitive data will be automatically masked
logger.Info("User data",
    String("username", "john_doe"),
    String("password", "secret123"), // Will be masked as "****"
    String("email", "john@example.com"),
)
```

### Custom Masking Strategies

```go
// Implement custom masking strategy
type CustomMaskingStrategy struct {
    patterns []string
}

func (s *CustomMaskingStrategy) Mask(key string, value interface{}) (interface{}, bool) {
    for _, pattern := range s.patterns {
        if strings.Contains(key, pattern) {
            return "***MASKED***", true
        }
    }
    return value, false
}

// Use custom masking strategy
maskingConfig := &MaskingConfig{
    Strategy: &CustomMaskingStrategy{
        patterns: []string{"password", "secret", "token"},
    },
}

logger, err := NewLoggerWithMasking(config, maskingConfig)
```

### Field-Level Masking

```go
// Mask specific fields
logger.Info("User registration",
    String("username", "john_doe"),
    MaskedString("password", "secret123"), // Explicitly masked
    String("email", "john@example.com"),
)

// Mask with custom mask
logger.Info("API request",
    String("endpoint", "/api/users"),
    CustomMasked("api_key", "sk-1234567890", "***"),
)
```

## Audit Logging

### Audit Logger

```go
// Create audit logger
auditLogger := NewAuditLogger(&AuditConfig{
    Level:     "info",
    Format:    "json",
    Output:    "file",
    File:      "/var/log/audit.log",
    IncludeIP: true,
    IncludeUA: true,
})

// Log audit events
auditLogger.LogUserAction("user_login",
    String("user_id", "123"),
    String("action", "login"),
    String("ip_address", "192.168.1.1"),
    String("user_agent", "Mozilla/5.0..."),
)
```

### Security Events

```go
// Log security events
auditLogger.LogSecurityEvent("failed_login",
    String("user_id", "123"),
    String("email", "user@example.com"),
    String("ip_address", "192.168.1.1"),
    String("reason", "invalid_password"),
)

auditLogger.LogSecurityEvent("suspicious_activity",
    String("user_id", "123"),
    String("activity", "multiple_failed_logins"),
    String("ip_address", "192.168.1.1"),
    Int("attempts", 5),
)
```

## Performance Monitoring

### Logging Metrics

```go
// Create logger with metrics
logger, err := NewLoggerWithMetrics(&LoggingConfig{
    Level: "info",
    Format: "json",
}, &MetricsConfig{
    EnableMetrics: true,
    MetricsPort:   9090,
})

// Metrics will be available at http://localhost:9090/metrics
```

### Custom Metrics

```go
// Implement custom metrics
type CustomMetrics struct {
    logCount    int64
    errorCount  int64
    warnCount   int64
}

func (m *CustomMetrics) IncrementLogCount() {
    atomic.AddInt64(&m.logCount, 1)
}

func (m *CustomMetrics) IncrementErrorCount() {
    atomic.AddInt64(&m.errorCount, 1)
}

// Use custom metrics
metrics := &CustomMetrics{}
logger, err := NewLoggerWithCustomMetrics(config, metrics)
```

## Log Rotation

### Automatic Rotation

```go
// Configure log rotation
config := &LoggingConfig{
    Level:      "info",
    Format:     "json",
    Output:     "file",
    File:       "/var/log/app.log",
    MaxSize:    100,    // 100MB
    MaxBackups: 5,      // Keep 5 backup files
    MaxAge:     30,     // Keep logs for 30 days
    Compress:   true,   // Compress old logs
}

logger, err := NewLogger(config)
```

### Manual Rotation

```go
// Rotate logs manually
err := logger.Rotate()
if err != nil {
    log.Error("Failed to rotate logs", err)
}
```

## Error Handling

### Error Logging

```go
// Log errors with context
func processUser(userID string) error {
    user, err := getUser(userID)
    if err != nil {
        logger.Error("Failed to get user",
            Error(err),
            String("user_id", userID),
            String("operation", "get_user"),
        )
        return err
    }

    err = validateUser(user)
    if err != nil {
        logger.Error("User validation failed",
            Error(err),
            String("user_id", userID),
            Any("user", user),
        )
        return err
    }

    return nil
}
```

### Panic Recovery

```go
// Recover from panics and log them
defer func() {
    if r := recover(); r != nil {
        logger.Panic("Application panic",
            Any("panic", r),
            String("stack", string(debug.Stack())),
        )
    }
}()
```

## Testing

### Mock Logger

```go
// Create mock logger for testing
type MockLogger struct {
    logs []LogEntry
}

func (m *MockLogger) Info(msg string, fields ...Field) {
    m.logs = append(m.logs, LogEntry{
        Level: InfoLevel,
        Message: msg,
        Fields: fields,
    })
}

// Use in tests
func TestUserService(t *testing.T) {
    mockLogger := &MockLogger{}
    service := NewUserService(mockLogger)

    // Test service
    err := service.CreateUser("john", "john@example.com")
    assert.NoError(t, err)

    // Verify logging
    assert.Len(t, mockLogger.logs, 1)
    assert.Equal(t, "User created", mockLogger.logs[0].Message)
}
```

### Test Logging

```go
// Test logging behavior
func TestLogging(t *testing.T) {
    // Create test logger
    logger, err := NewLogger(&LoggingConfig{
        Level: "debug",
        Format: "console",
        Output: "stdout",
    })
    require.NoError(t, err)

    // Test different log levels
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")

    // Test structured logging
    logger.Info("Structured log",
        String("key", "value"),
        Int("number", 42),
    )
}
```

## Best Practices

### 1. Log Levels

- **Debug**: Detailed information for debugging
- **Info**: General information about application flow
- **Warn**: Warning messages for potential issues
- **Error**: Error messages for recoverable errors
- **Fatal**: Critical errors that cause application termination

### 2. Structured Logging

- Use consistent field names
- Include relevant context
- Avoid logging sensitive data
- Use appropriate data types

### 3. Performance

- Use appropriate log levels
- Avoid expensive operations in log statements
- Use structured logging for better performance
- Consider log volume in production

### 4. Security

- Mask sensitive data
- Avoid logging passwords or tokens
- Use audit logging for security events
- Implement log access controls

### 5. Monitoring

- Monitor log volume
- Set up log-based alerts
- Use log aggregation tools
- Implement log retention policies

## Examples

### Complete Application Logging

```go
// Application setup
func main() {
    // Create logger
    logger, err := NewLogger(&LoggingConfig{
        Level:  "info",
        Format: "json",
        Output: "file",
        File:   "/var/log/app.log",
    })
    if err != nil {
        log.Fatal("Failed to create logger", err)
    }
    defer logger.Close()

    // Log application start
    logger.Info("Application starting",
        String("version", "1.0.0"),
        String("environment", "production"),
    )

    // Start HTTP server
    server := &http.Server{
        Addr: ":8080",
        Handler: loggingMiddleware(logger, http.DefaultServeMux),
    }

    logger.Info("Server starting",
        String("addr", server.Addr),
    )

    if err := server.ListenAndServe(); err != nil {
        logger.Fatal("Server failed to start", Error(err))
    }
}

// HTTP middleware
func loggingMiddleware(logger Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Log request
        logger.Info("Request started",
            String("method", r.Method),
            String("path", r.URL.Path),
            String("remote_addr", r.RemoteAddr),
        )

        // Process request
        next.ServeHTTP(w, r)

        // Log response
        logger.Info("Request completed",
            String("method", r.Method),
            String("path", r.URL.Path),
            Duration("duration", time.Since(start)),
        )
    })
}
```

## Migration Guide

When upgrading logging:

1. **Check breaking changes** in the changelog
2. **Update configuration** if needed
3. **Test log output** in all environments
4. **Update log parsing** if using external tools
5. **Monitor performance** after upgrade

## Future Enhancements

- **Distributed tracing** - Integration with distributed tracing systems
- **Log sampling** - Intelligent log sampling for high-volume applications
- **Log streaming** - Real-time log streaming capabilities
- **Advanced masking** - More sophisticated data masking strategies
- **Log analytics** - Built-in log analytics and insights

package examples

import (
	"backend-core/logging"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ExampleGrafanaLogging demonstrates how to use Grafana-compatible logging
func ExampleGrafanaLogging() {
	// Configure Grafana logging
	config := &logging.GrafanaLogConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Tags: map[string]string{
			"team":    "backend",
			"project": "auth-service",
			"region":  "us-west-2",
		},
		EnableTraceID: true,
		EnableSpanID:  true,
		EnableMetrics: true,
	}

	// Create Grafana logger
	logger, err := logging.NewGrafanaLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Example 1: Basic logging with service context
	logger.Info("Service started",
		zap.String("service", "auth-service"),
		zap.String("version", "1.0.0"),
		zap.String("environment", "production"),
	)

	// Example 2: Request logging with correlation IDs
	requestID := "req-12345"
	userID := "user-67890"
	traceID := "trace-abc123"
	spanID := "span-def456"

	logger.Info("Processing request",
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("trace_id", traceID),
		zap.String("span_id", spanID),
		zap.String("method", "POST"),
		zap.String("url", "/api/v1/auth/login"),
		zap.String("ip_address", "192.168.1.100"),
		zap.String("user_agent", "Mozilla/5.0..."),
	)

	// Example 3: Performance logging with metrics
	startTime := time.Now()
	// Simulate some work
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(startTime)

	logger.Info("Request completed",
		zap.String("request_id", requestID),
		zap.Int("status_code", 200),
		zap.Int64("duration_ms", duration.Milliseconds()),
		zap.Float64("m_response_time", float64(duration.Milliseconds())),
		zap.Float64("m_throughput", 1.0),
	)

	// Example 4: Error logging with stack trace
	logger.Error("Database connection failed",
		zap.String("request_id", requestID),
		zap.String("error", "connection timeout"),
		zap.String("database", "postgresql"),
		zap.String("host", "db.example.com"),
		zap.Int("port", 5432),
		zap.Int("retry_count", 3),
		zap.Float64("m_error_rate", 0.1),
	)

	// Example 5: Security logging
	logger.Warn("Suspicious activity detected",
		zap.String("request_id", requestID),
		zap.String("ip_address", "192.168.1.100"),
		zap.String("user_agent", "curl/7.68.0"),
		zap.String("security_event", "multiple_failed_logins"),
		zap.String("threat_level", "medium"),
		zap.String("blocked_reason", "rate_limit_exceeded"),
		zap.Float64("m_security_events", 1.0),
	)

	// Example 6: Business logic logging
	logger.Info("User registered successfully",
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("username", "john_doe"),
		zap.String("email", "john@example.com"),
		zap.String("registration_method", "email"),
		zap.Float64("m_user_registrations", 1.0),
	)

	// Example 7: Cache logging
	logger.Info("Cache operation",
		zap.String("request_id", requestID),
		zap.String("operation", "get"),
		zap.String("cache_key", "user:12345"),
		zap.Bool("cache_hit", true),
		zap.Int64("duration_ms", 5),
		zap.Float64("m_cache_hit_rate", 0.95),
	)

	// Example 8: Database logging
	logger.Info("Database query executed",
		zap.String("request_id", requestID),
		zap.String("query", "SELECT * FROM users WHERE id = ?"),
		zap.String("database", "postgresql"),
		zap.Int64("duration_ms", 25),
		zap.Int("rows_affected", 1),
		zap.Float64("m_db_query_time", 25.0),
	)

	// Example 9: External service logging
	logger.Info("External API call",
		zap.String("request_id", requestID),
		zap.String("service", "notification-service"),
		zap.String("endpoint", "POST /api/v1/notifications"),
		zap.Int("status_code", 201),
		zap.Int64("duration_ms", 150),
		zap.Float64("m_external_api_time", 150.0),
	)

	// Example 10: Metrics logging
	logger.Info("System metrics",
		zap.Float64("m_memory_usage_mb", 512.5),
		zap.Float64("m_cpu_usage_percent", 45.2),
		zap.Float64("m_disk_usage_percent", 78.9),
		zap.Float64("m_active_connections", 150.0),
		zap.Float64("m_requests_per_second", 25.5),
	)
}

// ExampleContextualLogging demonstrates contextual logging
func ExampleContextualLogging() {
	config := &logging.GrafanaLogConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		EnableTraceID:  true,
		EnableSpanID:   true,
		EnableMetrics:  true,
	}

	logger, err := logging.NewGrafanaLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Create contextual logger with common fields
	contextualLogger := logger.WithFields(
		zap.String("service", "auth-service"),
		zap.String("component", "authentication"),
		zap.String("version", "1.0.0"),
	)

	// All subsequent logs will include the common fields
	contextualLogger.Info("Authentication component started")
	contextualLogger.Info("Processing login request",
		zap.String("username", "john_doe"),
		zap.String("ip_address", "192.168.1.100"),
	)
}

// ExampleErrorHandling demonstrates proper error logging
func ExampleErrorHandling() {
	config := &logging.GrafanaLogConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		EnableTraceID:  true,
		EnableSpanID:   true,
		EnableMetrics:  true,
	}

	logger, err := logging.NewGrafanaLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Example of error handling with context
	requestID := "req-12345"
	userID := "user-67890"

	// Simulate an error
	err = fmt.Errorf("database connection failed: connection timeout")

	logger.Error("Failed to process request",
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("error", err.Error()),
		zap.String("error_type", "database_error"),
		zap.String("error_code", "DB_CONNECTION_TIMEOUT"),
		zap.Int("retry_count", 3),
		zap.Float64("m_error_rate", 0.05),
	)
}

// ExamplePerformanceLogging demonstrates performance logging
func ExamplePerformanceLogging() {
	config := &logging.GrafanaLogConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		EnableTraceID:  true,
		EnableSpanID:   true,
		EnableMetrics:  true,
	}

	logger, err := logging.NewGrafanaLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Example of performance logging
	requestID := "req-12345"
	startTime := time.Now()

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	duration := time.Since(startTime)

	logger.Info("Request processed",
		zap.String("request_id", requestID),
		zap.Int64("duration_ms", duration.Milliseconds()),
		zap.Float64("m_response_time", float64(duration.Milliseconds())),
		zap.Float64("m_throughput", 1.0),
		zap.Float64("m_success_rate", 1.0),
	)
}

// ExampleSecurityLogging demonstrates security logging
func ExampleSecurityLogging() {
	config := &logging.GrafanaLogConfig{
		ServiceName:    "auth-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		EnableTraceID:  true,
		EnableSpanID:   true,
		EnableMetrics:  true,
	}

	logger, err := logging.NewGrafanaLogger(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	defer logger.Sync()

	// Example of security logging
	requestID := "req-12345"
	ipAddress := "192.168.1.100"
	userAgent := "curl/7.68.0"

	logger.Warn("Security event detected",
		zap.String("request_id", requestID),
		zap.String("ip_address", ipAddress),
		zap.String("user_agent", userAgent),
		zap.String("security_event", "multiple_failed_logins"),
		zap.String("threat_level", "high"),
		zap.String("blocked_reason", "rate_limit_exceeded"),
		zap.Int("failed_attempts", 5),
		zap.Float64("m_security_events", 1.0),
		zap.Float64("m_blocked_requests", 1.0),
	)
}

package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GrafanaLogConfig holds configuration for Grafana-compatible logging
type GrafanaLogConfig struct {
	ServiceName    string            `json:"service_name"`
	ServiceVersion string            `json:"service_version"`
	Environment    string            `json:"environment"`
	Hostname       string            `json:"hostname"`
	Tags           map[string]string `json:"tags"`
	EnableTraceID  bool              `json:"enable_trace_id"`
	EnableSpanID   bool              `json:"enable_span_id"`
	EnableMetrics  bool              `json:"enable_metrics"`
}

// GrafanaLogEntry represents a structured log entry for Grafana
type GrafanaLogEntry struct {
	Timestamp   string                 `json:"@timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Hostname    string                 `json:"hostname"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Method      string                 `json:"method,omitempty"`
	URL         string                 `json:"url,omitempty"`
	StatusCode  int                    `json:"status_code,omitempty"`
	Duration    int64                  `json:"duration_ms,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Metrics     map[string]float64     `json:"metrics,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// GrafanaFormatter implements structured logging for Grafana
type GrafanaFormatter struct {
	config *GrafanaLogConfig
}

// NewGrafanaFormatter creates a new Grafana formatter
func NewGrafanaFormatter(config *GrafanaLogConfig) *GrafanaFormatter {
	if config.Hostname == "" {
		hostname, _ := os.Hostname()
		config.Hostname = hostname
	}

	return &GrafanaFormatter{
		config: config,
	}
}

// Format formats a log entry for Grafana
func (f *GrafanaFormatter) Format(entry zapcore.Entry, fields []zapcore.Field) ([]byte, error) {
	logEntry := &GrafanaLogEntry{
		Timestamp:   entry.Time.Format(time.RFC3339Nano),
		Level:       entry.Level.String(),
		Message:     entry.Message,
		Service:     f.config.ServiceName,
		Version:     f.config.ServiceVersion,
		Environment: f.config.Environment,
		Hostname:    f.config.Hostname,
		Tags:        f.config.Tags,
		Fields:      make(map[string]interface{}),
	}

	// Process fields
	for _, field := range fields {
		switch field.Key {
		case "trace_id":
			if f.config.EnableTraceID {
				logEntry.TraceID = fmt.Sprintf("%v", field.Interface)
			}
		case "span_id":
			if f.config.EnableSpanID {
				logEntry.SpanID = fmt.Sprintf("%v", field.Interface)
			}
		case "request_id":
			logEntry.RequestID = fmt.Sprintf("%v", field.Interface)
		case "user_id":
			logEntry.UserID = fmt.Sprintf("%v", field.Interface)
		case "session_id":
			logEntry.SessionID = fmt.Sprintf("%v", field.Interface)
		case "method":
			logEntry.Method = fmt.Sprintf("%v", field.Interface)
		case "url":
			logEntry.URL = fmt.Sprintf("%v", field.Interface)
		case "status_code":
			if code, ok := field.Interface.(int); ok {
				logEntry.StatusCode = code
			}
		case "duration_ms":
			if duration, ok := field.Interface.(int64); ok {
				logEntry.Duration = duration
			}
		case "ip_address":
			logEntry.IPAddress = fmt.Sprintf("%v", field.Interface)
		case "user_agent":
			logEntry.UserAgent = fmt.Sprintf("%v", field.Interface)
		case "error":
			logEntry.Error = fmt.Sprintf("%v", field.Interface)
		case "stack":
			logEntry.Stack = fmt.Sprintf("%v", field.Interface)
		default:
			// Add to fields map
			logEntry.Fields[field.Key] = field.Interface
		}
	}

	// Add stack trace for error level
	if entry.Level >= zapcore.ErrorLevel {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		logEntry.Stack = string(buf[:n])
	}

	// Add metrics if enabled
	if f.config.EnableMetrics {
		logEntry.Metrics = f.extractMetrics(fields)
	}

	return json.Marshal(logEntry)
}

// extractMetrics extracts metrics from log fields
func (f *GrafanaFormatter) extractMetrics(fields []zapcore.Field) map[string]float64 {
	metrics := make(map[string]float64)

	for _, field := range fields {
		if field.Key[:2] == "m_" { // Metrics fields start with "m_"
			if value, ok := field.Interface.(float64); ok {
				metrics[field.Key[2:]] = value
			}
		}
	}

	return metrics
}

// GrafanaLogger wraps zap.Logger with Grafana formatting
type GrafanaLogger struct {
	logger    *zap.Logger
	formatter *GrafanaFormatter
}

// NewGrafanaLogger creates a new Grafana logger
func NewGrafanaLogger(config *GrafanaLogConfig) (*GrafanaLogger, error) {
	formatter := NewGrafanaFormatter(config)

	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.TimeKey = "@timestamp"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.StacktraceKey = "stack"

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &GrafanaLogger{
		logger:    logger,
		formatter: formatter,
	}, nil
}

// Info logs an info message
func (g *GrafanaLogger) Info(message string, fields ...zap.Field) {
	g.logger.Info(message, fields...)
}

// Error logs an error message
func (g *GrafanaLogger) Error(message string, fields ...zap.Field) {
	g.logger.Error(message, fields...)
}

// Warn logs a warning message
func (g *GrafanaLogger) Warn(message string, fields ...zap.Field) {
	g.logger.Warn(message, fields...)
}

// Debug logs a debug message
func (g *GrafanaLogger) Debug(message string, fields ...zap.Field) {
	g.logger.Debug(message, fields...)
}

// Fatal logs a fatal message
func (g *GrafanaLogger) Fatal(message string, fields ...zap.Field) {
	g.logger.Fatal(message, fields...)
}

// WithFields creates a new logger with additional fields
func (g *GrafanaLogger) WithFields(fields ...zap.Field) *GrafanaLogger {
	return &GrafanaLogger{
		logger:    g.logger.With(fields...),
		formatter: g.formatter,
	}
}

// Sync flushes any buffered log entries
func (g *GrafanaLogger) Sync() error {
	return g.logger.Sync()
}

// GrafanaLoggingBestPractices provides best practices for Grafana logging
type GrafanaLoggingBestPractices struct{}

// GetBestPractices returns best practices for Grafana logging
func (g *GrafanaLoggingBestPractices) GetBestPractices() map[string]interface{} {
	return map[string]interface{}{
		"structured_logging": map[string]interface{}{
			"description": "Use structured logging with consistent field names",
			"fields": []string{
				"@timestamp", "level", "message", "service", "version",
				"environment", "hostname", "trace_id", "span_id", "request_id",
				"user_id", "method", "url", "status_code", "duration_ms",
				"ip_address", "user_agent", "error", "stack",
			},
		},
		"log_levels": map[string]interface{}{
			"description": "Use appropriate log levels",
			"levels": map[string]string{
				"DEBUG": "Detailed information for debugging",
				"INFO":  "General information about application flow",
				"WARN":  "Warning messages for potential issues",
				"ERROR": "Error messages for failed operations",
				"FATAL": "Critical errors that cause application termination",
			},
		},
		"correlation_ids": map[string]interface{}{
			"description": "Use correlation IDs for request tracing",
			"fields": []string{
				"trace_id", "span_id", "request_id", "user_id", "session_id",
			},
		},
		"performance_metrics": map[string]interface{}{
			"description": "Include performance metrics in logs",
			"metrics": []string{
				"duration_ms", "response_time", "throughput", "error_rate",
				"memory_usage", "cpu_usage", "database_queries", "cache_hits",
			},
		},
		"error_handling": map[string]interface{}{
			"description": "Proper error logging with context",
			"fields": []string{
				"error", "stack", "error_code", "error_type", "retry_count",
			},
		},
		"security_logging": map[string]interface{}{
			"description": "Security-related logging",
			"fields": []string{
				"ip_address", "user_agent", "authentication_method", "authorization_result",
				"security_event", "threat_level", "blocked_reason",
			},
		},
		"grafana_integration": map[string]interface{}{
			"description": "Grafana-specific logging considerations",
			"recommendations": []string{
				"Use consistent timestamp format (RFC3339Nano)",
				"Include service and version information",
				"Use structured fields for easy querying",
				"Include correlation IDs for distributed tracing",
				"Add performance metrics for monitoring",
				"Use appropriate log levels for alerting",
				"Include error context and stack traces",
			},
		},
	}
}

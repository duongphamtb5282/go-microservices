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

// APMLogConfig holds configuration for APM-compatible logging
type APMLogConfig struct {
	ServiceName    string            `json:"service_name"`
	ServiceVersion string            `json:"service_version"`
	Environment    string            `json:"environment"`
	Hostname       string            `json:"hostname"`
	Tags           map[string]string `json:"tags"`
	EnableTraceID  bool              `json:"enable_trace_id"`
	EnableSpanID   bool              `json:"enable_span_id"`
	EnableMetrics  bool              `json:"enable_metrics"`
	LogFormat      string            `json:"log_format"` // "json", "logfmt", "text"
}

// APMLogEntry represents a structured log entry for APM systems
type APMLogEntry struct {
	// Standard fields (compatible with most APM systems)
	Timestamp   string `json:"@timestamp"`
	Level       string `json:"level"`
	Message     string `json:"message"`
	Service     string `json:"service"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Hostname    string `json:"hostname"`

	// Correlation and tracing
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`

	// HTTP context
	Method     string `json:"method,omitempty"`
	URL        string `json:"url,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`
	Duration   int64  `json:"duration_ms,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`

	// Error context
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"error_type,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Stack     string `json:"stack,omitempty"`

	// Additional context
	Tags    map[string]string      `json:"tags,omitempty"`
	Metrics map[string]float64     `json:"metrics,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// APMFormatter implements structured logging for APM systems
type APMFormatter struct {
	config *APMLogConfig
}

// NewAPMFormatter creates a new APM formatter
func NewAPMFormatter(config *APMLogConfig) *APMFormatter {
	if config.Hostname == "" {
		hostname, _ := os.Hostname()
		config.Hostname = hostname
	}

	return &APMFormatter{
		config: config,
	}
}

// Format formats a log entry for APM systems
func (f *APMFormatter) Format(entry zapcore.Entry, fields []zapcore.Field) ([]byte, error) {
	logEntry := &APMLogEntry{
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
		case "error_type":
			logEntry.ErrorType = fmt.Sprintf("%v", field.Interface)
		case "error_code":
			logEntry.ErrorCode = fmt.Sprintf("%v", field.Interface)
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

	// Format based on configuration
	switch f.config.LogFormat {
	case "logfmt":
		return f.formatLogfmt(logEntry)
	case "text":
		return f.formatText(logEntry)
	default:
		return json.Marshal(logEntry)
	}
}

// formatLogfmt formats the log entry in logfmt format
func (f *APMFormatter) formatLogfmt(entry *APMLogEntry) ([]byte, error) {
	// This is a simplified logfmt implementation
	// In production, use a proper logfmt library
	return json.Marshal(entry)
}

// formatText formats the log entry in text format
func (f *APMFormatter) formatText(entry *APMLogEntry) ([]byte, error) {
	// This is a simplified text format implementation
	// In production, use a proper text formatter
	return json.Marshal(entry)
}

// extractMetrics extracts metrics from log fields
func (f *APMFormatter) extractMetrics(fields []zapcore.Field) map[string]float64 {
	metrics := make(map[string]float64)

	for _, field := range fields {
		if len(field.Key) > 2 && field.Key[:2] == "m_" { // Metrics fields start with "m_"
			if value, ok := field.Interface.(float64); ok {
				metrics[field.Key[2:]] = value
			}
		}
	}

	return metrics
}

// APMLogger wraps zap.Logger with APM formatting
type APMLogger struct {
	logger    *zap.Logger
	formatter *APMFormatter
}

// NewAPMLogger creates a new APM logger
func NewAPMLogger(config *APMLogConfig) (*APMLogger, error) {
	formatter := NewAPMFormatter(config)

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

	return &APMLogger{
		logger:    logger,
		formatter: formatter,
	}, nil
}

// Info logs an info message
func (a *APMLogger) Info(message string, fields ...zap.Field) {
	a.logger.Info(message, fields...)
}

// Error logs an error message
func (a *APMLogger) Error(message string, fields ...zap.Field) {
	a.logger.Error(message, fields...)
}

// Warn logs a warning message
func (a *APMLogger) Warn(message string, fields ...zap.Field) {
	a.logger.Warn(message, fields...)
}

// Debug logs a debug message
func (a *APMLogger) Debug(message string, fields ...zap.Field) {
	a.logger.Debug(message, fields...)
}

// Fatal logs a fatal message
func (a *APMLogger) Fatal(message string, fields ...zap.Field) {
	a.logger.Fatal(message, fields...)
}

// WithFields creates a new logger with additional fields
func (a *APMLogger) WithFields(fields ...zap.Field) *APMLogger {
	return &APMLogger{
		logger:    a.logger.With(fields...),
		formatter: a.formatter,
	}
}

// Sync flushes any buffered log entries
func (a *APMLogger) Sync() error {
	return a.logger.Sync()
}

// APMLoggingStandards provides standards for APM logging
type APMLoggingStandards struct{}

// GetStandards returns standards for APM logging based on industry best practices
func (a *APMLoggingStandards) GetStandards() map[string]interface{} {
	return map[string]interface{}{
		"standard_fields": []string{
			"@timestamp", "level", "message", "service", "version",
			"environment", "hostname", "trace_id", "span_id", "request_id",
			"user_id", "method", "url", "status_code", "duration_ms",
			"ip_address", "user_agent", "error", "stack",
		},
		"correlation_fields": []string{
			"trace_id", "span_id", "request_id", "user_id", "session_id",
		},
		"performance_metrics": []string{
			"duration_ms", "response_time", "throughput", "error_rate",
			"memory_usage", "cpu_usage", "database_queries", "cache_hits",
		},
		"error_fields": []string{
			"error", "error_type", "error_code", "stack", "retry_count",
		},
		"security_fields": []string{
			"ip_address", "user_agent", "authentication_method", "authorization_result",
			"security_event", "threat_level", "blocked_reason",
		},
		"formats": []string{
			"json", "logfmt", "text",
		},
		"timestamp_format": "RFC3339Nano",
		"log_levels": []string{
			"DEBUG", "INFO", "WARN", "ERROR", "FATAL",
		},
	}
}

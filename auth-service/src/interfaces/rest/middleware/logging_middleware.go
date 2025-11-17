package middleware

import (
	"bytes"
	"io"
	"time"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware provides comprehensive logging functionality
type LoggingMiddleware struct {
	logger *logging.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// RequestLogging provides request logging
func (m *LoggingMiddleware) RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request body if needed
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture response
		responseWriter := &LoggingResponseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		m.logger.Info("HTTP Request",
			logging.String("method", c.Request.Method),
			logging.String("path", c.Request.URL.Path),
			logging.String("query", c.Request.URL.RawQuery),
			logging.String("ip", c.ClientIP()),
			logging.String("user_agent", c.Request.UserAgent()),
			logging.Int("status", responseWriter.statusCode),
			logging.Duration("duration", duration),
			logging.Int("response_size", len(responseWriter.body)),
			logging.String("request_id", c.GetString("request_id")),
		)

		// Log request body for debugging (if enabled)
		if len(requestBody) > 0 && len(requestBody) < 1024 {
			m.logger.Debug("Request body", logging.String("body", string(requestBody)))
		}

		// Log response body for debugging (if enabled)
		if len(responseWriter.body) > 0 && len(responseWriter.body) < 1024 {
			m.logger.Debug("Response body", logging.String("body", string(responseWriter.body)))
		}

		// Log error for failed requests
		if responseWriter.statusCode >= 400 {
			m.logger.Error("HTTP Request Failed",
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.Int("status", responseWriter.statusCode),
				logging.Duration("duration", duration),
				logging.String("error", string(responseWriter.body)),
			)
		}
	}
}

// AuditLogging provides audit logging
func (m *LoggingMiddleware) AuditLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue with the next handler
		c.Next()

		// Log audit information for specific operations
		if m.shouldAudit(c) {
			m.logger.Info("Audit Log",
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.String("ip", c.ClientIP()),
				logging.String("user_agent", c.Request.UserAgent()),
				logging.Int("status", c.Writer.Status()),
				logging.String("user_id", c.GetString("user_id")),
				logging.String("request_id", c.GetString("request_id")),
				logging.Time("timestamp", time.Now()),
			)
		}
	}
}

// ErrorLogging provides error logging
func (m *LoggingMiddleware) ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue with the next handler
		c.Next()

		// Log errors
		if c.Writer.Status() >= 400 {
			m.logger.Error("HTTP Error",
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.Int("status", c.Writer.Status()),
				logging.String("ip", c.ClientIP()),
				logging.String("user_agent", c.Request.UserAgent()),
				logging.String("error", c.Errors.String()),
			)
		}
	}
}

// PerformanceLogging provides performance logging
func (m *LoggingMiddleware) PerformanceLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Continue with the next handler
		c.Next()

		duration := time.Since(start)

		// Log performance metrics
		m.logger.Info("Performance Metrics",
			logging.String("method", c.Request.Method),
			logging.String("path", c.Request.URL.Path),
			logging.Duration("duration", duration),
			logging.String("request_id", c.GetString("request_id")),
		)

		// Log slow requests
		if duration > 1*time.Second {
			m.logger.Warn("Slow Request",
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.Duration("duration", duration),
				logging.String("request_id", c.GetString("request_id")),
			)
		}
	}
}

// SecurityLogging provides security logging
func (m *LoggingMiddleware) SecurityLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log security events
		if m.isSecurityEvent(c) {
			m.logger.Warn("Security Event",
				logging.String("method", c.Request.Method),
				logging.String("path", c.Request.URL.Path),
				logging.String("ip", c.ClientIP()),
				logging.String("user_agent", c.Request.UserAgent()),
				logging.String("event", "suspicious_activity"),
				logging.Time("timestamp", time.Now()),
			)
		}

		// Continue with the next handler
		c.Next()
	}
}

// shouldAudit checks if the request should be audited
func (m *LoggingMiddleware) shouldAudit(c *gin.Context) bool {
	// Audit specific operations
	auditMethods := []string{"POST", "PUT", "DELETE"}
	auditPaths := []string{"/api/v1/users"}

	for _, method := range auditMethods {
		if c.Request.Method == method {
			for _, path := range auditPaths {
				if c.Request.URL.Path == path ||
					(len(path) > 0 && path[len(path)-1] == '/' &&
						c.Request.URL.Path[:len(path)-1] == path[:len(path)-1]) {
					return true
				}
			}
		}
	}

	return false
}

// isSecurityEvent checks if the request is a security event
func (m *LoggingMiddleware) isSecurityEvent(c *gin.Context) bool {
	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"../", "..\\", "script", "javascript:", "vbscript:",
		"<script", "</script", "onload=", "onerror=",
	}

	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery

	for _, pattern := range suspiciousPatterns {
		if contains(path, pattern) || contains(query, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// LoggingResponseWriter is a custom response writer for logging
type LoggingResponseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

// Write captures the response body
func (w *LoggingResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

// WriteHeader captures the status code
func (w *LoggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Status returns the status code
func (w *LoggingResponseWriter) Status() int {
	return w.statusCode
}

// Body returns the response body
func (w *LoggingResponseWriter) Body() []byte {
	return w.body
}

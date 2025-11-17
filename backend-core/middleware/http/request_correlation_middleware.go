package http

import (
	"backend-core/logging"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestCorrelationConfig holds unified request/correlation ID configuration
type RequestCorrelationConfig struct {
	// Header configuration
	RequestIDHeader     string
	CorrelationIDHeader string

	// ID generation
	GenerateID    func() string
	UseUUIDFormat bool

	// Behavior flags
	SetInContext  bool
	SetInResponse bool
	LogRequests   bool
	LogTiming     bool

	// Propagation settings
	PropagateTo    []string
	StoreInContext bool
}

// DefaultRequestCorrelationConfig returns default configuration
func DefaultRequestCorrelationConfig() *RequestCorrelationConfig {
	return &RequestCorrelationConfig{
		RequestIDHeader:     "X-Request-ID",
		CorrelationIDHeader: "X-Correlation-ID",
		GenerateID:          generateOptimizedID,
		UseUUIDFormat:       true,
		SetInContext:        true,
		SetInResponse:       true,
		LogRequests:         true,
		LogTiming:           true,
		PropagateTo:         []string{},
		StoreInContext:      true,
	}
}

// RequestCorrelationMiddleware provides unified request/correlation ID functionality
type RequestCorrelationMiddleware struct {
	config *RequestCorrelationConfig
	logger *logging.Logger
}

// NewRequestCorrelationMiddleware creates a new unified middleware
func NewRequestCorrelationMiddleware(config *RequestCorrelationConfig, logger *logging.Logger) *RequestCorrelationMiddleware {
	if config == nil {
		config = DefaultRequestCorrelationConfig()
	}
	return &RequestCorrelationMiddleware{
		config: config,
		logger: logger,
	}
}

// Handler returns the unified middleware handler
func (r *RequestCorrelationMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get or generate request ID
		requestID := ctx.GetHeader(r.config.RequestIDHeader)
		if requestID == "" {
			requestID = r.config.GenerateID()
		}

		// Get or generate correlation ID (use request ID if not provided)
		correlationID := ctx.GetHeader(r.config.CorrelationIDHeader)
		if correlationID == "" {
			correlationID = requestID // Use request ID as correlation ID for simplicity
		}

		// Set headers in response
		if r.config.SetInResponse {
			ctx.Header(r.config.RequestIDHeader, requestID)
			ctx.Header(r.config.CorrelationIDHeader, correlationID)
		}

		// Set in context for use in handlers
		if r.config.SetInContext {
			ctx.Set("request_id", requestID)
			ctx.Set("correlation_id", correlationID)
		}

		// Add to request context for propagation
		reqCtx := context.WithValue(ctx.Request.Context(), "request_id", requestID)
		reqCtx = context.WithValue(reqCtx, "correlation_id", correlationID)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		// Log request start
		if r.config.LogRequests {
			r.logger.Info("Request started, request_id: %s, correlation_id: %s, method: %s, path: %s, ip_address: %s, user_agent: %s",
				requestID, correlationID, ctx.Request.Method, ctx.Request.URL.Path, ctx.ClientIP(), ctx.Request.UserAgent())
		}

		// Start timing
		start := time.Now()

		// Process request
		ctx.Next()

		// Log completion with timing
		if r.config.LogTiming {
			duration := time.Since(start)
			r.logger.Info("Request completed, request_id: %s, correlation_id: %s, method: %s, path: %s, status_code: %d, duration_ms: %d, response_time: %.2f",
				requestID, correlationID, ctx.Request.Method, ctx.Request.URL.Path, ctx.Writer.Status(), duration.Milliseconds(), float64(duration.Milliseconds()))
		}
	}
}

// generateOptimizedID generates an optimized ID (UUID v4 or timestamp-based)
func generateOptimizedID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Set version (4) and variant bits for UUID v4
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant bits

	// Convert to hex string
	hexStr := hex.EncodeToString(bytes)

	// Format as UUID
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32])
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

// GetCorrelationIDFromContext extracts correlation ID from context
func GetCorrelationIDFromContext(ctx context.Context) string {
	if correlationID, ok := ctx.Value("correlation_id").(string); ok {
		return correlationID
	}
	return ""
}

// GetRequestIDFromGin extracts request ID from Gin context
func GetRequestIDFromGin(ctx *gin.Context) string {
	if requestID, exists := ctx.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetCorrelationIDFromGin extracts correlation ID from Gin context
func GetCorrelationIDFromGin(ctx *gin.Context) string {
	if correlationID, exists := ctx.Get("correlation_id"); exists {
		if id, ok := correlationID.(string); ok {
			return id
		}
	}
	return ""
}

// RequestCorrelationPropagator handles ID propagation to other services
type RequestCorrelationPropagator struct {
	config *RequestCorrelationConfig
	logger *logging.Logger
}

// NewRequestCorrelationPropagator creates a new propagator
func NewRequestCorrelationPropagator(config *RequestCorrelationConfig, logger *logging.Logger) *RequestCorrelationPropagator {
	return &RequestCorrelationPropagator{
		config: config,
		logger: logger,
	}
}

// PropagateToHTTPRequest adds IDs to HTTP request headers
func (p *RequestCorrelationPropagator) PropagateToHTTPRequest(req *http.Request, requestID, correlationID string) {
	if requestID != "" {
		req.Header.Set(p.config.RequestIDHeader, requestID)
	}
	if correlationID != "" {
		req.Header.Set(p.config.CorrelationIDHeader, correlationID)
	}

	p.logger.Debug("IDs propagated to HTTP request, request_id: %s, correlation_id: %s, url: %s",
		requestID, correlationID, req.URL.String())
}

// PropagateToKafkaMessage adds IDs to Kafka message headers
func (p *RequestCorrelationPropagator) PropagateToKafkaMessage(headers map[string]string, requestID, correlationID string) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}

	if requestID != "" {
		headers[p.config.RequestIDHeader] = requestID
		headers["request_id"] = requestID // Also add as request_id for compatibility
	}
	if correlationID != "" {
		headers[p.config.CorrelationIDHeader] = correlationID
		headers["correlation_id"] = correlationID // Also add as correlation_id for compatibility
	}

	p.logger.Debug("IDs propagated to Kafka message, request_id: %s, correlation_id: %s", requestID, correlationID)
	return headers
}

// PropagateToDatabaseQuery adds IDs to database query context
func (p *RequestCorrelationPropagator) PropagateToDatabaseQuery(ctx context.Context, requestID, correlationID string) context.Context {
	if requestID != "" {
		ctx = context.WithValue(ctx, "request_id", requestID)
	}
	if correlationID != "" {
		ctx = context.WithValue(ctx, "correlation_id", correlationID)
	}

	p.logger.Debug("IDs propagated to database query, request_id: %s, correlation_id: %s", requestID, correlationID)
	return ctx
}

// ExtractFromHTTPRequest extracts IDs from HTTP request
func (p *RequestCorrelationPropagator) ExtractFromHTTPRequest(req *http.Request) (requestID, correlationID string) {
	requestID = req.Header.Get(p.config.RequestIDHeader)
	correlationID = req.Header.Get(p.config.CorrelationIDHeader)

	// Try alternative header names if not found
	if requestID == "" {
		requestID = req.Header.Get("X-Request-ID")
	}
	if correlationID == "" {
		correlationID = req.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = req.Header.Get("X-Trace-ID")
		}
	}

	return requestID, correlationID
}

// ExtractFromKafkaMessage extracts IDs from Kafka message headers
func (p *RequestCorrelationPropagator) ExtractFromKafkaMessage(headers map[string]string) (requestID, correlationID string) {
	if headers == nil {
		return "", ""
	}

	requestID = headers[p.config.RequestIDHeader]
	correlationID = headers[p.config.CorrelationIDHeader]

	// Try alternative header names if not found
	if requestID == "" {
		requestID = headers["X-Request-ID"]
		if requestID == "" {
			requestID = headers["request_id"]
		}
	}
	if correlationID == "" {
		correlationID = headers["X-Correlation-ID"]
		if correlationID == "" {
			correlationID = headers["correlation_id"]
		}
	}

	return requestID, correlationID
}

// RequestCorrelationContext provides unified context for services
type RequestCorrelationContext struct {
	RequestID     string
	CorrelationID string
	UserID        string
	SessionID     string
	TraceID       string
	SpanID        string
	ServiceName   string
	StartTime     time.Time
	Metadata      map[string]string
}

// NewRequestCorrelationContext creates a new unified context
func NewRequestCorrelationContext(requestID, correlationID string) *RequestCorrelationContext {
	return &RequestCorrelationContext{
		RequestID:     requestID,
		CorrelationID: correlationID,
		StartTime:     time.Now(),
		Metadata:      make(map[string]string),
	}
}

// WithUserID sets the user ID in the context
func (c *RequestCorrelationContext) WithUserID(userID string) *RequestCorrelationContext {
	c.UserID = userID
	return c
}

// WithSessionID sets the session ID in the context
func (c *RequestCorrelationContext) WithSessionID(sessionID string) *RequestCorrelationContext {
	c.SessionID = sessionID
	return c
}

// WithTraceID sets the trace ID in the context
func (c *RequestCorrelationContext) WithTraceID(traceID string) *RequestCorrelationContext {
	c.TraceID = traceID
	return c
}

// WithSpanID sets the span ID in the context
func (c *RequestCorrelationContext) WithSpanID(spanID string) *RequestCorrelationContext {
	c.SpanID = spanID
	return c
}

// WithServiceName sets the service name in the context
func (c *RequestCorrelationContext) WithServiceName(serviceName string) *RequestCorrelationContext {
	c.ServiceName = serviceName
	return c
}

// WithMetadata adds metadata to the context
func (c *RequestCorrelationContext) WithMetadata(key, value string) *RequestCorrelationContext {
	c.Metadata[key] = value
	return c
}

// GetDuration returns the duration since the context was created
func (c *RequestCorrelationContext) GetDuration() time.Duration {
	return time.Since(c.StartTime)
}

// ToLogFields converts the context to log fields
func (c *RequestCorrelationContext) ToLogFields() []logging.Field {
	// Build log message with all available context
	message := fmt.Sprintf("request_id: %s, correlation_id: %s", c.RequestID, c.CorrelationID)

	if c.UserID != "" {
		message += fmt.Sprintf(", user_id: %s", c.UserID)
	}

	if c.SessionID != "" {
		message += fmt.Sprintf(", session_id: %s", c.SessionID)
	}

	if c.TraceID != "" {
		message += fmt.Sprintf(", trace_id: %s", c.TraceID)
	}

	if c.SpanID != "" {
		message += fmt.Sprintf(", span_id: %s", c.SpanID)
	}

	if c.ServiceName != "" {
		message += fmt.Sprintf(", service: %s", c.ServiceName)
	}

	message += fmt.Sprintf(", duration_ms: %d", c.GetDuration().Milliseconds())

	// Return a simple field with the formatted message
	return []logging.Field{logging.String("context", message)}
}

// RequestCorrelationTestHelper provides utilities for testing
type RequestCorrelationTestHelper struct {
	logger *logging.Logger
}

// NewRequestCorrelationTestHelper creates a new test helper
func NewRequestCorrelationTestHelper(logger *logging.Logger) *RequestCorrelationTestHelper {
	return &RequestCorrelationTestHelper{
		logger: logger,
	}
}

// TestSyncPropagation tests synchronous ID propagation
func (h *RequestCorrelationTestHelper) TestSyncPropagation(requestID, correlationID string) {
	h.logger.Info("Testing synchronous propagation, request_id: %s, correlation_id: %s, test_type: sync", requestID, correlationID)

	// Simulate synchronous service calls
	h.logger.Info("Calling user service, request_id: %s, correlation_id: %s, service: user-service", requestID, correlationID)
	h.logger.Info("Calling auth service, request_id: %s, correlation_id: %s, service: auth-service", requestID, correlationID)
	h.logger.Info("Calling notification service, request_id: %s, correlation_id: %s, service: notification-service", requestID, correlationID)
}

// TestAsyncPropagation tests asynchronous ID propagation
func (h *RequestCorrelationTestHelper) TestAsyncPropagation(requestID, correlationID string) {
	h.logger.Info("Testing asynchronous propagation, request_id: %s, correlation_id: %s, test_type: async", requestID, correlationID)

	// Simulate asynchronous event publishing
	h.logger.Info("Publishing user registered event, request_id: %s, correlation_id: %s, event_type: user.registered, topic: user.events", requestID, correlationID)
	h.logger.Info("Publishing user activated event, request_id: %s, correlation_id: %s, event_type: user.activated, topic: user.events", requestID, correlationID)

	// Simulate event consumption
	h.logger.Info("Consuming user registered event, request_id: %s, correlation_id: %s, service: notification-service, event_type: user.registered", requestID, correlationID)
}

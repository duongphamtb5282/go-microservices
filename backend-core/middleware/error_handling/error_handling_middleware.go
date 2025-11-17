package error_handling

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"backend-core/logging"
	"backend-shared/errors"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
	TraceID string            `json:"trace_id,omitempty"`
}

// ErrorMiddleware provides comprehensive error handling
type ErrorMiddleware struct {
	logger *logging.Logger
}

// CreateErrorMiddleware creates an error middleware
func CreateErrorMiddleware(logger *logging.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

// Handler returns the Gin middleware handler
func (em *ErrorMiddleware) Handler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		em.handlePanic(c, recovered)
	})
}

// ErrorHandler handles errors from the service layer
func (em *ErrorMiddleware) ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			em.handleErrors(c)
		}
	}
}

// handlePanic handles panic recovery
func (em *ErrorMiddleware) handlePanic(c *gin.Context, recovered interface{}) {
	// Log the panic
	em.logger.Error("Panic recovered: %v, path: %s, method: %s, stack: %s",
		recovered, c.Request.URL.Path, c.Request.Method, string(debug.Stack()))

	// Create error response
	response := ErrorResponse{
		Error:   "Internal Server Error",
		Message: "An unexpected error occurred",
		Code:    "INTERNAL_ERROR",
		TraceID: c.GetString("trace_id"),
	}

	c.JSON(http.StatusInternalServerError, response)
	c.Abort()
}

// handleErrors handles service layer errors
func (em *ErrorMiddleware) handleErrors(c *gin.Context) {
	// Get the last error
	err := c.Errors.Last()
	if err == nil {
		return
	}

	// Log the error
	em.logger.Error("Service error: %v, path: %s, method: %s, trace_id: %s",
		err.Err, c.Request.URL.Path, c.Request.Method, c.GetString("trace_id"))

	// Determine error type and create appropriate response
	response := em.createErrorResponse(err.Err, c)
	statusCode := em.getStatusCode(err.Err)

	c.JSON(statusCode, response)
	c.Abort()
}

// createErrorResponse creates an error response based on the error type
func (em *ErrorMiddleware) createErrorResponse(err error, c *gin.Context) ErrorResponse {
	response := ErrorResponse{
		TraceID: c.GetString("trace_id"),
	}

	// Check for specific error types
	switch e := err.(type) {
	case errors.ValidationError:
		response.Error = "Validation Error"
		response.Message = e.Message
		response.Code = e.Code
		response.Details = map[string]string{
			"field": e.Field,
			"value": fmt.Sprintf("%v", e.Value),
		}

	case errors.DomainError:
		response.Error = "Domain Error"
		response.Message = e.Message
		response.Code = e.Code
		// Convert details to string map
		response.Details = make(map[string]string)
		for k, v := range e.Details {
			response.Details[k] = fmt.Sprintf("%v", v)
		}

	default:
		// Generic error handling
		response.Error = "Internal Server Error"
		response.Message = err.Error()
		response.Code = "INTERNAL_ERROR"
	}

	return response
}

// getStatusCode returns the appropriate HTTP status code for an error
func (em *ErrorMiddleware) getStatusCode(err error) int {
	switch e := err.(type) {
	case errors.ValidationError:
		return http.StatusBadRequest
	case errors.DomainError:
		// Check domain error code for specific status codes
		switch e.Code {
		case errors.ErrCodeNotFound:
			return http.StatusNotFound
		case errors.ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case errors.ErrCodeForbidden:
			return http.StatusForbidden
		case errors.ErrCodeAlreadyExists:
			return http.StatusConflict
		default:
			return http.StatusBadRequest
		}
	default:
		return http.StatusInternalServerError
	}
}

// ErrorHandlerFunc is a function that handles errors
type ErrorHandlerFunc func(c *gin.Context, err error)

// CustomErrorHandler creates a custom error handler
func (em *ErrorMiddleware) CustomErrorHandler(handler ErrorHandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			handler(c, err.Err)
		}
	}
}

// LogError logs an error with context
func (em *ErrorMiddleware) LogError(c *gin.Context, err error, message string) {
	em.logger.Error("%s: %v, path: %s, method: %s, trace_id: %s, user_id: %s",
		message, err, c.Request.URL.Path, c.Request.Method, c.GetString("trace_id"), c.GetString("user_id"))
}

// LogWarning logs a warning with context
func (em *ErrorMiddleware) LogWarning(c *gin.Context, message string, fields ...logging.Field) {
	em.logger.Warn("%s, path: %s, method: %s, trace_id: %s, user_id: %s",
		message, c.Request.URL.Path, c.Request.Method, c.GetString("trace_id"), c.GetString("user_id"))
}

// LogInfo logs an info message with context
func (em *ErrorMiddleware) LogInfo(c *gin.Context, message string, fields ...logging.Field) {
	em.logger.Info("%s, path: %s, method: %s, trace_id: %s, user_id: %s",
		message, c.Request.URL.Path, c.Request.Method, c.GetString("trace_id"), c.GetString("user_id"))
}

// JSONError sends a JSON error response
func (em *ErrorMiddleware) JSONError(c *gin.Context, statusCode int, err error) {
	response := em.createErrorResponse(err, c)
	c.JSON(statusCode, response)
}

// ValidationError sends a validation error response
func (em *ErrorMiddleware) ValidationError(c *gin.Context, message string, details map[string]string) {
	response := ErrorResponse{
		Error:   "Validation Error",
		Message: message,
		Code:    "VALIDATION_ERROR",
		Details: details,
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusBadRequest, response)
}

// BusinessError sends a business error response
func (em *ErrorMiddleware) BusinessError(c *gin.Context, message string, code string) {
	response := ErrorResponse{
		Error:   "Business Error",
		Message: message,
		Code:    code,
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusBadRequest, response)
}

// NotFoundError sends a not found error response
func (em *ErrorMiddleware) NotFoundError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Not Found",
		Message: message,
		Code:    "NOT_FOUND",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusNotFound, response)
}

// UnauthorizedError sends an unauthorized error response
func (em *ErrorMiddleware) UnauthorizedError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Unauthorized",
		Message: message,
		Code:    "UNAUTHORIZED",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusUnauthorized, response)
}

// ForbiddenError sends a forbidden error response
func (em *ErrorMiddleware) ForbiddenError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Forbidden",
		Message: message,
		Code:    "FORBIDDEN",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusForbidden, response)
}

// ConflictError sends a conflict error response
func (em *ErrorMiddleware) ConflictError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Conflict",
		Message: message,
		Code:    "CONFLICT",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusConflict, response)
}

// RateLimitError sends a rate limit error response
func (em *ErrorMiddleware) RateLimitError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Rate Limit Exceeded",
		Message: message,
		Code:    "RATE_LIMIT_EXCEEDED",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusTooManyRequests, response)
}

// ServiceUnavailableError sends a service unavailable error response
func (em *ErrorMiddleware) ServiceUnavailableError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Service Unavailable",
		Message: message,
		Code:    "SERVICE_UNAVAILABLE",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusServiceUnavailable, response)
}

// InternalServerError sends an internal server error response
func (em *ErrorMiddleware) InternalServerError(c *gin.Context, message string) {
	response := ErrorResponse{
		Error:   "Internal Server Error",
		Message: message,
		Code:    "INTERNAL_ERROR",
		TraceID: c.GetString("trace_id"),
	}
	c.JSON(http.StatusInternalServerError, response)
}

// RequestInfo represents request information for logging
type RequestInfo struct {
	RequestID string
	Method    string
	Path      string
	UserAgent string
	IP        string
	UserID    string
}

// getRequestID extracts request ID from context or headers
func (em *ErrorMiddleware) getRequestID(c *gin.Context) string {
	// Check for request ID in context
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}

	// Check for request ID in headers
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate new request ID
	requestID := em.generateRequestID()
	c.Set("request_id", requestID)
	return requestID
}

// getUserID extracts user ID from context
func (em *ErrorMiddleware) getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// generateRequestID generates a unique request ID
func (em *ErrorMiddleware) generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + em.randomString(8)
}

// randomString generates a random string of specified length
func (em *ErrorMiddleware) randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// logRequestCompletion logs the completion of a request
func (em *ErrorMiddleware) logRequestCompletion(c *gin.Context, duration time.Duration) {
	requestID := em.getRequestID(c)
	userID := em.getUserID(c)

	em.logger.Info("Request completed",
		logging.String("request_id", requestID),
		logging.String("method", c.Request.Method),
		logging.String("path", c.Request.URL.Path),
		logging.String("user_id", userID),
		logging.Duration("duration", duration),
		logging.Int("status", c.Writer.Status()),
	)
}

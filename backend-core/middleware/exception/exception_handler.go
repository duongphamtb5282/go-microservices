package exception

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"backend-core/logging"
	"backend-shared/errors"

	"go.uber.org/zap"
)

// ExceptionHandler defines the interface for exception handling
type ExceptionHandler interface {
	HandleException(ctx context.Context, err error, requestInfo RequestInfo) ErrorResponse
	HandlePanic(ctx context.Context, panicValue interface{}, requestInfo RequestInfo) ErrorResponse
}

// RequestInfo contains information about the HTTP request
type RequestInfo struct {
	RequestID string
	Method    string
	Path      string
	UserAgent string
	IP        string
	UserID    string
}

// DefaultExceptionHandler provides default exception handling
type DefaultExceptionHandler struct {
	logger       *logging.Logger
	includeStack bool
	hideInternal bool
	serviceName  string
	version      string
}

// NewDefaultExceptionHandler creates a new default exception handler
func NewDefaultExceptionHandler(logger *logging.Logger, serviceName, version string) *DefaultExceptionHandler {
	return &DefaultExceptionHandler{
		logger:       logger,
		includeStack: true,
		hideInternal: true,
		serviceName:  serviceName,
		version:      version,
	}
}

// WithStackTrace enables or disables stack trace inclusion
func (h *DefaultExceptionHandler) WithStackTrace(include bool) *DefaultExceptionHandler {
	h.includeStack = include
	return h
}

// WithHideInternal enables or disables hiding internal error details
func (h *DefaultExceptionHandler) WithHideInternal(hide bool) *DefaultExceptionHandler {
	h.hideInternal = hide
	return h
}

// HandleException handles application exceptions
func (h *DefaultExceptionHandler) HandleException(ctx context.Context, err error, requestInfo RequestInfo) ErrorResponse {
	// Log the exception
	h.logException(ctx, err, requestInfo)

	// Determine error type and create appropriate response
	switch e := err.(type) {
	case errors.DomainError:
		return h.handleDomainError(e, requestInfo)
	case errors.ValidationError:
		return h.handleValidationError(e, requestInfo)
	case errors.ValidationErrors:
		return h.handleValidationErrors(e, requestInfo)
	default:
		return h.handleGenericError(err, requestInfo)
	}
}

// HandlePanic handles panic recovery
func (h *DefaultExceptionHandler) HandlePanic(ctx context.Context, panicValue interface{}, requestInfo RequestInfo) ErrorResponse {
	// Log the panic
	h.logPanic(ctx, panicValue, requestInfo)

	// Create panic error response
	details := make(map[string]interface{})
	if h.includeStack {
		details["stack"] = h.getStackTrace()
	}
	details["panic_value"] = fmt.Sprintf("%v", panicValue)

	response := NewInternalErrorResponse("Internal server error occurred")
	response = response.WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)
	response.Error.Details = details

	return response
}

// handleDomainError handles domain-specific errors
func (h *DefaultExceptionHandler) handleDomainError(err errors.DomainError, requestInfo RequestInfo) ErrorResponse {
	response := NewDomainErrorResponse(err)
	response = response.WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)

	// Add additional details if needed
	if h.includeStack {
		response.Error.Details["stack"] = h.getStackTrace()
	}

	return response
}

// handleValidationError handles single validation errors
func (h *DefaultExceptionHandler) handleValidationError(err errors.ValidationError, requestInfo RequestInfo) ErrorResponse {
	details := map[string]interface{}{
		"field": err.Field,
		"value": err.Value,
	}

	response := NewValidationErrorResponse(err.Message, details)
	response = response.WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)

	return response
}

// handleValidationErrors handles multiple validation errors
func (h *DefaultExceptionHandler) handleValidationErrors(errs errors.ValidationErrors, requestInfo RequestInfo) ErrorResponse {
	response := NewValidationErrorsResponse(errs)
	response = response.WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)

	return response
}

// handleGenericError handles generic errors
func (h *DefaultExceptionHandler) handleGenericError(err error, requestInfo RequestInfo) ErrorResponse {
	// Check if it's a known error type by message
	errorMessage := err.Error()

	var response ErrorResponse

	switch {
	case strings.Contains(errorMessage, "not found"):
		response = NewNotFoundErrorResponse("Resource not found")
	case strings.Contains(errorMessage, "unauthorized"):
		response = NewUnauthorizedErrorResponse("Unauthorized access")
	case strings.Contains(errorMessage, "forbidden"):
		response = NewForbiddenErrorResponse("Access forbidden")
	case strings.Contains(errorMessage, "validation"):
		response = NewValidationErrorResponse("Validation failed", nil)
	default:
		// Internal server error
		message := "Internal server error"
		if !h.hideInternal {
			message = errorMessage
		}
		response = NewInternalErrorResponse(message)
	}

	response = response.WithRequestInfo(requestInfo.RequestID, requestInfo.Path, requestInfo.Method)

	// Add stack trace if enabled and not hiding internal details
	if h.includeStack && !h.hideInternal {
		response.Error.Details["stack"] = h.getStackTrace()
	}

	return response
}

// logException logs the exception with context
func (h *DefaultExceptionHandler) logException(ctx context.Context, err error, requestInfo RequestInfo) {
	// Create context with request information
	ctx = context.WithValue(ctx, "request_id", requestInfo.RequestID)
	ctx = context.WithValue(ctx, "user_id", requestInfo.UserID)
	ctx = context.WithValue(ctx, "service_name", h.serviceName)

	// Get logger with context
	logger := h.logger.WithContext(ctx)

	// Create fields map for additional context
	fields := map[string]interface{}{
		"method":     requestInfo.Method,
		"path":       requestInfo.Path,
		"user_agent": requestInfo.UserAgent,
		"ip":         requestInfo.IP,
		"version":    h.version,
	}

	// Add stack trace for internal errors
	if h.includeStack {
		fields["stack"] = h.getStackTrace()
	}

	// Convert fields to zap fields
	zapFields := make([]zap.Field, 0, len(fields)+1)
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	zapFields = append(zapFields, zap.Error(err))

	// Log with error
	logger.Error("Exception occurred", zapFields...)
}

// logPanic logs the panic with context
func (h *DefaultExceptionHandler) logPanic(ctx context.Context, panicValue interface{}, requestInfo RequestInfo) {
	// Create context with request information
	ctx = context.WithValue(ctx, "request_id", requestInfo.RequestID)
	ctx = context.WithValue(ctx, "user_id", requestInfo.UserID)
	ctx = context.WithValue(ctx, "service_name", h.serviceName)

	// Get logger with context
	logger := h.logger.WithContext(ctx)

	// Create fields map for additional context
	fields := map[string]interface{}{
		"panic_value": panicValue,
		"method":      requestInfo.Method,
		"path":        requestInfo.Path,
		"user_agent":  requestInfo.UserAgent,
		"ip":          requestInfo.IP,
		"version":     h.version,
		"stack":       h.getStackTrace(),
	}

	// Convert fields to zap fields
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	// Log panic
	logger.Error("Panic occurred", zapFields...)
}

// getStackTrace returns the current stack trace
func (h *DefaultExceptionHandler) getStackTrace() string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		buf = make([]byte, 2*len(buf))
	}
}

// GetHTTPStatusCode returns the appropriate HTTP status code for an error
func GetHTTPStatusCode(err error) int {
	switch e := err.(type) {
	case errors.DomainError:
		switch e.Code {
		case errors.ErrCodeValidation:
			return http.StatusBadRequest
		case errors.ErrCodeNotFound:
			return http.StatusNotFound
		case errors.ErrCodeAlreadyExists:
			return http.StatusConflict
		case errors.ErrCodeUnauthorized:
			return http.StatusUnauthorized
		case errors.ErrCodeForbidden:
			return http.StatusForbidden
		case errors.ErrCodeInvalidState:
			return http.StatusBadRequest
		case errors.ErrCodeExternal:
			return http.StatusBadGateway
		default:
			return http.StatusInternalServerError
		}
	case errors.ValidationError, errors.ValidationErrors:
		return http.StatusBadRequest
	default:
		// Check error message for common patterns
		errorMessage := err.Error()
		switch {
		case strings.Contains(errorMessage, "not found"):
			return http.StatusNotFound
		case strings.Contains(errorMessage, "unauthorized"):
			return http.StatusUnauthorized
		case strings.Contains(errorMessage, "forbidden"):
			return http.StatusForbidden
		case strings.Contains(errorMessage, "validation"):
			return http.StatusBadRequest
		default:
			return http.StatusInternalServerError
		}
	}
}

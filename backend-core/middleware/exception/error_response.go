package exception

import (
	"encoding/json"
	"net/http"
	"time"

	"backend-shared/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     ErrorDetails `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
	Path      string       `json:"path,omitempty"`
	Method    string       `json:"method,omitempty"`
}

// ErrorDetails contains detailed error information
type ErrorDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Type    string                 `json:"type"`
}

// NewErrorResponse creates a new standardized error response
func NewErrorResponse(code, message, errorType string, details map[string]interface{}) ErrorResponse {
	return ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
			Type:    errorType,
		},
		Timestamp: time.Now(),
	}
}

// WithRequestInfo adds request information to the error response
func (er ErrorResponse) WithRequestInfo(requestID, path, method string) ErrorResponse {
	er.RequestID = requestID
	er.Path = path
	er.Method = method
	return er
}

// ToJSON converts the error response to JSON
func (er ErrorResponse) ToJSON() ([]byte, error) {
	return json.Marshal(er)
}

// WriteToResponse writes the error response to HTTP response
func (er ErrorResponse) WriteToResponse(w http.ResponseWriter, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	jsonData, err := er.ToJSON()
	if err != nil {
		return err
	}

	_, err = w.Write(jsonData)
	return err
}

// Common error response factory functions

// NewValidationErrorResponse creates a validation error response
func NewValidationErrorResponse(message string, details map[string]interface{}) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeValidation,
		message,
		"ValidationError",
		details,
	)
}

// NewNotFoundErrorResponse creates a not found error response
func NewNotFoundErrorResponse(message string) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeNotFound,
		message,
		"NotFoundError",
		nil,
	)
}

// NewUnauthorizedErrorResponse creates an unauthorized error response
func NewUnauthorizedErrorResponse(message string) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeUnauthorized,
		message,
		"UnauthorizedError",
		nil,
	)
}

// NewForbiddenErrorResponse creates a forbidden error response
func NewForbiddenErrorResponse(message string) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeForbidden,
		message,
		"ForbiddenError",
		nil,
	)
}

// NewInternalErrorResponse creates an internal server error response
func NewInternalErrorResponse(message string) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeInternal,
		message,
		"InternalError",
		nil,
	)
}

// NewExternalErrorResponse creates an external service error response
func NewExternalErrorResponse(message string, details map[string]interface{}) ErrorResponse {
	return NewErrorResponse(
		errors.ErrCodeExternal,
		message,
		"ExternalError",
		details,
	)
}

// NewDomainErrorResponse creates a domain error response
func NewDomainErrorResponse(domainErr errors.DomainError) ErrorResponse {
	return NewErrorResponse(
		domainErr.Code,
		domainErr.Message,
		"DomainError",
		domainErr.Details,
	)
}

// NewValidationErrorsResponse creates a validation errors response
func NewValidationErrorsResponse(validationErrs errors.ValidationErrors) ErrorResponse {
	details := make(map[string]interface{})
	errorList := make([]map[string]interface{}, len(validationErrs.Errors))

	for i, err := range validationErrs.Errors {
		errorList[i] = map[string]interface{}{
			"field":   err.Field,
			"value":   err.Value,
			"message": err.Message,
		}
	}

	details["errors"] = errorList
	details["count"] = len(validationErrs.Errors)

	return NewErrorResponse(
		errors.ErrCodeValidation,
		"Validation failed",
		"ValidationErrors",
		details,
	)
}

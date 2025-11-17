package errors

import "fmt"

// DomainError represents a domain-specific error
type DomainError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// NewDomainErrorWithDetails creates a new domain error with details
func NewDomainErrorWithDetails(code, message string, details map[string]interface{}) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Error implements the error interface
func (e DomainError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithDetail adds a detail to the error
func (e DomainError) WithDetail(key string, value interface{}) DomainError {
	e.Details[key] = value
	return e
}

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	_, ok := err.(DomainError)
	return ok
}

// GetDomainErrorCode returns the domain error code if the error is a domain error
func GetDomainErrorCode(err error) (string, bool) {
	if domainErr, ok := err.(DomainError); ok {
		return domainErr.Code, true
	}
	return "", false
}

// Common domain error codes
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeInvalidState  = "INVALID_STATE"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeExternal      = "EXTERNAL_ERROR"
)

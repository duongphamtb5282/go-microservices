package errors

import "fmt"

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Code    string
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    ErrCodeValidation,
	}
}

// Error implements the error interface
func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError
}

// NewValidationErrors creates a new validation errors collection
func NewValidationErrors() ValidationErrors {
	return ValidationErrors{
		Errors: make([]ValidationError, 0),
	}
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field string, value interface{}, message string) {
	ve.Errors = append(ve.Errors, NewValidationError(field, value, message))
}

// HasErrors checks if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}

	message := "validation errors:"
	for _, err := range ve.Errors {
		message += fmt.Sprintf("\n- %s", err.Error())
	}
	return message
}

// ToDomainErrors converts validation errors to domain errors
func (ve ValidationErrors) ToDomainErrors() []DomainError {
	domainErrors := make([]DomainError, len(ve.Errors))
	for i, err := range ve.Errors {
		domainErrors[i] = NewDomainErrorWithDetails(
			ErrCodeValidation,
			err.Message,
			map[string]interface{}{
				"field": err.Field,
				"value": err.Value,
			},
		)
	}
	return domainErrors
}

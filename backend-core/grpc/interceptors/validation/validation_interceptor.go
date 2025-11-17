package validation

import (
	"context"
	"fmt"

	grpcerrors "backend-core/grpc/errors"

	"google.golang.org/grpc"
)

// Validator interface for validating messages
type Validator interface {
	Validate() error
}

// UnaryServerInterceptor returns a new unary server interceptor that validates requests
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Check if request implements Validator interface
		if v, ok := req.(Validator); ok {
			if err := v.Validate(); err != nil {
				return nil, grpcerrors.NewInvalidArgumentError(fmt.Sprintf("validation failed: %v", err))
			}
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor that validates messages
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Wrap the server stream to validate messages
		wrapped := &validatingServerStream{ServerStream: ss}
		return handler(srv, wrapped)
	}
}

// validatingServerStream wraps grpc.ServerStream to add validation
type validatingServerStream struct {
	grpc.ServerStream
}

func (s *validatingServerStream) RecvMsg(m interface{}) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	// Validate the message
	if v, ok := m.(Validator); ok {
		if err := v.Validate(); err != nil {
			return grpcerrors.NewInvalidArgumentError(fmt.Sprintf("validation failed: %v", err))
		}
	}

	return nil
}

// ValidateFields validates common field types
type FieldValidator struct {
	errors []string
}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{
		errors: []string{},
	}
}

// Required checks if a string field is not empty
func (v *FieldValidator) Required(fieldName, value string) *FieldValidator {
	if value == "" {
		v.errors = append(v.errors, fmt.Sprintf("%s is required", fieldName))
	}
	return v
}

// MinLength checks if a string field has minimum length
func (v *FieldValidator) MinLength(fieldName, value string, min int) *FieldValidator {
	if len(value) < min {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at least %d characters", fieldName, min))
	}
	return v
}

// MaxLength checks if a string field has maximum length
func (v *FieldValidator) MaxLength(fieldName, value string, max int) *FieldValidator {
	if len(value) > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at most %d characters", fieldName, max))
	}
	return v
}

// Email checks if a string is a valid email
func (v *FieldValidator) Email(fieldName, value string) *FieldValidator {
	if value != "" && !isValidEmail(value) {
		v.errors = append(v.errors, fmt.Sprintf("%s must be a valid email", fieldName))
	}
	return v
}

// UUID checks if a string is a valid UUID
func (v *FieldValidator) UUID(fieldName, value string) *FieldValidator {
	if value != "" && !isValidUUID(value) {
		v.errors = append(v.errors, fmt.Sprintf("%s must be a valid UUID", fieldName))
	}
	return v
}

// Positive checks if a number is positive
func (v *FieldValidator) Positive(fieldName string, value int64) *FieldValidator {
	if value <= 0 {
		v.errors = append(v.errors, fmt.Sprintf("%s must be positive", fieldName))
	}
	return v
}

// Min checks if a number is at least min
func (v *FieldValidator) Min(fieldName string, value, min int64) *FieldValidator {
	if value < min {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at least %d", fieldName, min))
	}
	return v
}

// Max checks if a number is at most max
func (v *FieldValidator) Max(fieldName string, value, max int64) *FieldValidator {
	if value > max {
		v.errors = append(v.errors, fmt.Sprintf("%s must be at most %d", fieldName, max))
	}
	return v
}

// Error returns the validation error
func (v *FieldValidator) Error() error {
	if len(v.errors) == 0 {
		return nil
	}
	if len(v.errors) == 1 {
		return fmt.Errorf("%s", v.errors[0])
	}
	return fmt.Errorf("validation errors: %v", v.errors)
}

// isValidEmail checks if a string is a valid email (simplified)
func isValidEmail(email string) bool {
	// Simplified email validation
	return len(email) > 3 && containsChar(email, '@') && containsChar(email, '.')
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(uuid string) bool {
	// Simplified UUID validation (format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
	return len(uuid) == 36 && uuid[8] == '-' && uuid[13] == '-' && uuid[18] == '-' && uuid[23] == '-'
}

// containsChar checks if a string contains a character
func containsChar(s string, c rune) bool {
	for _, ch := range s {
		if ch == c {
			return true
		}
	}
	return false
}

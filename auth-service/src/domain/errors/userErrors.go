package errors

import "errors"

// Domain errors for User entity
var (
	ErrUserAlreadyActive     = errors.New("user is already active")
	ErrUserAlreadyInactive   = errors.New("user is already inactive")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserLocked            = errors.New("user account is locked")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrPasswordTooWeak       = errors.New("password is too weak")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
)

// UserError represents a user-specific error
type UserError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *UserError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *UserError) Unwrap() error {
	return e.Err
}

// NewUserError creates a new UserError
func NewUserError(code, message string, err error) *UserError {
	return &UserError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

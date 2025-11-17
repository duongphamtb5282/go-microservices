package valueObjects

import (
	"errors"
	"regexp"
	"strings"
)

// Email represents an email value object
type Email struct {
	value string
}

// NewEmail creates a new Email
func NewEmail(email string) (Email, error) {
	if email == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	// Trim whitespace and convert to lowercase
	email = strings.TrimSpace(strings.ToLower(email))

	// Validate length
	if len(email) > 254 {
		return Email{}, errors.New("email must be at most 254 characters long")
	}

	// Validate format
	if !isValidEmail(email) {
		return Email{}, errors.New("invalid email format")
	}

	return Email{value: email}, nil
}

// String returns the string representation of Email
func (e Email) String() string {
	return e.value
}

// Equals checks if two Emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsValid checks if the email is valid
func (e Email) IsValid() bool {
	return e.value != "" && isValidEmail(e.value)
}

// isValidEmail checks if an email is valid using regex
func isValidEmail(email string) bool {
	// RFC 5322 compliant regex (simplified)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

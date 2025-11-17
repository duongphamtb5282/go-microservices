package valueObjects

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Password represents a password value object
type Password struct {
	hash string
}

// PlainPassword represents a plain text password (for comparison only)
type PlainPassword string

// NewPasswordPlain creates a PlainPassword (use only for verification)
func NewPasswordPlain(plain string) PlainPassword {
	return PlainPassword(plain)
}

// String returns the plain password string
func (p PlainPassword) String() string {
	return string(p)
}

// NewPassword creates a new Password from plain text using bcrypt
func NewPassword(plainPassword string) (Password, error) {
	if plainPassword == "" {
		return Password{}, errors.New("password cannot be empty")
	}

	// Validate password strength
	if err := validatePasswordStrength(plainPassword); err != nil {
		return Password{}, err
	}

	// Generate bcrypt hash (bcrypt includes salt automatically)
	hash, err := hashPassword(plainPassword)
	if err != nil {
		return Password{}, err
	}

	return Password{
		hash: hash,
	}, nil
}

// NewPasswordFromHash creates a Password from existing bcrypt hash
func NewPasswordFromHash(hash string) (Password, error) {
	if hash == "" {
		return Password{}, errors.New("password hash cannot be empty")
	}

	// Verify it's a valid bcrypt hash format
	// Bcrypt hashes start with $2a$, $2b$, or $2y$ and are exactly 60 characters
	if len(hash) != 60 {
		return Password{}, errors.New("invalid bcrypt hash length (must be 60 characters)")
	}

	if hash[0] != '$' || (hash[1] != '2') {
		return Password{}, errors.New("invalid bcrypt hash format (must start with $2)")
	}

	return Password{
		hash: hash,
	}, nil
}

// Hash returns the password hash
func (p Password) Hash() string {
	return p.hash
}

// Verify verifies a plain password against the bcrypt hash
func (p Password) Verify(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainPassword))
	return err == nil
}

// Equals checks if two passwords are equal
func (p Password) Equals(other Password) bool {
	return p.hash == other.hash
}

// validatePasswordStrength validates password strength
func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 100 {
		return errors.New("password must be at most 100 characters long")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(plainPassword string) (string, error) {
	// Use bcrypt with default cost (10)
	// bcrypt automatically generates and includes the salt in the hash
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

package utils

import (
	"regexp"
	"strings"
)

// EmailRegex is a compiled regex for email validation
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// IsValidEmail validates an email address
func IsValidEmail(email string) bool {
	if IsEmpty(email) {
		return false
	}
	return EmailRegex.MatchString(strings.ToLower(email))
}

// IsValidUsername validates a username
func IsValidUsername(username string) bool {
	if IsEmpty(username) {
		return false
	}
	// Username should be 3-20 characters, alphanumeric and underscores only
	if len(username) < 3 || len(username) > 20 {
		return false
	}

	// Check if contains only valid characters
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// IsValidPassword validates a password
func IsValidPassword(password string) bool {
	if IsEmpty(password) {
		return false
	}
	// Password should be at least 8 characters
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase, one lowercase, one digit
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		} else if r >= 'a' && r <= 'z' {
			hasLower = true
		} else if r >= '0' && r <= '9' {
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\t", "")
	return strings.TrimSpace(s)
}

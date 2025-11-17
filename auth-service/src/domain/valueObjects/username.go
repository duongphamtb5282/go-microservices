package valueObjects

import (
	"errors"
	"regexp"
	"strings"
)

// Username represents a username value object
type Username struct {
	value string
}

// NewUsername creates a new Username
func NewUsername(username string) (Username, error) {
	if username == "" {
		return Username{}, errors.New("username cannot be empty")
	}

	// Trim whitespace
	username = strings.TrimSpace(username)

	// Validate length
	if len(username) < 3 {
		return Username{}, errors.New("username must be at least 3 characters long")
	}

	if len(username) > 20 {
		return Username{}, errors.New("username must be at most 20 characters long")
	}

	// Validate format (alphanumeric and underscore only)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if !matched {
		return Username{}, errors.New("username can only contain letters, numbers, and underscores")
	}

	// Cannot start with underscore
	if strings.HasPrefix(username, "_") {
		return Username{}, errors.New("username cannot start with underscore")
	}

	// Cannot end with underscore
	if strings.HasSuffix(username, "_") {
		return Username{}, errors.New("username cannot end with underscore")
	}

	return Username{value: username}, nil
}

// String returns the string representation of Username
func (u Username) String() string {
	return u.value
}

// Equals checks if two Usernames are equal
func (u Username) Equals(other Username) bool {
	return u.value == other.value
}

// IsValid checks if the username is valid
func (u Username) IsValid() bool {
	return u.value != "" && len(u.value) >= 3 && len(u.value) <= 20
}

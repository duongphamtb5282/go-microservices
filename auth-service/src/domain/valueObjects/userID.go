package valueObjects

import (
	"errors"
	"regexp"

	"github.com/google/uuid"
)

// UserID represents a user identifier value object
type UserID struct {
	value string
}

// NewUserID creates a new UserID
func NewUserID() (UserID, error) {
	id := uuid.New().String()
	return UserID{value: id}, nil
}

// NewUserIDFromString creates a UserID from a string
func NewUserIDFromString(id string) (UserID, error) {
	if id == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}

	// Validate UUID format
	if !isValidUUID(id) {
		return UserID{}, errors.New("invalid user ID format")
	}

	return UserID{value: id}, nil
}

// String returns the string representation of UserID
func (u UserID) String() string {
	return u.value
}

// Equals checks if two UserIDs are equal
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(u string) bool {
	matched, _ := regexp.MatchString(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, u)
	return matched
}

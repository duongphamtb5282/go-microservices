package models

import (
	"errors"

	"github.com/google/uuid"
)

// EntityID represents a base entity identifier
type EntityID struct {
	value string
}

// NewEntityID creates a new EntityID
func NewEntityID() EntityID {
	return EntityID{
		value: uuid.New().String(),
	}
}

// NewEntityIDFromString creates an EntityID from a string
func NewEntityIDFromString(id string) (EntityID, error) {
	if id == "" {
		return EntityID{}, errors.New("entity ID cannot be empty")
	}

	// Validate UUID format
	if _, err := uuid.Parse(id); err != nil {
		return EntityID{}, errors.New("invalid entity ID format")
	}

	return EntityID{value: id}, nil
}

// String returns the string representation of the EntityID
func (e EntityID) String() string {
	return e.value
}

// IsEmpty checks if the EntityID is empty
func (e EntityID) IsEmpty() bool {
	return e.value == ""
}

// Equals checks if two EntityIDs are equal
func (e EntityID) Equals(other EntityID) bool {
	return e.value == other.value
}

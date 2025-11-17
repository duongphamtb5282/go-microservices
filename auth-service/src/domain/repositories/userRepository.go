package repositories

import (
	"context"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Save saves a user entity
	Save(ctx context.Context, user *entities.User) error

	// FindByID finds a user by ID
	FindByID(ctx context.Context, id valueObjects.UserID) (*entities.User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error)

	// FindByUsername finds a user by username
	FindByUsername(ctx context.Context, username valueObjects.Username) (*entities.User, error)

	// FindAll finds all users with pagination
	FindAll(ctx context.Context, offset, limit int) ([]*entities.User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)

	// Delete deletes a user by ID
	Delete(ctx context.Context, id valueObjects.UserID) error

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email valueObjects.Email) (bool, error)

	// ExistsByUsername checks if a user exists with the given username
	ExistsByUsername(ctx context.Context, username valueObjects.Username) (bool, error)

	// UpdateLastLogin updates the last login time for a user
	UpdateLastLogin(ctx context.Context, id valueObjects.UserID) error

	// UpdateLoginAttempts updates the login attempts for a user
	UpdateLoginAttempts(ctx context.Context, id valueObjects.UserID, attempts int) error
}

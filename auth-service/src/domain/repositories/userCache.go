package repositories

import (
	"context"

	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
)

// UserCache defines the interface for user caching
type UserCache interface {
	// CacheUser caches a user entity
	CacheUser(ctx context.Context, user *entities.User) error

	// GetUser retrieves a user from cache
	GetUser(ctx context.Context, userID valueObjects.UserID) (*entities.User, error)

	// DeleteUser removes a user from cache
	DeleteUser(ctx context.Context, userID valueObjects.UserID) error

	// CacheUserByEmail caches a user by email
	CacheUserByEmail(ctx context.Context, email valueObjects.Email, user *entities.User) error

	// GetUserByEmail retrieves a user by email from cache
	GetUserByEmail(ctx context.Context, email valueObjects.Email) (*entities.User, error)
}

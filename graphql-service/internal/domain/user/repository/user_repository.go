package repository

import (
	"context"

	"graphql-service/internal/domain/user/entity"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetAll(ctx context.Context, filter map[string]interface{}, pagination map[string]interface{}) ([]*entity.User, error)
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)
}

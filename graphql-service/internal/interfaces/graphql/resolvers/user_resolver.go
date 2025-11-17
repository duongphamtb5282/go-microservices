package resolvers

import (
	"context"
	"fmt"

	"graphql-service/internal/domain/user/entity"
	"graphql-service/internal/domain/user/repository"
)

// UserResolver handles GraphQL user queries and mutations
type UserResolver struct {
	userRepo repository.UserRepository
	logger   interface{} // Replace with actual logger type
}

// NewUserResolver creates a new user resolver
func NewUserResolver(userRepo repository.UserRepository, logger interface{}) *UserResolver {
	return &UserResolver{
		userRepo: userRepo,
		logger:   logger,
	}
}

// User resolves a single user by ID
func (r *UserResolver) User(ctx context.Context, id string) (*entity.User, error) {
	user, err := r.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// Users resolves multiple users with filter and pagination
func (r *UserResolver) Users(ctx context.Context, filter map[string]interface{}, pagination map[string]interface{}) ([]*entity.User, error) {
	// Convert GraphQL filter to repository filter
	repoFilter := make(map[string]interface{})
	if filter != nil {
		for key, value := range filter {
			if value != nil {
				repoFilter[key] = value
			}
		}
	}

	// Convert GraphQL pagination to repository pagination
	repoPagination := make(map[string]interface{})
	if pagination != nil {
		if limit, ok := pagination["limit"].(int); ok {
			repoPagination["limit"] = limit
		}
		if offset, ok := pagination["offset"].(int); ok {
			repoPagination["offset"] = offset
		}
		if page, ok := pagination["page"].(int); ok {
			if limit, ok := pagination["limit"].(int); ok {
				repoPagination["offset"] = (page - 1) * limit
			}
		}
	}

	users, err := r.userRepo.GetAll(ctx, repoFilter, repoPagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	return users, nil
}

// CreateUser creates a new user
func (r *UserResolver) CreateUser(ctx context.Context, input map[string]interface{}) (*entity.User, error) {
	username, _ := input["username"].(string)
	email, _ := input["email"].(string)
	firstName, _ := input["firstName"].(string)
	lastName, _ := input["lastName"].(string)

	user := entity.NewUser(username, email, firstName, lastName)

	err := r.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (r *UserResolver) UpdateUser(ctx context.Context, id string, input map[string]interface{}) (*entity.User, error) {
	user, err := r.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if provided
	if username, ok := input["username"].(string); ok {
		user.Username = username
	}
	if email, ok := input["email"].(string); ok {
		user.Email = email
	}
	if firstName, ok := input["firstName"].(string); ok {
		user.FirstName = firstName
	}
	if lastName, ok := input["lastName"].(string); ok {
		user.LastName = lastName
	}

	err = r.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (r *UserResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
	err := r.userRepo.Delete(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete user: %w", err)
	}

	return true, nil
}

// Orders resolves orders for a user (placeholder - would need order repository)
func (r *UserResolver) Orders(ctx context.Context, user *entity.User) ([]interface{}, error) {
	// This would typically fetch orders from an order repository
	// For now, return empty slice
	return []interface{}{}, nil
}

// Notifications resolves notifications for a user (placeholder - would need notification repository)
func (r *UserResolver) Notifications(ctx context.Context, user *entity.User) ([]interface{}, error) {
	// This would typically fetch notifications from a notification repository
	// For now, return empty slice
	return []interface{}{}, nil
}

// CreatedAt returns the user's creation date as string
func (r *UserResolver) CreatedAt(ctx context.Context, user *entity.User) (string, error) {
	return user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

// UpdatedAt returns the user's update date as string
func (r *UserResolver) UpdatedAt(ctx context.Context, user *entity.User) (string, error) {
	return user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"), nil
}

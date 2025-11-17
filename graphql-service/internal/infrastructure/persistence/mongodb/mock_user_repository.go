package mongodb

import (
	"context"
	"time"

	"graphql-service/internal/domain/user/entity"
	"graphql-service/internal/domain/user/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository implements the user repository interface for testing
type MockUserRepository struct {
	users map[string]*entity.User
	logger interface{} // Replace with actual logger type
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository(logger interface{}) repository.UserRepository {
	return &MockUserRepository{
		users: make(map[string]*entity.User),
		logger: logger,
	}
}

// Create creates a new user
func (r *MockUserRepository) Create(ctx context.Context, user *entity.User) error {
	// Generate a mock ID
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	// Store in mock map
	r.users[user.GetID()] = user
	return nil
}

// GetByID finds a user by ID
func (r *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, nil
	}
	return user, nil
}

// GetByEmail finds a user by email
func (r *MockUserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

// GetByUsername finds a user by username
func (r *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, nil
}

// Update updates a user
func (r *MockUserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()
	r.users[user.GetID()] = user
	return nil
}

// Delete deletes a user by ID
func (r *MockUserRepository) Delete(ctx context.Context, id string) error {
	delete(r.users, id)
	return nil
}

// GetAll gets all users with filter and pagination
func (r *MockUserRepository) GetAll(ctx context.Context, filter map[string]interface{}, pagination map[string]interface{}) ([]*entity.User, error) {
	var users []*entity.User
	
	// Apply filters
	for _, user := range r.users {
		include := true
		
		// Apply username filter
		if username, ok := filter["username"].(string); ok && username != "" {
			if user.Username != username {
				include = false
			}
		}
		
		// Apply email filter
		if email, ok := filter["email"].(string); ok && email != "" {
			if user.Email != email {
				include = false
			}
		}
		
		// Apply firstName filter
		if firstName, ok := filter["firstName"].(string); ok && firstName != "" {
			if user.FirstName != firstName {
				include = false
			}
		}
		
		// Apply lastName filter
		if lastName, ok := filter["lastName"].(string); ok && lastName != "" {
			if user.LastName != lastName {
				include = false
			}
		}
		
		if include {
			users = append(users, user)
		}
	}
	
	// Apply pagination
	if limit, ok := pagination["limit"].(int); ok && limit > 0 {
		if len(users) > limit {
			users = users[:limit]
		}
	}
	
	return users, nil
}

// Count returns the total number of users
func (r *MockUserRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	users, err := r.GetAll(ctx, filter, nil)
	if err != nil {
		return 0, err
	}
	return int64(len(users)), nil
}

// AddSampleData adds sample data for testing
func (r *MockUserRepository) AddSampleData() {
	sampleUsers := []*entity.User{
		entity.NewUser("john_doe", "john@example.com", "John", "Doe"),
		entity.NewUser("jane_smith", "jane@example.com", "Jane", "Smith"),
		entity.NewUser("bob_wilson", "bob@example.com", "Bob", "Wilson"),
	}
	
	for _, user := range sampleUsers {
		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		r.users[user.GetID()] = user
	}
}

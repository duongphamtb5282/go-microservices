package mongodb

import (
	"context"
	"fmt"

	"graphql-service/internal/domain/user/entity"
	"graphql-service/internal/domain/user/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository implements the user repository interface using MongoDB
type UserRepository struct {
	collection *mongo.Collection
	logger     interface{} // Replace with actual logger type
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(collection *mongo.Collection, logger interface{}) repository.UserRepository {
	return &UserRepository{
		collection: collection,
		logger:     logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var user entity.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return &user, nil
}

// GetByEmail finds a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return &user, nil
}

// GetByUsername finds a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	objectID, err := primitive.ObjectIDFromHex(user.GetID())
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetAll gets all users with filter and pagination
func (r *UserRepository) GetAll(ctx context.Context, filter map[string]interface{}, pagination map[string]interface{}) ([]*entity.User, error) {
	// Build filter
	bsonFilter := bson.M{}
	for key, value := range filter {
		if value != nil {
			bsonFilter[key] = value
		}
	}

	// Build options
	opts := options.Find()

	// Apply pagination
	if limit, ok := pagination["limit"].(int); ok && limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset, ok := pagination["offset"].(int); ok && offset > 0 {
		opts.SetSkip(int64(offset))
	}

	// Sort by createdAt descending
	opts.SetSort(bson.D{{"createdAt", -1}})

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	bsonFilter := bson.M{}
	for key, value := range filter {
		if value != nil {
			bsonFilter[key] = value
		}
	}

	count, err := r.collection.CountDocuments(ctx, bsonFilter)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

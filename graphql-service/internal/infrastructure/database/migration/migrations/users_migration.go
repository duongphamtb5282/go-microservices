package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UsersMigration001 creates the users collection with indexes
type UsersMigration001 struct{}

// NewUsersMigration001 creates a new users migration
func NewUsersMigration001() *UsersMigration001 {
	return &UsersMigration001{}
}

// Version returns the migration version
func (m *UsersMigration001) Version() string {
	return "001"
}

// Description returns the migration description
func (m *UsersMigration001) Description() string {
	return "Create users collection with indexes"
}

// Up applies the migration
func (m *UsersMigration001) Up(ctx context.Context, db *mongo.Database) error {
	// Create users collection
	err := db.CreateCollection(ctx, "users")
	if err != nil {
		return fmt.Errorf("failed to create users collection: %w", err)
	}

	// Create indexes
	collection := db.Collection("users")

	// Index on email (unique)
	emailIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"email": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	// Index on username (unique)
	usernameIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"username": 1,
		},
		Options: options.Index().SetUnique(true),
	}

	// Index on created_at
	createdAtIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"created_at": 1,
		},
	}

	// Index on is_active
	activeIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"is_active": 1,
		},
	}

	// Create all indexes
	indexes := []mongo.IndexModel{emailIndex, usernameIndex, createdAtIndex, activeIndex}
	_, err = collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	fmt.Println("Created users collection with indexes")
	return nil
}

// Down rolls back the migration
func (m *UsersMigration001) Down(ctx context.Context, db *mongo.Database) error {
	// Drop the users collection
	err := db.Collection("users").Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop users collection: %w", err)
	}

	fmt.Println("Dropped users collection")
	return nil
}

// Checksum returns a checksum for the migration
func (m *UsersMigration001) Checksum() string {
	return "users_001_checksum"
}

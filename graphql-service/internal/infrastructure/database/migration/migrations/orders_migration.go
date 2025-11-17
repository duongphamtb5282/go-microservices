package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// OrdersMigration002 creates the orders collection with indexes
type OrdersMigration002 struct{}

// NewOrdersMigration002 creates a new orders migration
func NewOrdersMigration002() *OrdersMigration002 {
	return &OrdersMigration002{}
}

// Version returns the migration version
func (m *OrdersMigration002) Version() string {
	return "002"
}

// Description returns the migration description
func (m *OrdersMigration002) Description() string {
	return "Create orders collection with indexes"
}

// Up applies the migration
func (m *OrdersMigration002) Up(ctx context.Context, db *mongo.Database) error {
	// Create orders collection
	err := db.CreateCollection(ctx, "orders")
	if err != nil {
		return fmt.Errorf("failed to create orders collection: %w", err)
	}

	// Create indexes
	collection := db.Collection("orders")

	// Index on user_id
	userIDIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"user_id": 1,
		},
	}

	// Index on status
	statusIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"status": 1,
		},
	}

	// Index on created_at
	createdAtIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"created_at": 1,
		},
	}

	// Index on total_amount
	totalAmountIndex := mongo.IndexModel{
		Keys: map[string]interface{}{
			"total_amount": 1,
		},
	}

	// Create all indexes
	indexes := []mongo.IndexModel{userIDIndex, statusIndex, createdAtIndex, totalAmountIndex}
	_, err = collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	fmt.Println("Created orders collection with indexes")
	return nil
}

// Down rolls back the migration
func (m *OrdersMigration002) Down(ctx context.Context, db *mongo.Database) error {
	// Drop the orders collection
	err := db.Collection("orders").Drop(ctx)
	if err != nil {
		return fmt.Errorf("failed to drop orders collection: %w", err)
	}

	fmt.Println("Dropped orders collection")
	return nil
}

// Checksum returns a checksum for the migration
func (m *OrdersMigration002) Checksum() string {
	return "orders_002_checksum"
}

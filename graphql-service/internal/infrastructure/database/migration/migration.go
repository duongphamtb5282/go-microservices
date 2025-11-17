package migration

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration represents a database migration
type Migration struct {
	ID          string    `bson:"_id"`
	Version     string    `bson:"version"`
	Description string    `bson:"description"`
	AppliedAt   time.Time `bson:"appliedAt"`
	Checksum    string    `bson:"checksum"`
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db     *mongo.Database
	logger interface{} // Replace with actual logger type
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *mongo.Database, logger interface{}) *MigrationRunner {
	return &MigrationRunner{
		db:     db,
		logger: logger,
	}
}

// RunMigrations executes all pending migrations
func (m *MigrationRunner) RunMigrations(ctx context.Context) error {
	log.Println("Starting MongoDB migrations...")

	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure migrations collection: %w", err)
	}

	// Get all migrations
	migrations := m.getMigrations()

	// Apply each migration
	for _, migration := range migrations {
		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// ensureMigrationsCollection creates the migrations collection if it doesn't exist
func (m *MigrationRunner) ensureMigrationsCollection(ctx context.Context) error {
	collections, err := m.db.ListCollectionNames(ctx, bson.M{"name": "migrations"})
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		err = m.db.CreateCollection(ctx, "migrations")
		if err != nil {
			return err
		}
		log.Println("Created migrations collection")
	}

	return nil
}

// getMigrations returns all available migrations
func (m *MigrationRunner) getMigrations() []Migration {
	return []Migration{
		{
			Version:     "001",
			Description: "Create users collection with indexes",
			Checksum:    "users_001",
		},
		{
			Version:     "002",
			Description: "Create orders collection with indexes",
			Checksum:    "orders_002",
		},
		{
			Version:     "003",
			Description: "Create products collection with indexes",
			Checksum:    "products_003",
		},
		{
			Version:     "004",
			Description: "Create notifications collection with indexes",
			Checksum:    "notifications_004",
		},
		{
			Version:     "005",
			Description: "Insert sample data",
			Checksum:    "sample_data_005",
		},
	}
}

// applyMigration applies a single migration
func (m *MigrationRunner) applyMigration(ctx context.Context, migration Migration) error {
	// Check if migration already applied
	applied, err := m.isMigrationApplied(ctx, migration.Version)
	if err != nil {
		return err
	}

	if applied {
		log.Printf("Migration %s already applied, skipping", migration.Version)
		return nil
	}

	log.Printf("Applying migration %s: %s", migration.Version, migration.Description)

	// Apply migration based on version
	switch migration.Version {
	case "001":
		err = m.migration001(ctx)
	case "002":
		err = m.migration002(ctx)
	case "003":
		err = m.migration003(ctx)
	case "004":
		err = m.migration004(ctx)
	case "005":
		err = m.migration005(ctx)
	default:
		return fmt.Errorf("unknown migration version: %s", migration.Version)
	}

	if err != nil {
		return err
	}

	// Record migration as applied
	return m.recordMigration(ctx, migration)
}

// isMigrationApplied checks if a migration has already been applied
func (m *MigrationRunner) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	collection := m.db.Collection("migrations")
	count, err := collection.CountDocuments(ctx, bson.M{"version": version})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// recordMigration records a migration as applied
func (m *MigrationRunner) recordMigration(ctx context.Context, migration Migration) error {
	collection := m.db.Collection("migrations")
	migration.AppliedAt = time.Now()
	_, err := collection.InsertOne(ctx, migration)
	return err
}

// migration001 creates users collection with indexes
func (m *MigrationRunner) migration001(ctx context.Context) error {
	collection := m.db.Collection("users")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"createdAt", 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	log.Println("Created users collection with indexes")
	return nil
}

// migration002 creates orders collection with indexes
func (m *MigrationRunner) migration002(ctx context.Context) error {
	collection := m.db.Collection("orders")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"userId", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"createdAt", 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	log.Println("Created orders collection with indexes")
	return nil
}

// migration003 creates products collection with indexes
func (m *MigrationRunner) migration003(ctx context.Context) error {
	collection := m.db.Collection("products")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"name", 1}},
		},
		{
			Keys: bson.D{{"category", 1}},
		},
		{
			Keys: bson.D{{"price", 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	log.Println("Created products collection with indexes")
	return nil
}

// migration004 creates notifications collection with indexes
func (m *MigrationRunner) migration004(ctx context.Context) error {
	collection := m.db.Collection("notifications")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"userId", 1}},
		},
		{
			Keys: bson.D{{"type", 1}},
		},
		{
			Keys: bson.D{{"read", 1}},
		},
		{
			Keys: bson.D{{"createdAt", 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return err
	}

	log.Println("Created notifications collection with indexes")
	return nil
}

// migration005 inserts sample data
func (m *MigrationRunner) migration005(ctx context.Context) error {
	// Insert sample users
	users := []interface{}{
		bson.M{
			"username":  "john_doe",
			"email":     "john@example.com",
			"firstName": "John",
			"lastName":  "Doe",
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		},
		bson.M{
			"username":  "jane_smith",
			"email":     "jane@example.com",
			"firstName": "Jane",
			"lastName":  "Smith",
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		},
		bson.M{
			"username":  "bob_wilson",
			"email":     "bob@example.com",
			"firstName": "Bob",
			"lastName":  "Wilson",
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	usersCollection := m.db.Collection("users")
	_, err := usersCollection.InsertMany(ctx, users)
	if err != nil {
		return err
	}

	// Insert sample products
	products := []interface{}{
		bson.M{
			"name":        "Laptop",
			"description": "High-performance laptop",
			"price":       999.99,
			"category":    "Electronics",
			"stock":       50,
			"createdAt":   time.Now(),
			"updatedAt":   time.Now(),
		},
		bson.M{
			"name":        "Smartphone",
			"description": "Latest smartphone model",
			"price":       699.99,
			"category":    "Electronics",
			"stock":       100,
			"createdAt":   time.Now(),
			"updatedAt":   time.Now(),
		},
		bson.M{
			"name":        "Book",
			"description": "Programming book",
			"price":       29.99,
			"category":    "Books",
			"stock":       200,
			"createdAt":   time.Now(),
			"updatedAt":   time.Now(),
		},
		bson.M{
			"name":        "Headphones",
			"description": "Wireless headphones",
			"price":       199.99,
			"category":    "Electronics",
			"stock":       75,
			"createdAt":   time.Now(),
			"updatedAt":   time.Now(),
		},
	}

	productsCollection := m.db.Collection("products")
	_, err = productsCollection.InsertMany(ctx, products)
	if err != nil {
		return err
	}

	// Insert sample notifications
	notifications := []interface{}{
		bson.M{
			"userId":    "user_id_1", // This would be replaced with actual user ID
			"type":      "WELCOME",
			"title":     "Welcome to our platform!",
			"message":   "Thank you for joining us. Explore our features and get started.",
			"read":      false,
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		},
		bson.M{
			"userId":    "user_id_2",
			"type":      "PROMOTION",
			"title":     "Special Offer!",
			"message":   "Get 20% off on all electronics. Limited time offer!",
			"read":      false,
			"createdAt": time.Now(),
			"updatedAt": time.Now(),
		},
	}

	notificationsCollection := m.db.Collection("notifications")
	_, err = notificationsCollection.InsertMany(ctx, notifications)
	if err != nil {
		return err
	}

	log.Println("Inserted sample data")
	return nil
}

// GetMigrationStatus returns the status of all migrations
func (m *MigrationRunner) GetMigrationStatus(ctx context.Context) ([]Migration, error) {
	collection := m.db.Collection("migrations")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var migrations []Migration
	if err = cursor.All(ctx, &migrations); err != nil {
		return nil, err
	}

	return migrations, nil
}

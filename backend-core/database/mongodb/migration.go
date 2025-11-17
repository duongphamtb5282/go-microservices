package mongodb

import (
	"context"
	"fmt"
	"time"

	"backend-core/database"
	"backend-core/database/core"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBMigrationManager manages MongoDB migrations
type MongoDBMigrationManager struct {
	database   *mongo.Database
	migrations []core.Migration
}

// NewMongoDBMigrationManager creates a new MongoDB migration manager
func NewMongoDBMigrationManager(database *mongo.Database) *MongoDBMigrationManager {
	return &MongoDBMigrationManager{
		database:   database,
		migrations: make([]core.Migration, 0),
	}
}

// RunMigrations runs all pending migrations
func (m *MongoDBMigrationManager) RunMigrations(ctx context.Context) error {
	// Create migrations collection if it doesn't exist
	if err := m.createMigrationsCollection(ctx); err != nil {
		return fmt.Errorf("failed to create migrations collection: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if !m.isMigrationApplied(appliedMigrations, migration.Version()) {
			if err := m.runMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version(), err)
			}
		}
	}

	return nil
}

// RunMigration runs a specific migration
func (m *MongoDBMigrationManager) RunMigration(ctx context.Context, migration database.Migration) error {
	return m.runMigration(ctx, migration)
}

// RollbackMigration rolls back a specific migration
func (m *MongoDBMigrationManager) RollbackMigration(ctx context.Context, version int) error {
	// Find the migration
	var targetMigration database.Migration
	for _, migration := range m.migrations {
		if migration.Version() == version {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration with version %d not found", version)
	}

	// Check if migration is applied
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if !m.isMigrationApplied(appliedMigrations, version) {
		return fmt.Errorf("migration %d is not applied", version)
	}

	// Rollback the migration
	if err := targetMigration.Down(ctx, m.database); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", version, err)
	}

	// Remove from applied migrations
	if err := m.removeAppliedMigration(ctx, version); err != nil {
		return fmt.Errorf("failed to remove migration %d from applied list: %w", version, err)
	}

	return nil
}

// GetMigrationHistory returns the migration history
func (m *MongoDBMigrationManager) GetMigrationHistory(ctx context.Context) ([]database.Migration, error) {
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var history []database.Migration
	for _, applied := range appliedMigrations {
		// Find the migration in our list
		for _, migration := range m.migrations {
			if migration.Version() == applied.Version {
				history = append(history, migration)
				break
			}
		}
	}

	return history, nil
}

// AddMigration adds a migration to the manager
func (m *MongoDBMigrationManager) AddMigration(migration database.Migration) {
	m.migrations = append(m.migrations, migration)
}

// Helper methods

func (m *MongoDBMigrationManager) createMigrationsCollection(ctx context.Context) error {
	// Create the migrations collection
	collection := m.database.Collection("schema_migrations")

	// Create index on version field
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"version": 1},
		Options: options.Index().SetUnique(true),
	})

	return err
}

func (m *MongoDBMigrationManager) getAppliedMigrations(ctx context.Context) ([]AppliedMigration, error) {
	collection := m.database.Collection("schema_migrations")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var appliedMigrations []AppliedMigration
	err = cursor.All(ctx, &appliedMigrations)
	return appliedMigrations, err
}

func (m *MongoDBMigrationManager) isMigrationApplied(appliedMigrations []AppliedMigration, version int) bool {
	for _, applied := range appliedMigrations {
		if applied.Version == version {
			return true
		}
	}
	return false
}

func (m *MongoDBMigrationManager) runMigration(ctx context.Context, migration database.Migration) error {
	// Start session for transaction
	session, err := m.database.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Run migration in transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Run migration
		if err := migration.Up(sessCtx, m.database); err != nil {
			return nil, err
		}

		// Record migration as applied
		if err := m.recordAppliedMigration(sessCtx, migration); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (m *MongoDBMigrationManager) recordAppliedMigration(ctx context.Context, migration database.Migration) error {
	collection := m.database.Collection("schema_migrations")

	appliedMigration := AppliedMigration{
		Version:     migration.Version(),
		Description: migration.Description(),
		AppliedAt:   time.Now(),
	}

	_, err := collection.InsertOne(ctx, appliedMigration)
	return err
}

func (m *MongoDBMigrationManager) removeAppliedMigration(ctx context.Context, version int) error {
	collection := m.database.Collection("schema_migrations")

	_, err := collection.DeleteOne(ctx, bson.M{"version": version})
	return err
}

// AppliedMigration represents a migration that has been applied
type AppliedMigration struct {
	Version     int       `bson:"version"`
	Description string    `bson:"description"`
	AppliedAt   time.Time `bson:"applied_at"`
}

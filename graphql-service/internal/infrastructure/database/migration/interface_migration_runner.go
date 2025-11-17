package migration

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrationInterface defines the interface for individual migrations
type MigrationInterface interface {
	Version() string
	Description() string
	Up(ctx context.Context, db *mongo.Database) error
	Down(ctx context.Context, db *mongo.Database) error
	Checksum() string
}

// InterfaceMigrationRunner handles interface-based migrations
type InterfaceMigrationRunner struct {
	client     *mongo.Client
	database   *mongo.Database
	migrations []MigrationInterface
}

// NewInterfaceMigrationRunner creates a new interface-based migration runner
func NewInterfaceMigrationRunner(client *mongo.Client, database *mongo.Database, migrations []MigrationInterface) *InterfaceMigrationRunner {
	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version() < migrations[j].Version()
	})

	return &InterfaceMigrationRunner{
		client:     client,
		database:   database,
		migrations: migrations,
	}
}

// RunMigrations runs all pending migrations
func (m *InterfaceMigrationRunner) RunMigrations(ctx context.Context) error {
	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(ctx); err != nil {
		return err
	}

	for _, migration := range m.migrations {
		if err := m.applyInterfaceMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version(), err)
		}
	}

	return nil
}

// applyInterfaceMigration applies a single interface-based migration
func (m *InterfaceMigrationRunner) applyInterfaceMigration(ctx context.Context, migration MigrationInterface) error {
	// Check if migration already applied
	applied, err := m.isMigrationApplied(ctx, migration.Version())
	if err != nil {
		return err
	}

	if applied {
		fmt.Printf("Migration %s already applied, skipping\n", migration.Version())
		return nil
	}

	fmt.Printf("Applying migration %s: %s\n", migration.Version(), migration.Description())

	// Execute the migration
	if err := migration.Up(ctx, m.database); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration as applied
	if err := m.recordMigrationApplied(ctx, migration); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	fmt.Printf("Migration %s applied successfully\n", migration.Version())
	return nil
}

// RollbackMigration rolls back a specific migration
func (m *InterfaceMigrationRunner) RollbackMigration(ctx context.Context, version string) error {
	// Find the migration
	var targetMigration MigrationInterface
	for _, migration := range m.migrations {
		if migration.Version() == version {
			targetMigration = migration
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %s not found", version)
	}

	// Check if migration is applied
	applied, err := m.isMigrationApplied(ctx, version)
	if err != nil {
		return err
	}

	if !applied {
		return fmt.Errorf("migration %s is not applied", version)
	}

	fmt.Printf("Rolling back migration %s: %s\n", targetMigration.Version(), targetMigration.Description())

	// Execute the rollback
	if err := targetMigration.Down(ctx, m.database); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	// Remove migration record
	if err := m.removeMigrationRecord(ctx, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("Migration %s rolled back successfully\n", version)
	return nil
}

// recordMigrationApplied records that a migration has been applied
func (m *InterfaceMigrationRunner) recordMigrationApplied(ctx context.Context, migration MigrationInterface) error {
	collection := m.database.Collection("migrations")

	record := bson.M{
		"version":     migration.Version(),
		"description": migration.Description(),
		"checksum":    migration.Checksum(),
		"applied_at":  time.Now(),
	}

	_, err := collection.InsertOne(ctx, record)
	return err
}

// removeMigrationRecord removes a migration record
func (m *InterfaceMigrationRunner) removeMigrationRecord(ctx context.Context, version string) error {
	collection := m.database.Collection("migrations")

	filter := bson.M{"version": version}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

// isMigrationApplied checks if a migration has been applied
func (m *InterfaceMigrationRunner) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	collection := m.database.Collection("migrations")

	filter := bson.M{"version": version}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ensureMigrationsCollection ensures the migrations collection exists
func (m *InterfaceMigrationRunner) ensureMigrationsCollection(ctx context.Context) error {
	collections, err := m.database.ListCollectionNames(ctx, bson.M{"name": "migrations"})
	if err != nil {
		return err
	}

	if len(collections) == 0 {
		err := m.database.CreateCollection(ctx, "migrations")
		if err != nil {
			return err
		}
		fmt.Println("Created migrations collection")
	}

	return nil
}

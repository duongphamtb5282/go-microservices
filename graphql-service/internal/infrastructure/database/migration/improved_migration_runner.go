package migration

import (
	"context"
	"fmt"
	"os"

	"graphql-service/internal/infrastructure/database/migration/migrations"

	"go.mongodb.org/mongo-driver/mongo"
)

// ImprovedMigrationRunner provides a flexible migration runner
type ImprovedMigrationRunner struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewImprovedMigrationRunner creates a new improved migration runner
func NewImprovedMigrationRunner(client *mongo.Client, database *mongo.Database) *ImprovedMigrationRunner {
	return &ImprovedMigrationRunner{
		client:   client,
		database: database,
	}
}

// RunMigrations runs all pending migrations using the interface-based approach
func (m *ImprovedMigrationRunner) RunMigrations(ctx context.Context) error {
	// Get all migrations from registry
	migrationList := migrations.GetAllMigrations()

	// Convert to local interface type
	localMigrations := make([]MigrationInterface, len(migrationList))
	for i, migration := range migrationList {
		localMigrations[i] = migration
	}

	// Create interface-based migration runner
	runner := NewInterfaceMigrationRunner(m.client, m.database, localMigrations)

	// Run migrations
	return runner.RunMigrations(ctx)
}

// RunMigrationsFromConfig runs migrations from configuration file
func (m *ImprovedMigrationRunner) RunMigrationsFromConfig(ctx context.Context, configFile string) error {
	// Create config-based migration runner
	runner := NewConfigMigrationRunner(m.client, m.database, configFile)

	// Run migrations
	return runner.RunMigrations(ctx)
}

// RunMigrationsFromFiles runs migrations from file system
func (m *ImprovedMigrationRunner) RunMigrationsFromFiles(ctx context.Context, migrationsFS interface{}) error {
	// Create file-based migration runner
	// Note: This would need to be implemented with proper file system handling
	// runner := NewFileMigrationRunner(m.client, m.database, migrationsFS)

	// For now, fall back to interface-based migrations
	return m.RunMigrations(ctx)
}

// AddNewMigration adds a new migration to the registry
func (m *ImprovedMigrationRunner) AddNewMigration(migration MigrationInterface) error {
	// This would typically involve updating the registry
	// For now, we'll just validate the migration
	if migration.Version() == "" {
		return fmt.Errorf("migration version cannot be empty")
	}
	if migration.Description() == "" {
		return fmt.Errorf("migration description cannot be empty")
	}
	if migration.Checksum() == "" {
		return fmt.Errorf("migration checksum cannot be empty")
	}

	fmt.Printf("Migration %s validated successfully\n", migration.Version())
	return nil
}

// RollbackMigration rolls back a specific migration
func (m *ImprovedMigrationRunner) RollbackMigration(ctx context.Context, version string) error {
	// Get the specific migration
	migration := migrations.GetMigrationByVersion(version)
	if migration == nil {
		return fmt.Errorf("migration %s not found", version)
	}

	// Create interface-based migration runner
	runner := NewInterfaceMigrationRunner(m.client, m.database, []MigrationInterface{migration})

	// Rollback the migration
	return runner.RollbackMigration(ctx, version)
}

// GetMigrationStatus returns the status of all migrations
func (m *ImprovedMigrationRunner) GetMigrationStatus(ctx context.Context) (map[string]interface{}, error) {
	// This would check the migrations collection and return status
	collection := m.database.Collection("migrations")

	cursor, err := collection.Find(ctx, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get migration status: %w", err)
	}
	defer cursor.Close(ctx)

	var appliedMigrations []map[string]interface{}
	if err := cursor.All(ctx, &appliedMigrations); err != nil {
		return nil, fmt.Errorf("failed to decode migration status: %w", err)
	}

	status := map[string]interface{}{
		"applied_migrations": appliedMigrations,
		"total_applied":      len(appliedMigrations),
	}

	return status, nil
}

// CreateNewMigrationTemplate creates a template for a new migration
func (m *ImprovedMigrationRunner) CreateNewMigrationTemplate(version, description string) error {
	template := fmt.Sprintf(`package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// %sMigration%s creates the %s
type %sMigration%s struct{}

// New%sMigration%s creates a new %s migration
func New%sMigration%s() *%sMigration%s {
	return &%sMigration%s{}
}

// Version returns the migration version
func (m *%sMigration%s) Version() string {
	return "%s"
}

// Description returns the migration description
func (m *%sMigration%s) Description() string {
	return "%s"
}

// Up applies the migration
func (m *%sMigration%s) Up(ctx context.Context, db *mongo.Database) error {
	// TODO: Implement your migration logic here
	fmt.Println("Applying migration %s: %s")
	return nil
}

// Down rolls back the migration
func (m *%sMigration%s) Down(ctx context.Context, db *mongo.Database) error {
	// TODO: Implement your rollback logic here
	fmt.Println("Rolling back migration %s: %s")
	return nil
}

// Checksum returns a checksum for the migration
func (m *%sMigration%s) Checksum() string {
	return "%s_%s_checksum"
}`,
		description, version, description,
		description, version, description, version, description, version, description, version,
		description, version, version,
		description, version, description,
		description, version, description,
		description, version, version, version)

	filename := fmt.Sprintf("migrations/%s_migration.go", version)
	return os.WriteFile(filename, []byte(template), 0644)
}

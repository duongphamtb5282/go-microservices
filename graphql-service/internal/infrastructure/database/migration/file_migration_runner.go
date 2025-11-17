package migration

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// FileMigrationRunner handles file-based migrations
type FileMigrationRunner struct {
	client     *mongo.Client
	database   *mongo.Database
	migrations fs.FS // File system for migration files
}

// NewFileMigrationRunner creates a new file-based migration runner
func NewFileMigrationRunner(client *mongo.Client, database *mongo.Database, migrations fs.FS) *FileMigrationRunner {
	return &FileMigrationRunner{
		client:     client,
		database:   database,
		migrations: migrations,
	}
}

// FileMigration represents a migration loaded from file
type FileMigration struct {
	Version     string
	Description string
	UpSQL       string
	DownSQL     string
	Checksum    string
}

// getMigrationsFromFiles loads migrations from file system
func (m *FileMigrationRunner) getMigrationsFromFiles() ([]FileMigration, error) {
	var migrations []FileMigration

	// Walk through migration files
	err := fs.WalkDir(m.migrations, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-SQL files
		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		// Parse migration file name (e.g., "001_create_users.sql")
		parts := strings.Split(d.Name(), "_")
		if len(parts) < 2 {
			return fmt.Errorf("invalid migration file name: %s", d.Name())
		}

		version := parts[0]
		description := strings.TrimSuffix(strings.Join(parts[1:], "_"), ".sql")

		// Read migration content
		content, err := fs.ReadFile(m.migrations, path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		// Parse SQL content (up and down sections)
		upSQL, downSQL := m.parseMigrationContent(string(content))

		migration := FileMigration{
			Version:     version,
			Description: description,
			UpSQL:       upSQL,
			DownSQL:     downSQL,
			Checksum:    m.calculateChecksum(string(content)),
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		vi, _ := strconv.Atoi(migrations[i].Version)
		vj, _ := strconv.Atoi(migrations[j].Version)
		return vi < vj
	})

	return migrations, nil
}

// parseMigrationContent parses SQL content into up and down sections
func (m *FileMigrationRunner) parseMigrationContent(content string) (upSQL, downSQL string) {
	lines := strings.Split(content, "\n")
	var upLines, downLines []string
	var inUp, inDown bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "-- +up") {
			inUp = true
			inDown = false
			continue
		}
		if strings.HasPrefix(line, "-- +down") {
			inUp = false
			inDown = true
			continue
		}

		if inUp && !strings.HasPrefix(line, "--") {
			upLines = append(upLines, line)
		}
		if inDown && !strings.HasPrefix(line, "--") {
			downLines = append(downLines, line)
		}
	}

	return strings.Join(upLines, "\n"), strings.Join(downLines, "\n")
}

// calculateChecksum calculates a simple checksum for the migration
func (m *FileMigrationRunner) calculateChecksum(content string) string {
	// Simple checksum implementation
	hash := 0
	for _, char := range content {
		hash = hash*31 + int(char)
	}
	return fmt.Sprintf("%x", hash)
}

// RunMigrations runs all pending migrations
func (m *FileMigrationRunner) RunMigrations(ctx context.Context) error {
	migrations, err := m.getMigrationsFromFiles()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(ctx); err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := m.applyFileMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// applyFileMigration applies a single file-based migration
func (m *FileMigrationRunner) applyFileMigration(ctx context.Context, migration FileMigration) error {
	// Check if migration already applied
	applied, err := m.isMigrationApplied(ctx, migration.Version)
	if err != nil {
		return err
	}

	if applied {
		fmt.Printf("Migration %s already applied, skipping\n", migration.Version)
		return nil
	}

	fmt.Printf("Applying migration %s: %s\n", migration.Version, migration.Description)

	// Execute the migration SQL
	if migration.UpSQL != "" {
		// For MongoDB, we might need to convert SQL to MongoDB operations
		// This is a simplified example
		if err := m.executeMigrationSQL(ctx, migration.UpSQL); err != nil {
			return fmt.Errorf("failed to execute migration SQL: %w", err)
		}
	}

	// Record migration as applied
	if err := m.recordMigrationApplied(ctx, migration); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	fmt.Printf("Migration %s applied successfully\n", migration.Version)
	return nil
}

// executeMigrationSQL executes the migration SQL (simplified for MongoDB)
func (m *FileMigrationRunner) executeMigrationSQL(ctx context.Context, sql string) error {
	// This is a simplified implementation
	// In a real implementation, you would parse SQL and convert to MongoDB operations
	fmt.Printf("Executing SQL: %s\n", sql)
	return nil
}

// recordMigrationApplied records that a migration has been applied
func (m *FileMigrationRunner) recordMigrationApplied(ctx context.Context, migration FileMigration) error {
	collection := m.database.Collection("migrations")

	record := bson.M{
		"version":     migration.Version,
		"description": migration.Description,
		"checksum":    migration.Checksum,
		"applied_at":  time.Now(),
	}

	_, err := collection.InsertOne(ctx, record)
	return err
}

// isMigrationApplied checks if a migration has been applied
func (m *FileMigrationRunner) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	collection := m.database.Collection("migrations")

	filter := bson.M{"version": version}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ensureMigrationsCollection ensures the migrations collection exists
func (m *FileMigrationRunner) ensureMigrationsCollection(ctx context.Context) error {
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

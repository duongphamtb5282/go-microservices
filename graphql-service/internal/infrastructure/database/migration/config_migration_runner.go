package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ConfigMigration represents a migration defined in configuration
type ConfigMigration struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	Checksum    string `json:"checksum"`
	UpScript    string `json:"up_script"`
	DownScript  string `json:"down_script"`
	Enabled     bool   `json:"enabled"`
}

// ConfigMigrationRunner handles configuration-based migrations
type ConfigMigrationRunner struct {
	client     *mongo.Client
	database   *mongo.Database
	configFile string
}

// NewConfigMigrationRunner creates a new configuration-based migration runner
func NewConfigMigrationRunner(client *mongo.Client, database *mongo.Database, configFile string) *ConfigMigrationRunner {
	return &ConfigMigrationRunner{
		client:     client,
		database:   database,
		configFile: configFile,
	}
}

// loadMigrationsFromConfig loads migrations from configuration file
func (m *ConfigMigrationRunner) loadMigrationsFromConfig() ([]ConfigMigration, error) {
	file, err := os.Open(m.configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var migrations []ConfigMigration
	if err := json.Unmarshal(data, &migrations); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Filter enabled migrations
	var enabledMigrations []ConfigMigration
	for _, migration := range migrations {
		if migration.Enabled {
			enabledMigrations = append(enabledMigrations, migration)
		}
	}

	// Sort by version
	sort.Slice(enabledMigrations, func(i, j int) bool {
		return enabledMigrations[i].Version < enabledMigrations[j].Version
	})

	return enabledMigrations, nil
}

// RunMigrations runs all pending migrations from configuration
func (m *ConfigMigrationRunner) RunMigrations(ctx context.Context) error {
	migrations, err := m.loadMigrationsFromConfig()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(ctx); err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := m.applyConfigMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}
	}

	return nil
}

// applyConfigMigration applies a single configuration-based migration
func (m *ConfigMigrationRunner) applyConfigMigration(ctx context.Context, migration ConfigMigration) error {
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

	// Execute the migration script
	if migration.UpScript != "" {
		if err := m.executeMigrationScript(ctx, migration.UpScript); err != nil {
			return fmt.Errorf("failed to execute migration script: %w", err)
		}
	}

	// Record migration as applied
	if err := m.recordMigrationApplied(ctx, migration); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	fmt.Printf("Migration %s applied successfully\n", migration.Version)
	return nil
}

// executeMigrationScript executes a migration script
func (m *ConfigMigrationRunner) executeMigrationScript(ctx context.Context, script string) error {
	// This is a simplified implementation
	// In a real implementation, you would parse and execute the script
	fmt.Printf("Executing script: %s\n", script)
	return nil
}

// recordMigrationApplied records that a migration has been applied
func (m *ConfigMigrationRunner) recordMigrationApplied(ctx context.Context, migration ConfigMigration) error {
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
func (m *ConfigMigrationRunner) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	collection := m.database.Collection("migrations")

	filter := bson.M{"version": version}
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ensureMigrationsCollection ensures the migrations collection exists
func (m *ConfigMigrationRunner) ensureMigrationsCollection(ctx context.Context) error {
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

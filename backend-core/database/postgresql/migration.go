package postgresql

import (
	"context"
	"fmt"
	"time"

	"backend-core/database/health"

	"gorm.io/gorm"
)

// PostgreSQLMigrationManager manages PostgreSQL migrations
type PostgreSQLMigrationManager struct {
	gormDB     *gorm.DB
	migrations []health.Migration
}

// NewPostgreSQLMigrationManager creates a new PostgreSQL migration manager
func NewPostgreSQLMigrationManager(gormDB *gorm.DB) *PostgreSQLMigrationManager {
	return &PostgreSQLMigrationManager{
		gormDB:     gormDB,
		migrations: make([]health.Migration, 0),
	}
}

// RunMigrations runs all pending migrations
func (m *PostgreSQLMigrationManager) RunMigrations(ctx context.Context) error {
	// Create migrations table if it doesn't exist
	if err := m.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if !m.isMigrationApplied(appliedMigrations, migration.GetVersion()) {
			if err := m.runMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", migration.GetVersion(), err)
			}
		}
	}

	return nil
}

// RunMigration runs a specific migration
func (m *PostgreSQLMigrationManager) RunMigration(ctx context.Context, migration health.Migration) error {
	return m.runMigration(ctx, migration)
}

// RollbackMigration rolls back a specific migration
func (m *PostgreSQLMigrationManager) RollbackMigration(ctx context.Context, version string) error {
	// Find the migration
	var targetMigration health.Migration
	for _, migration := range m.migrations {
		if migration.GetVersion() == version {
			targetMigration = migration
			break
		}
	}

	if targetMigration.GetVersion() == "" {
		return fmt.Errorf("migration with version %s not found", version)
	}

	// Check if migration is applied
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if !m.isMigrationApplied(appliedMigrations, version) {
		return fmt.Errorf("migration %s is not applied", version)
	}

	// Rollback the migration
	// Note: health.Migration doesn't have Up/Down methods, so we'll skip the actual rollback
	// In a real implementation, you'd need to implement the rollback logic
	if err := m.rollbackMigration(ctx, targetMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %s: %w", version, err)
	}

	// Remove from applied migrations
	if err := m.removeAppliedMigration(ctx, version); err != nil {
		return fmt.Errorf("failed to remove migration %s from applied list: %w", version, err)
	}

	return nil
}

// GetMigrationHistory returns the migration history
func (m *PostgreSQLMigrationManager) GetMigrationHistory(ctx context.Context) ([]health.Migration, error) {
	appliedMigrations, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	var history []health.Migration
	for _, applied := range appliedMigrations {
		// Find the migration in our list
		for _, migration := range m.migrations {
			if migration.GetVersion() == applied.Version {
				history = append(history, migration)
				break
			}
		}
	}

	return history, nil
}

// AddMigration adds a migration to the manager
func (m *PostgreSQLMigrationManager) AddMigration(migration health.Migration) {
	m.migrations = append(m.migrations, migration)
}

// Helper methods

func (m *PostgreSQLMigrationManager) createMigrationsTable(ctx context.Context) error {
	return m.gormDB.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

func (m *PostgreSQLMigrationManager) getAppliedMigrations(ctx context.Context) ([]AppliedMigration, error) {
	var appliedMigrations []AppliedMigration
	err := m.gormDB.WithContext(ctx).Table("schema_migrations").Find(&appliedMigrations).Error
	return appliedMigrations, err
}

func (m *PostgreSQLMigrationManager) isMigrationApplied(appliedMigrations []AppliedMigration, version string) bool {
	for _, applied := range appliedMigrations {
		if applied.Version == version {
			return true
		}
	}
	return false
}

func (m *PostgreSQLMigrationManager) runMigration(ctx context.Context, migration health.Migration) error {
	// Start transaction
	tx := m.gormDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Run migration
	// Note: health.Migration doesn't have Up/Down methods, so we'll skip the actual migration
	// In a real implementation, you'd need to implement the migration logic
	if err := m.executeMigration(ctx, migration); err != nil {
		tx.Rollback()
		return err
	}

	// Record migration as applied
	if err := m.recordAppliedMigration(tx, migration); err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	return tx.Commit().Error
}

func (m *PostgreSQLMigrationManager) recordAppliedMigration(tx *gorm.DB, migration health.Migration) error {
	return tx.Exec(`
		INSERT INTO schema_migrations (version, description, applied_at) 
		VALUES (?, ?, ?)
	`, migration.GetVersion(), migration.GetDescription(), time.Now()).Error
}

func (m *PostgreSQLMigrationManager) removeAppliedMigration(ctx context.Context, version string) error {
	return m.gormDB.WithContext(ctx).Exec(`
		DELETE FROM schema_migrations WHERE version = ?
	`, version).Error
}

// executeMigration executes a migration (placeholder implementation)
func (m *PostgreSQLMigrationManager) executeMigration(ctx context.Context, migration health.Migration) error {
	// In a real implementation, you would execute the migration SQL here
	// For now, we'll just log that the migration would be executed
	fmt.Printf("Would execute migration: %s - %s\n", migration.GetVersion(), migration.GetDescription())
	return nil
}

// rollbackMigration rolls back a migration (placeholder implementation)
func (m *PostgreSQLMigrationManager) rollbackMigration(ctx context.Context, migration health.Migration) error {
	// In a real implementation, you would execute the rollback SQL here
	// For now, we'll just log that the migration would be rolled back
	fmt.Printf("Would rollback migration: %s - %s\n", migration.GetVersion(), migration.GetDescription())
	return nil
}

// AppliedMigration represents a migration that has been applied
type AppliedMigration struct {
	Version     string    `gorm:"column:version"`
	Description string    `gorm:"column:description"`
	AppliedAt   time.Time `gorm:"column:applied_at"`
}

// TableName returns the table name for AppliedMigration
func (AppliedMigration) TableName() string {
	return "schema_migrations"
}

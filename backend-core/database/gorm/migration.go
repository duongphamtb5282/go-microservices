package gorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Migration implements the Migration interface for GORM
type Migration struct {
	version     int
	description string
	upFunc      func(ctx context.Context, db *gorm.DB) error
	downFunc    func(ctx context.Context, db *gorm.DB) error
}

// NewMigration creates a new migration
func NewMigration(version int, description string, upFunc, downFunc func(ctx context.Context, db *gorm.DB) error) *Migration {
	return &Migration{
		version:     version,
		description: description,
		upFunc:      upFunc,
		downFunc:    downFunc,
	}
}

// Up runs the migration
func (m *Migration) Up(ctx context.Context, db interface{}) error {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return fmt.Errorf("invalid database type for GORM migration")
	}
	return m.upFunc(ctx, gormDB)
}

// Down rolls back the migration
func (m *Migration) Down(ctx context.Context, db interface{}) error {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return fmt.Errorf("invalid database type for GORM migration")
	}
	return m.downFunc(ctx, gormDB)
}

// Version returns the migration version
func (m *Migration) Version() int {
	return m.version
}

// Description returns the migration description
func (m *Migration) Description() string {
	return m.description
}

// MigrationManager manages database migrations
type MigrationManager struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	return &MigrationManager{
		db:         db,
		migrations: []Migration{},
	}
}

// AddMigration adds a migration to the manager
func (m *MigrationManager) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// RunMigrations runs all pending migrations
func (m *MigrationManager) RunMigrations(ctx context.Context) error {
	// Create migrations table if not exists
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
		if !m.isMigrationApplied(appliedMigrations, migration.Version()) {
			if err := migration.Up(ctx, m.db); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version(), err)
			}

			if err := m.recordMigration(ctx, migration); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Version(), err)
			}
		}
	}

	return nil
}

// RollbackMigration rolls back a specific migration
func (m *MigrationManager) RollbackMigration(ctx context.Context, version int) error {
	// Find the migration
	var targetMigration Migration
	for _, migration := range m.migrations {
		if migration.Version() == version {
			targetMigration = migration
			break
		}
	}

	if targetMigration.Version() == 0 {
		return fmt.Errorf("migration version %d not found", version)
	}

	// Run the down migration
	if err := targetMigration.Down(ctx, m.db); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", version, err)
	}

	// Remove from migrations table
	if err := m.removeMigrationRecord(ctx, version); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", version, err)
	}

	return nil
}

// GetMigrationHistory returns the migration history
func (m *MigrationManager) GetMigrationHistory(ctx context.Context) ([]Migration, error) {
	var appliedMigrations []Migration

	// Get applied migration versions
	appliedVersions, err := m.getAppliedMigrations(ctx)
	if err != nil {
		return nil, err
	}

	// Find corresponding migrations
	for _, version := range appliedVersions {
		for _, migration := range m.migrations {
			if migration.Version() == version {
				appliedMigrations = append(appliedMigrations, migration)
				break
			}
		}
	}

	return appliedMigrations, nil
}

// createMigrationsTable creates the migrations table
func (m *MigrationManager) createMigrationsTable(ctx context.Context) error {
	return m.db.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			version INTEGER NOT NULL UNIQUE,
			description TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
}

// getAppliedMigrations returns the list of applied migration versions
func (m *MigrationManager) getAppliedMigrations(ctx context.Context) ([]int, error) {
	var versions []int
	err := m.db.WithContext(ctx).Table("migrations").Select("version").Scan(&versions).Error
	return versions, err
}

// isMigrationApplied checks if a migration has been applied
func (m *MigrationManager) isMigrationApplied(appliedMigrations []int, version int) bool {
	for _, applied := range appliedMigrations {
		if applied == version {
			return true
		}
	}
	return false
}

// recordMigration records a migration as applied
func (m *MigrationManager) recordMigration(ctx context.Context, migration Migration) error {
	return m.db.WithContext(ctx).Exec(`
		INSERT INTO migrations (version, description) VALUES (?, ?)
	`, migration.Version(), migration.Description()).Error
}

// removeMigrationRecord removes a migration record
func (m *MigrationManager) removeMigrationRecord(ctx context.Context, version int) error {
	return m.db.WithContext(ctx).Exec(`
		DELETE FROM migrations WHERE version = ?
	`, version).Error
}

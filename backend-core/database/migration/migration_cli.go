package migration

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"backend-core/logging"

	_ "github.com/lib/pq"
)

// MigrationCLIWrapper provides a simple wrapper for migration CLI operations
type MigrationCLIWrapper struct {
	db     *sql.DB
	logger *logging.Logger
}

// NewMigrationCLIWrapper creates a new migration CLI wrapper
func NewMigrationCLIWrapper(dsn string, migrationsPath string, logger *logging.Logger) (*MigrationCLIWrapper, error) {
	// Initialize database connection
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &MigrationCLIWrapper{
		db:     sqlDB,
		logger: logger,
	}, nil
}

// RunUp runs all pending migrations
func (w *MigrationCLIWrapper) RunUp(migrationsPath string) error {
	w.logger.Info("Running up migrations...")

	// TODO: Fix this - MigrationManager expects gorm.Database but we have *sql.DB
	// For now, return an error indicating this needs to be fixed
	return fmt.Errorf("migration CLI needs to be updated to work with GORM database interface")
}

// RunDown runs down migrations to a specific version
func (w *MigrationCLIWrapper) RunDown(targetVersion int, migrationsPath string) error {
	w.logger.Info("Running down migrations to version", logging.Int("target", targetVersion))

	// TODO: Fix this - MigrationManager expects gorm.Database but we have *sql.DB
	// For now, return an error indicating this needs to be fixed
	return fmt.Errorf("migration CLI needs to be updated to work with GORM database interface")
}

// RunStatus shows migration status
func (w *MigrationCLIWrapper) RunStatus(migrationsPath string) error {
	w.logger.Info("Checking migration status...")

	// TODO: Fix this - MigrationManager expects gorm.Database but we have *sql.DB
	// For now, return an error indicating this needs to be fixed
	return fmt.Errorf("migration CLI needs to be updated to work with GORM database interface")
}

// RunCreate creates a new migration
func (w *MigrationCLIWrapper) RunCreate(name string, migrationsPath string) error {
	if name == "" {
		return fmt.Errorf("migration name is required")
	}

	// Generate next version number
	version, err := w.getNextVersionNumber(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get next version number: %w", err)
	}

	// Create migration file
	filename := fmt.Sprintf("%s/%03d_%s.sql", migrationsPath, version, name)
	content := w.generateMigrationTemplate(name)

	err = os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	w.logger.Info("Migration file created", logging.String("file", filename))
	return nil
}

// RunValidate validates migration files
func (w *MigrationCLIWrapper) RunValidate(migrationsPath string) error {
	w.logger.Info("Validating migrations...")

	// TODO: Fix this - MigrationManager expects gorm.Database but we have *sql.DB
	// For now, return an error indicating this needs to be fixed
	return fmt.Errorf("migration CLI needs to be updated to work with GORM database interface")
}

// getNextVersionNumber generates the next version number
func (w *MigrationCLIWrapper) getNextVersionNumber(migrationsPath string) (int, error) {
	// List existing migration files
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return 1, nil // Start with version 1 if no migrations directory
	}

	maxVersion := 0
	for _, file := range files {
		if file.IsDir() || len(file.Name()) < 3 || file.Name()[0] < '0' || file.Name()[0] > '9' {
			continue
		}

		// Extract version from filename (first 3 characters)
		if len(file.Name()) >= 3 {
			versionStr := file.Name()[:3]
			version, err := strconv.Atoi(versionStr)
			if err == nil && version > maxVersion {
				maxVersion = version
			}
		}
	}

	return maxVersion + 1, nil
}

// generateMigrationTemplate generates a migration file template
func (w *MigrationCLIWrapper) generateMigrationTemplate(name string) string {
	return fmt.Sprintf(`-- Migration: %s
-- Description: Add your migration description here

-- +++++ UP
-- Add your up migration SQL here
-- Example:
-- CREATE TABLE example_table (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT NOW()
-- );

-- +++++ DOWN
-- Add your down migration SQL here
-- Example:
-- DROP TABLE IF EXISTS example_table;
`, name)
}

// ExecuteCommand executes a migration command
func (w *MigrationCLIWrapper) ExecuteCommand(command string, target string, name string, migrationsPath string) error {
	switch command {
	case "up":
		return w.RunUp(migrationsPath)
	case "down":
		if target == "" {
			return fmt.Errorf("target version is required for down migration")
		}
		targetVersion, err := strconv.Atoi(target)
		if err != nil {
			return fmt.Errorf("invalid target version: %w", err)
		}
		return w.RunDown(targetVersion, migrationsPath)
	case "status":
		return w.RunStatus(migrationsPath)
	case "create":
		return w.RunCreate(name, migrationsPath)
	case "validate":
		return w.RunValidate(migrationsPath)
	default:
		return fmt.Errorf("unknown command: %s. Available commands: up, down, status, create, validate", command)
	}
}

// Close closes the database connection
func (w *MigrationCLIWrapper) Close() error {
	return w.db.Close()
}

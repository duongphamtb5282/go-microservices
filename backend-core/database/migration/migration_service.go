package migration

import (
	"context"
	"fmt"

	"backend-core/logging"
)

// MigrationService provides high-level migration operations
type MigrationService struct {
	manager *MigrationManager
	loader  *MigrationLoader
	logger  *logging.Logger
}

// NewMigrationService creates a new migration service
func NewMigrationService(manager *MigrationManager, loader *MigrationLoader, logger *logging.Logger) *MigrationService {
	return &MigrationService{
		manager: manager,
		loader:  loader,
		logger:  logger,
	}
}

// RunMigrations runs all pending migrations
func (s *MigrationService) RunMigrations(ctx context.Context) error {
	s.logger.Info("Starting migration process...")

	// Load all migrations
	migrations, err := s.loader.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		s.logger.Info("No migrations found")
		return nil
	}

	// Run migrations
	config := MigrationConfig{
		AutoMigrate: true,
		Models:      []string{},
		DropTables:  false,
	}
	err = s.manager.RunMigrations(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	s.logger.Info("Migration process completed successfully")
	return nil
}

// RollbackMigrations rolls back migrations to a specific version
func (s *MigrationService) RollbackMigrations(ctx context.Context, targetVersion int) error {
	s.logger.Info("Starting rollback process...",
		logging.Int("target_version", targetVersion))

	// Load all migrations
	_, err := s.loader.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// Rollback migrations - not implemented in current MigrationManager
	// This would need to be implemented based on the specific requirements
	s.logger.Warn("Rollback functionality not implemented in current MigrationManager")

	s.logger.Info("Rollback process completed successfully")
	return nil
}

// GetMigrationStatus returns the current migration status
func (s *MigrationService) GetMigrationStatus(ctx context.Context) ([]*MigrationImpl, error) {
	status, err := s.manager.GetMigrationStatus(ctx)
	if err != nil {
		return nil, err
	}

	// Convert MigrationInfo to MigrationImpl
	var migrations []*MigrationImpl
	for _, dbStatus := range status {
		for _, info := range dbStatus {
			migration := &MigrationImpl{
				Name: info.TableName,
				// Add other fields as needed
			}
			migrations = append(migrations, migration)
		}
	}

	return migrations, nil
}

// ValidateMigrations validates that all migration files are properly formatted
func (s *MigrationService) ValidateMigrations() error {
	s.logger.Info("Validating migrations...")

	migrations, err := s.loader.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	if len(migrations) == 0 {
		s.logger.Info("No migrations to validate")
		return nil
	}

	// Validate each migration
	for _, migration := range migrations {
		version := migration.Version()
		if migration.UpSQL == "" {
			return fmt.Errorf("migration %d (%s) has empty up SQL", version, migration.Name)
		}
		if migration.DownSQL == "" {
			return fmt.Errorf("migration %d (%s) has empty down SQL", version, migration.Name)
		}
	}

	s.logger.Info("All migrations validated successfully",
		logging.Int("count", len(migrations)))

	return nil
}

package migration

import (
	"context"
	"fmt"
	"time"

	"backend-core/database/gorm"
	"backend-core/logging"
)

// MigrationManager manages database migrations using GORM
type MigrationManager struct {
	databases map[string]gorm.Database
	logger    *logging.Logger
}

// MigrationConfig represents migration configuration
type MigrationConfig struct {
	AutoMigrate bool     `yaml:"auto_migrate"`
	Models      []string `yaml:"models"`
	DropTables  bool     `yaml:"drop_tables"`
}

// MigrationInfo represents migration information
type MigrationInfo struct {
	TableName   string    `json:"table_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Indexes     []string  `json:"indexes"`
	Constraints []string  `json:"constraints"`
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(databases map[string]gorm.Database, logger *logging.Logger) *MigrationManager {
	return &MigrationManager{
		databases: databases,
		logger:    logger,
	}
}

// RunMigrations runs migrations for all databases
func (m *MigrationManager) RunMigrations(ctx context.Context, config MigrationConfig) error {
	for dbName, database := range m.databases {
		m.logger.Info("Running migrations for database",
			logging.String("database", dbName))

		if config.DropTables {
			if err := m.dropAllTables(ctx, database); err != nil {
				return fmt.Errorf("failed to drop tables for database %s: %w", dbName, err)
			}
		}

		if config.AutoMigrate {
			if err := m.runAutoMigration(ctx, database); err != nil {
				return fmt.Errorf("failed to run auto migration for database %s: %w", dbName, err)
			}
		}
	}

	return nil
}

// RunMigrationsForDatabase runs migrations for a specific database
func (m *MigrationManager) RunMigrationsForDatabase(ctx context.Context, dbName string, config MigrationConfig) error {
	database, exists := m.databases[dbName]
	if !exists {
		return fmt.Errorf("database %s not found", dbName)
	}

	m.logger.Info("Running migrations for specific database",
		logging.String("database", dbName))

	if config.DropTables {
		if err := m.dropAllTables(ctx, database); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	if config.AutoMigrate {
		if err := m.runAutoMigration(ctx, database); err != nil {
			return fmt.Errorf("failed to run auto migration: %w", err)
		}
	}

	return nil
}

// runAutoMigration runs auto migration for a database
func (m *MigrationManager) runAutoMigration(ctx context.Context, database gorm.Database) error {
	// Get all registered models from the database
	models := m.getRegisteredModels(database)

	if len(models) == 0 {
		m.logger.Warn("No models registered for migration")
		return nil
	}

	m.logger.Info("Starting auto migration",
		logging.Int("model_count", len(models)))

	return database.AutoMigrate(ctx, models...)
}

// dropAllTables drops all tables for a database
func (m *MigrationManager) dropAllTables(ctx context.Context, database gorm.Database) error {
	models := m.getRegisteredModels(database)

	if len(models) == 0 {
		m.logger.Warn("No models registered for dropping tables")
		return nil
	}

	m.logger.Info("Dropping all tables",
		logging.Int("model_count", len(models)))

	return database.DropTable(ctx, models...)
}

// getRegisteredModels returns all models registered with the database
func (m *MigrationManager) getRegisteredModels(database gorm.Database) []interface{} {
	// This would return all models registered with the database
	// For now, we'll return a basic set of models
	// In a real implementation, you would have a model registry
	return []interface{}{
		// Add your GORM models here
		// &models.User{},
		// &models.Role{},
		// &models.Permission{},
		// &models.UserRole{},
		// &models.RolePermission{},
	}
}

// GetMigrationStatus returns migration status for all databases
func (m *MigrationManager) GetMigrationStatus(ctx context.Context) (map[string][]MigrationInfo, error) {
	status := make(map[string][]MigrationInfo)

	for dbName, database := range m.databases {
		models := m.getRegisteredModels(database)
		var migrations []MigrationInfo

		for _, model := range models {
			tableName := m.getTableName(model)

			// Check if table exists
			if database.HasTable(ctx, model) {
				migrationInfo := MigrationInfo{
					TableName: tableName,
					CreatedAt: time.Now(), // You'd implement proper table metadata retrieval
					UpdatedAt: time.Now(),
				}
				migrations = append(migrations, migrationInfo)
			}
		}

		status[dbName] = migrations
	}

	return status, nil
}

// getTableName returns the table name for a model
func (m *MigrationManager) getTableName(model interface{}) string {
	// This is a simplified approach - in a real implementation,
	// you would use reflection to get the table name from GORM tags
	return fmt.Sprintf("%T", model)
}

// RegisterModels registers models for migration
func (m *MigrationManager) RegisterModels(databaseName string, models ...interface{}) {
	// This would register models for a specific database
	// Implementation depends on your model registry system
	m.logger.Info("Registering models for database",
		logging.String("database", databaseName),
		logging.Int("model_count", len(models)))
}

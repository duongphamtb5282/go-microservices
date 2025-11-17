package gorm

import (
	"context"
	"fmt"
	"time"

	"backend-core/config"
	"backend-core/logging"

	"gorm.io/gorm"
)

// MigrationService handles database migration operations
type MigrationService struct {
	config *config.DatabaseConfig
	gormDB *gorm.DB
	logger *logging.Logger
	models []interface{}
}

// NewMigrationService creates a new migration service
func NewMigrationService(config *config.DatabaseConfig, gormDB *gorm.DB, logger *logging.Logger) *MigrationService {
	return &MigrationService{
		config: config,
		gormDB: gormDB,
		logger: logger,
		models: make([]interface{}, 0),
	}
}

// AutoMigrate performs automatic migration
func (m *MigrationService) AutoMigrate(ctx context.Context, models ...interface{}) error {
	m.logger.Info("starting auto migration")

	// Register models for future migrations
	m.models = append(m.models, models...)

	// Perform auto migration
	err := m.gormDB.WithContext(ctx).AutoMigrate(models...)
	if err != nil {
		m.logger.Error("auto migration failed", "error", err)
		return err
	}

	m.logger.Info("auto migration completed successfully")
	return nil
}

// Migrate performs manual migration
func (m *MigrationService) Migrate(ctx context.Context, models ...interface{}) error {
	m.logger.Info("starting manual migration", "database", m.config.Database)

	// Use GORM's Migrator for more control
	migrator := m.gormDB.WithContext(ctx).Migrator()

	for _, model := range models {
		// Check if table exists
		if !migrator.HasTable(model) {
			m.logger.Info("creating table", "model", fmt.Sprintf("%T", model))
			if err := migrator.CreateTable(model); err != nil {
				return err
			}
		} else {
			// Table exists, check for column changes
			m.logger.Info("updating table", "model", fmt.Sprintf("%T", model))
			if err := migrator.AutoMigrate(model); err != nil {
				return err
			}
		}
	}

	m.logger.Info("manual migration completed successfully")
	return nil
}

// DropTable drops tables
func (m *MigrationService) DropTable(ctx context.Context, models ...interface{}) error {
	migrator := m.gormDB.WithContext(ctx).Migrator()

	for _, model := range models {
		if migrator.HasTable(model) {
			m.logger.Info("dropping table", "model", fmt.Sprintf("%T", model))
			if err := migrator.DropTable(model); err != nil {
				return err
			}
		}
	}

	return nil
}

// HasTable checks if a table exists
func (m *MigrationService) HasTable(ctx context.Context, model interface{}) bool {
	migrator := m.gormDB.WithContext(ctx).Migrator()
	return migrator.HasTable(model)
}

// GetMigrationStatus returns migration status
func (m *MigrationService) GetMigrationStatus(ctx context.Context) ([]MigrationInfo, error) {
	var migrations []MigrationInfo

	// Get all tables
	var tables []string
	err := m.gormDB.WithContext(ctx).Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ?", m.config.Database).Scan(&tables).Error
	if err != nil {
		return nil, err
	}

	for _, table := range tables {
		migration := MigrationInfo{
			TableName:   table,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Indexes:     []string{},
			Constraints: []string{},
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// RegisterModels registers models for migration
func (m *MigrationService) RegisterModels(models ...interface{}) {
	m.models = append(m.models, models...)
}

// GetRegisteredModels returns registered models
func (m *MigrationService) GetRegisteredModels() []interface{} {
	return m.models
}

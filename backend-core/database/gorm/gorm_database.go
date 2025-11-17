package gorm

import (
	"context"

	"backend-core/config"
	"backend-core/logging"

	"gorm.io/gorm"
)

// GormDatabase implements the Database interface using GORM
type GormDatabase struct {
	connectionManager *ConnectionManager
	migrationService  *MigrationService
	repositoryFactory *RepositoryFactory
}

// NewGormDatabase creates a new GORM database instance
func NewGormDatabase(config *config.DatabaseConfig, gormDB *gorm.DB, logger *logging.Logger) Database {
	connectionManager := NewConnectionManager(config, gormDB, logger)
	migrationService := NewMigrationService(config, gormDB, logger)
	repositoryFactory := NewRepositoryFactory(config, gormDB, logger)

	return &GormDatabase{
		connectionManager: connectionManager,
		migrationService:  migrationService,
		repositoryFactory: repositoryFactory,
	}
}

// Connection Management
func (g *GormDatabase) Connect(ctx context.Context) error {
	return g.connectionManager.Connect(ctx)
}

func (g *GormDatabase) Disconnect(ctx context.Context) error {
	return g.connectionManager.Disconnect(ctx)
}

func (g *GormDatabase) IsConnected() bool {
	return g.connectionManager.IsConnected()
}

func (g *GormDatabase) HealthCheck(ctx context.Context) error {
	return g.connectionManager.HealthCheck(ctx)
}

// GORM Access
func (g *GormDatabase) GetGormDB() *gorm.DB {
	return g.connectionManager.GetGormDB()
}

// Repository Factory
func (g *GormDatabase) GetRepository(entityType string) interface{} {
	return g.repositoryFactory.GetRepository(entityType)
}

// Migration Management
func (g *GormDatabase) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return g.migrationService.AutoMigrate(ctx, models...)
}

func (g *GormDatabase) Migrate(ctx context.Context, models ...interface{}) error {
	return g.migrationService.Migrate(ctx, models...)
}

func (g *GormDatabase) DropTable(ctx context.Context, models ...interface{}) error {
	return g.migrationService.DropTable(ctx, models...)
}

func (g *GormDatabase) HasTable(ctx context.Context, model interface{}) bool {
	return g.migrationService.HasTable(ctx, model)
}

func (g *GormDatabase) GetMigrationStatus(ctx context.Context) ([]MigrationInfo, error) {
	return g.migrationService.GetMigrationStatus(ctx)
}

// Transaction Management
func (g *GormDatabase) BeginTransaction(ctx context.Context) (Transaction, error) {
	tx := g.GetGormDB().WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &GormTransaction{tx: tx}, nil
}

func (g *GormDatabase) WithTransaction(ctx context.Context, fn func(Transaction) error) error {
	return g.GetGormDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GormTransaction{tx: tx})
	})
}

// Configuration
func (g *GormDatabase) GetConfig() DatabaseConfig {
	config := g.connectionManager.GetConfig()
	return DatabaseConfig{
		DatabaseConfig: *config,
		AliasName:      config.Database, // Use database name as alias
	}
}

func (g *GormDatabase) GetLogger() *logging.Logger {
	return g.connectionManager.GetLogger()
}

// GetStats returns connection statistics
func (g *GormDatabase) GetStats() ConnectionStats {
	return g.connectionManager.GetStats()
}

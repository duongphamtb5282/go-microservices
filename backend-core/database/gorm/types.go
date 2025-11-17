package gorm

import (
	"context"
	"time"

	"backend-core/config"
	"backend-core/logging"

	"gorm.io/gorm"
)

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	config.DatabaseConfig
	AliasName string `yaml:"alias_name"` // "system", "user_db", etc.

	// GORM-specific settings
	Singular bool   `yaml:"singular"` // Disable plural table names
	LogZap   bool   `yaml:"log_zap"`  // Use zap logger
	Engine   string `yaml:"engine"`   // Database engine (MySQL)
}

// Database represents a unified GORM database interface
type Database interface {
	// Connection Management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	HealthCheck(ctx context.Context) error

	// GORM Access
	GetGormDB() *gorm.DB

	// Repository Factory
	GetRepository(entityType string) interface{}

	// Migration Management
	AutoMigrate(ctx context.Context, models ...interface{}) error
	Migrate(ctx context.Context, models ...interface{}) error
	DropTable(ctx context.Context, models ...interface{}) error
	HasTable(ctx context.Context, model interface{}) bool
	GetMigrationStatus(ctx context.Context) ([]MigrationInfo, error)

	// Transaction Management
	BeginTransaction(ctx context.Context) (Transaction, error)
	WithTransaction(ctx context.Context, fn func(Transaction) error) error

	// Configuration
	GetConfig() DatabaseConfig
	GetLogger() *logging.Logger
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
	GetGormDB() *gorm.DB
}

// ConnectionStats represents database connection statistics
type ConnectionStats struct {
	OpenConnections   int
	IdleConnections   int
	InUseConnections  int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxIdleTimeClosed int64
	MaxLifetimeClosed int64
}

// MigrationInfo represents migration information
type MigrationInfo struct {
	TableName   string    `json:"table_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Indexes     []string  `json:"indexes"`
	Constraints []string  `json:"constraints"`
}

// GormTransaction implements the Transaction interface
type GormTransaction struct {
	tx *gorm.DB
}

// NewGormTransaction creates a new GormTransaction
func NewGormTransaction(tx *gorm.DB) *GormTransaction {
	return &GormTransaction{tx: tx}
}

// Commit commits the transaction
func (t *GormTransaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *GormTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// GetGormDB returns the GORM database instance
func (t *GormTransaction) GetGormDB() *gorm.DB {
	return t.tx
}

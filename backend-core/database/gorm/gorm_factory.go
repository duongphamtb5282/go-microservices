package gorm

import (
	"fmt"
	"strings"
	"time"

	"backend-core/logging"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// GormDatabaseFactory creates and manages GORM database connections
type GormDatabaseFactory struct {
	configs   map[string]DatabaseConfig
	databases map[string]Database
	logger    *logging.Logger
}

// NewGormDatabaseFactory creates a new GORM database factory
func NewGormDatabaseFactory(logger *logging.Logger) *GormDatabaseFactory {
	return &GormDatabaseFactory{
		configs:   make(map[string]DatabaseConfig),
		databases: make(map[string]Database),
		logger:    logger,
	}
}

// CreateDatabase creates a database connection based on configuration
func (f *GormDatabaseFactory) CreateDatabase(config DatabaseConfig) (Database, error) {
	if config.Database == "" {
		return nil, fmt.Errorf("database %s is not configured", config.AliasName)
	}

	switch config.Type {
	case "postgresql":
		return f.createPostgreSQLDatabase(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// createPostgreSQLDatabase creates a PostgreSQL database connection
func (f *GormDatabaseFactory) createPostgreSQLDatabase(config DatabaseConfig) (Database, error) {
	// Build DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	// Create GORM configuration
	gormConfig := f.createGormConfig(config)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	f.logger.Info("PostgreSQL database connected successfully, database: %s, host: %s, port: %d",
		config.AliasName, config.Host, config.Port)

	// Extract the embedded DatabaseConfig
	return NewGormDatabase(&config.DatabaseConfig, db, f.logger), nil
}

// createGormConfig creates GORM configuration
func (f *GormDatabaseFactory) createGormConfig(config DatabaseConfig) *gorm.Config {
	// Create logger
	var gormLogger logger.Interface
	if config.LogZap {
		gormLogger = logger.Default.LogMode(f.getLogLevel("info"))
	} else {
		gormLogger = logger.Default.LogMode(f.getLogLevel("info"))
	}

	// Create naming strategy
	namingStrategy := schema.NamingStrategy{
		SingularTable: config.Singular,
	}

	return &gorm.Config{
		Logger:         gormLogger,
		NamingStrategy: namingStrategy,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// getLogLevel converts string log level to GORM log level
func (f *GormDatabaseFactory) getLogLevel(logMode string) logger.LogLevel {
	switch strings.ToLower(logMode) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}

// GetDatabase returns a database by alias name
func (f *GormDatabaseFactory) GetDatabase(aliasName string) (Database, error) {
	db, exists := f.databases[aliasName]
	if !exists {
		return nil, fmt.Errorf("database %s not found", aliasName)
	}
	return db, nil
}

// GetAllDatabases returns all databases
func (f *GormDatabaseFactory) GetAllDatabases() map[string]Database {
	return f.databases
}

// RegisterDatabase registers a database
func (f *GormDatabaseFactory) RegisterDatabase(aliasName string, database Database) {
	f.databases[aliasName] = database
}

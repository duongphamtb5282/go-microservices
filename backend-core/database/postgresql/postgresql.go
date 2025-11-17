package postgresql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend-core/config"
	"backend-core/database/gorm"
	"backend-core/database/health"
	"backend-core/database/interfaces"
	"backend-core/logging"

	"gorm.io/driver/postgres"
	gormDB "gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgreSQLDatabase implements the ExtendedDatabase interface for PostgreSQL
type PostgreSQLDatabase struct {
	config             *config.DatabaseConfig
	logger             *logging.Logger
	gormDB             *gormDB.DB
	sqlDB              *sql.DB
	queryBuilder       *PostgreSQLQueryBuilder
	migrationManager   *PostgreSQLMigrationManager
	transactionManager interface{}
	healthChecker      *PostgreSQLHealthChecker
	monitor            *PostgreSQLMonitor
	isConnected        bool
	connectedAt        time.Time
}

// NewPostgreSQLDatabase creates a new PostgreSQL database instance
func NewPostgreSQLDatabase(cfg *config.DatabaseConfig) *PostgreSQLDatabase {
	logger, _ := logging.NewLogger(&config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	return &PostgreSQLDatabase{
		config: cfg,
		logger: logger,
	}
}

// LogConnection logs connection events
func (p *PostgreSQLDatabase) LogConnection(event string, err error) {
	if err != nil {
		p.logger.Error("PostgreSQL connection event", "event", event, "error", err)
	} else {
		p.logger.Info("PostgreSQL connection event", "event", event)
	}
}

// SetDriver sets the database driver
func (p *PostgreSQLDatabase) SetDriver(driver interface{}) {
	// Implementation for setting driver
}

// SetConnectionPool sets the connection pool
func (p *PostgreSQLDatabase) SetConnectionPool(pool interface{}) {
	// Implementation for setting connection pool
}

// AutoMigrate performs automatic migration
func (p *PostgreSQLDatabase) AutoMigrate(ctx context.Context, models ...interface{}) error {
	return p.gormDB.WithContext(ctx).AutoMigrate(models...)
}

// Migrate performs manual migration
func (p *PostgreSQLDatabase) Migrate(ctx context.Context, models ...interface{}) error {
	migrator := p.gormDB.WithContext(ctx).Migrator()
	for _, model := range models {
		if !migrator.HasTable(model) {
			if err := migrator.CreateTable(model); err != nil {
				return err
			}
		} else {
			if err := migrator.AutoMigrate(model); err != nil {
				return err
			}
		}
	}
	return nil
}

// DropTable drops tables
func (p *PostgreSQLDatabase) DropTable(ctx context.Context, models ...interface{}) error {
	migrator := p.gormDB.WithContext(ctx).Migrator()
	for _, model := range models {
		if migrator.HasTable(model) {
			if err := migrator.DropTable(model); err != nil {
				return err
			}
		}
	}
	return nil
}

// HasTable checks if a table exists
func (p *PostgreSQLDatabase) HasTable(ctx context.Context, model interface{}) bool {
	return p.gormDB.WithContext(ctx).Migrator().HasTable(model)
}

// GetMigrationStatus returns migration status
func (p *PostgreSQLDatabase) GetMigrationStatus(ctx context.Context) ([]gorm.MigrationInfo, error) {
	// Simplified implementation
	return []gorm.MigrationInfo{}, nil
}

// BeginTransaction starts a new transaction
func (p *PostgreSQLDatabase) BeginTransaction(ctx context.Context) (gorm.Transaction, error) {
	tx := p.gormDB.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return gorm.NewGormTransaction(tx), nil
}

// WithTransaction executes a function within a transaction
func (p *PostgreSQLDatabase) WithTransaction(ctx context.Context, fn func(gorm.Transaction) error) error {
	return p.gormDB.WithContext(ctx).Transaction(func(tx *gormDB.DB) error {
		return fn(gorm.NewGormTransaction(tx))
	})
}

// GetRepository creates a repository (for gorm.Database interface)
func (p *PostgreSQLDatabase) GetRepository(entityType string) interface{} {
	return NewPostgreSQLRepository[interface{}](p, entityType, p.logger)
}

// GetRepositoryDefault creates a repository (parameterless version for Database interface)
func (p *PostgreSQLDatabase) GetRepositoryDefault() interface{} {
	return NewPostgreSQLRepository[interface{}](p, "default", p.logger)
}

// DatabaseWrapper wraps PostgreSQLDatabase to implement the main Database interface
type DatabaseWrapper struct {
	*PostgreSQLDatabase
}

// GetRepository implements the main Database interface
func (w *DatabaseWrapper) GetRepository() interface{} {
	return w.GetRepositoryDefault()
}

// WithTransaction implements the main Database interface
func (w *DatabaseWrapper) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return w.PostgreSQLDatabase.WithTransaction(ctx, func(tx gorm.Transaction) error {
		return fn(ctx)
	})
}

// GetLogger returns the logger
func (p *PostgreSQLDatabase) GetLogger() *logging.Logger {
	return p.logger
}

// GetGormDB returns the GORM database instance
func (p *PostgreSQLDatabase) GetGormDB() *gormDB.DB {
	return p.gormDB
}

// HealthCheck performs a health check on the database
func (p *PostgreSQLDatabase) HealthCheck(ctx context.Context) error {
	if p.healthChecker == nil {
		return fmt.Errorf("health checker not initialized")
	}
	return p.healthChecker.CheckHealth(ctx)
}

// GetConfig returns the database configuration
func (p *PostgreSQLDatabase) GetConfig() gorm.DatabaseConfig {
	// Convert config.DatabaseConfig to gorm.DatabaseConfig
	return gorm.DatabaseConfig{
		DatabaseConfig: *p.config,
		AliasName:      "postgresql",
		Singular:       true,
		LogZap:         true,
	}
}

// IsConnected returns the connection status
func (p *PostgreSQLDatabase) IsConnected() bool {
	return p.isConnected
}

// PostgreSQLTransaction implements the Transaction interface
type PostgreSQLTransaction struct {
	tx *gormDB.DB
}

// Commit commits the transaction
func (t *PostgreSQLTransaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *PostgreSQLTransaction) Rollback() error {
	return t.tx.Rollback().Error
}

// GetGormDB returns the GORM database instance
func (t *PostgreSQLTransaction) GetGormDB() *gormDB.DB {
	return t.tx
}

// SetConnected sets the connection status
func (p *PostgreSQLDatabase) SetConnected(connected bool) {
	p.isConnected = connected
	if connected {
		p.connectedAt = time.Now()
	}
}

// GetConnectedAt returns when the database was connected
func (p *PostgreSQLDatabase) GetConnectedAt() time.Time {
	return p.connectedAt
}

// Connect establishes connection to PostgreSQL
func (p *PostgreSQLDatabase) Connect(ctx context.Context) error {
	// Use DSN provider for connection string generation
	cfg := p.GetConfig()
	dsnProvider := config.NewDSNProvider(cfg.Type, &cfg.DatabaseConfig)
	if dsnProvider == nil {
		return fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	dsn := dsnProvider.Dsn()
	p.LogConnection("connecting", nil)

	// Configure GORM
	gormConfig := &gormDB.Config{
		Logger: logger.Default.LogMode(getLogLevel(p.GetConfig().LogLevel)),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	gormDB, err := gormDB.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		p.LogConnection("connection_failed", err)
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying sql.DB for connection pool management
	sqlDB, err := gormDB.DB()
	if err != nil {
		p.LogConnection("pool_creation_failed", err)
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(p.GetConfig().MaxOpenConns)
	sqlDB.SetMaxIdleConns(p.GetConfig().MaxIdleConns)
	sqlDB.SetConnMaxLifetime(p.GetConfig().ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(p.GetConfig().ConnMaxIdleTime)

	// Test connection
	if err := sqlDB.PingContext(ctx); err != nil {
		p.LogConnection("ping_failed", err)
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Set database components
	p.gormDB = gormDB
	p.sqlDB = sqlDB
	p.SetDriver(gormDB)
	p.SetConnectionPool(sqlDB)
	p.SetConnected(true)

	// Initialize components
	p.queryBuilder = NewPostgreSQLQueryBuilder(gormDB)
	p.migrationManager = NewPostgreSQLMigrationManager(gormDB)
	// Transaction manager will be created when needed
	p.transactionManager = nil
	cfg = p.GetConfig()
	p.healthChecker = NewPostgreSQLHealthChecker(sqlDB, &cfg.DatabaseConfig)
	p.monitor = NewPostgreSQLMonitor()

	p.LogConnection("connected", nil)
	return nil
}

// Disconnect closes the database connection
func (p *PostgreSQLDatabase) Disconnect(ctx context.Context) error {
	if p.sqlDB != nil {
		p.LogConnection("disconnecting", nil)
		if err := p.sqlDB.Close(); err != nil {
			p.LogConnection("disconnect_failed", err)
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}
		p.SetConnected(false)
		p.LogConnection("disconnected", nil)
	}
	return nil
}

// Ping tests the database connection
func (p *PostgreSQLDatabase) Ping(ctx context.Context) error {
	// Connection validation skipped for now

	if err := p.sqlDB.PingContext(ctx); err != nil {
		p.LogConnection("ping_failed", err)
		return fmt.Errorf("PostgreSQL ping failed: %w", err)
	}

	return nil
}

// IsHealthy checks if the database is healthy
func (p *PostgreSQLDatabase) IsHealthy(ctx context.Context) bool {
	if p.healthChecker != nil {
		return p.healthChecker.GetStatus() == health.HealthStatusHealthy
	}

	// Fallback to ping
	if err := p.Ping(ctx); err != nil {
		return false
	}

	return true
}

// RunMigrations runs database migrations
func (p *PostgreSQLDatabase) RunMigrations(ctx context.Context) error {
	if p.migrationManager != nil {
		return p.migrationManager.RunMigrations(ctx)
	}
	return fmt.Errorf("migration manager not initialized")
}

// RollbackMigration rolls back a specific migration
func (p *PostgreSQLDatabase) RollbackMigration(ctx context.Context, version string) error {
	if p.migrationManager != nil {
		return p.migrationManager.RollbackMigration(ctx, version)
	}
	return fmt.Errorf("migration manager not initialized")
}

// GetStats returns connection statistics
func (p *PostgreSQLDatabase) GetStats() interface{} {
	if p.sqlDB == nil {
		return health.NewConnectionStats(0, 0)
	}

	stats := p.sqlDB.Stats()
	return health.NewConnectionStats(
		int64(stats.OpenConnections),
		0, // ConnectionFailures would need to be tracked separately
	)
}

// SetMaxOpenConns sets the maximum number of open connections
func (p *PostgreSQLDatabase) SetMaxOpenConns(max int) {
	if p.sqlDB != nil {
		p.sqlDB.SetMaxOpenConns(max)
	}
}

// SetMaxIdleConns sets the maximum number of idle connections
func (p *PostgreSQLDatabase) SetMaxIdleConns(max int) {
	if p.sqlDB != nil {
		p.sqlDB.SetMaxIdleConns(max)
	}
}

// SetConnMaxLifetime sets the maximum lifetime of connections
func (p *PostgreSQLDatabase) SetConnMaxLifetime(d time.Duration) {
	if p.sqlDB != nil {
		p.sqlDB.SetConnMaxLifetime(d)
	}
}

// ExtendedDatabase interface methods

// GetDriver returns the GORM database instance
func (p *PostgreSQLDatabase) GetDriver() interface{} {
	return p.gormDB
}

// GetConnectionPool returns the underlying sql.DB
func (p *PostgreSQLDatabase) GetConnectionPool() interface{} {
	return p.sqlDB
}

// GetQueryBuilder returns the PostgreSQL query builder
func (p *PostgreSQLDatabase) GetQueryBuilder() interface{} {
	return p.queryBuilder
}

// GetMigrationManager returns the PostgreSQL migration manager
func (p *PostgreSQLDatabase) GetMigrationManager() interface{} {
	return p.migrationManager
}

// GetTransactionManager returns the PostgreSQL transaction manager
func (p *PostgreSQLDatabase) GetTransactionManager() interfaces.TransactionManager {
	// Return a simple transaction manager implementation
	return &SimpleTransactionManager{db: p.sqlDB}
}

// GetHealthChecker returns the PostgreSQL health checker
func (p *PostgreSQLDatabase) GetHealthChecker() interfaces.DatabaseHealthChecker {
	return p.healthChecker
}

// GetMonitor returns the PostgreSQL monitor
func (p *PostgreSQLDatabase) GetMonitor() interfaces.DatabaseMonitor {
	return p.monitor
}

// Helper functions

func getLogLevel(level string) logger.LogLevel {
	switch level {
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

func getTableName[T any]() string {
	// This is a simplified implementation
	// In a real implementation, you might use reflection or struct tags
	var t T
	return fmt.Sprintf("%T", t)
}

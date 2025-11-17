package core

import (
	"context"
	"time"

	"backend-core/database/health"
)

// DatabaseType represents supported database types
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
	MongoDB    DatabaseType = "mongodb"
)

// Database defines the interface for database operations
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Ping(ctx context.Context) error
	IsHealthy(ctx context.Context) bool

	// Repository factory
	GetRepository() interface{}

	// Transaction management
	WithTransaction(ctx context.Context, fn func(context.Context) error) error

	// Migration management
	RunMigrations(ctx context.Context) error
	RollbackMigration(ctx context.Context, version string) error

	// Connection pooling
	GetStats() interface{}
	SetMaxOpenConns(max int)
	SetMaxIdleConns(max int)
	SetConnMaxLifetime(d time.Duration)
}

// ExtendedDatabase extends Database with additional capabilities
type ExtendedDatabase interface {
	Database

	// Database-specific methods
	GetDriver() interface{}                    // Get underlying driver
	GetConnectionPool() interface{}            // Get connection pool
	GetQueryBuilder() QueryBuilder             // Get query builder
	GetMigrationManager() MigrationManager     // Get migration manager
	GetTransactionManager() TransactionManager // Get transaction manager
	GetHealthChecker() DatabaseHealthChecker   // Get health checker
	GetMonitor() DatabaseMonitor               // Get database monitor
}

// QueryBuilder helps build complex queries
type QueryBuilder interface {
	Where(field string, operator string, value interface{}) QueryBuilder
	And(conditions map[string]interface{}) QueryBuilder
	Or(conditions []map[string]interface{}) QueryBuilder
	Sort(field string, order string) QueryBuilder
	Limit(limit int) QueryBuilder
	Offset(offset int) QueryBuilder
	Build() (interface{}, error)
}

// MigrationManager manages database migrations
type MigrationManager interface {
	RunMigrations(ctx context.Context) error
	RollbackMigration(ctx context.Context, version int) error
	GetMigrationHistory(ctx context.Context) ([]Migration, error)
	AddMigration(migration Migration)
}

// Migration defines the interface for database migrations
type Migration interface {
	Up(ctx context.Context, db interface{}) error
	Down(ctx context.Context, db interface{}) error
	Version() int
	Description() string
}

// TransactionManager defines the interface for database transactions
type TransactionManager interface {
	// Transaction execution
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	BeginTransaction(ctx context.Context) (Transaction, error)

	// Transaction status
	IsInTransaction(ctx context.Context) bool
	GetTransactionLevel(ctx context.Context) int
}

// Transaction represents a database transaction
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	IsActive() bool
	GetID() string
}

// DatabaseHealthChecker defines the interface for database health checks
type DatabaseHealthChecker interface {
	// Health checks
	CheckHealth(ctx context.Context) error
	CheckConnection(ctx context.Context) error
	CheckPerformance(ctx context.Context) (health.MetricsResult, error)

	// Status information
	GetStatus() health.HealthStatus
	GetLastHealthCheck() time.Time
	GetHealthHistory() []health.HealthResult
}

// DatabaseMonitor defines the interface for monitoring database operations
type DatabaseMonitor interface {
	// Monitoring
	StartMonitoring(ctx context.Context) error
	StopMonitoring(ctx context.Context) error
	IsMonitoring() bool

	// Metrics collection
	RecordQuery(query string, duration time.Duration, err error)
	RecordConnection(connected bool, err error)
	RecordTransaction(operation string, duration time.Duration, err error)

	// Statistics
	GetQueryStats() health.QueryStats
	GetConnectionStats() health.ConnectionStats
	GetTransactionStats() health.TransactionStats
	GetOverallStats() health.StatsResult
}

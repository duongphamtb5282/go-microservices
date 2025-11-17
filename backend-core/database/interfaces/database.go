package interfaces

import (
	"context"

	"backend-core/database/health"
)

// Database defines the core database interface
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Ping(ctx context.Context) error

	// Health and monitoring
	GetHealthStatus() health.HealthStatus
	GetStats() health.ConnectionStats
	GetMonitor() DatabaseMonitor
	GetHealthChecker() DatabaseHealthChecker

	// Repository and query management
	GetRepository() interface{}
	GetQueryBuilder() QueryBuilder
	GetMigrationManager() MigrationManager
	GetTransactionManager() TransactionManager

	// Utility methods
	GetLogger() interface{} // Return logger interface
}

// ExtendedDatabase extends the basic Database interface
type ExtendedDatabase interface {
	Database
	// Additional methods can be added here
}

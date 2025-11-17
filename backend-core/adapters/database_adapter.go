package adapters

import (
	"context"
	"fmt"
	"time"

	"backend-core/database"
	"backend-core/database/gorm"
	"backend-core/database/health"
)

// DatabaseAdapter adapts gorm.Database to database.Database interface
// This adapter can be reused across all services that need to bridge
// GORM database implementations with the core.Database interface
type DatabaseAdapter struct {
	db gorm.Database
}

// NewDatabaseAdapter creates a new database adapter
func NewDatabaseAdapter(db gorm.Database) *DatabaseAdapter {
	return &DatabaseAdapter{db: db}
}

// GetRepository returns the repository for the given entity type
func (d *DatabaseAdapter) GetRepository() interface{} {
	return d.db.GetRepository("")
}

// Connect establishes a connection to the database
func (d *DatabaseAdapter) Connect(ctx context.Context) error {
	return d.db.Connect(ctx)
}

// Disconnect closes the database connection
func (d *DatabaseAdapter) Disconnect(ctx context.Context) error {
	return d.db.Disconnect(ctx)
}

// IsConnected returns true if the database is connected
func (d *DatabaseAdapter) IsConnected() bool {
	return d.db.IsConnected()
}

// IsHealthy checks if the database is healthy
func (d *DatabaseAdapter) IsHealthy(ctx context.Context) bool {
	err := d.db.HealthCheck(ctx)
	return err == nil
}

// Ping tests the database connection
func (d *DatabaseAdapter) Ping(ctx context.Context) error {
	return d.db.HealthCheck(ctx)
}

// RollbackMigration rolls back a database migration
func (d *DatabaseAdapter) RollbackMigration(ctx context.Context, version string) error {
	// GORM doesn't have rollback migration, return not implemented
	return fmt.Errorf("rollback migration not implemented for GORM")
}

// RunMigrations runs database migrations
func (d *DatabaseAdapter) RunMigrations(ctx context.Context) error {
	// GORM auto-migration is handled during initialization
	return nil
}

// SetConnMaxLifetime sets the maximum lifetime of connections
func (d *DatabaseAdapter) SetConnMaxLifetime(duration time.Duration) {
	// GORM connection lifetime is set during initialization
	// This is a no-op for the adapter
}

// SetMaxIdleConns sets the maximum number of idle connections
func (d *DatabaseAdapter) SetMaxIdleConns(n int) {
	// GORM max idle connections are set during initialization
	// This is a no-op for the adapter
}

// SetMaxOpenConns sets the maximum number of open connections
func (d *DatabaseAdapter) SetMaxOpenConns(n int) {
	// GORM max open connections are set during initialization
	// This is a no-op for the adapter
}

// WithTransaction executes a function within a database transaction
func (d *DatabaseAdapter) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	// Use GORM's transaction support
	return d.db.WithTransaction(ctx, func(tx gorm.Transaction) error {
		return fn(ctx)
	})
}

// HealthCheck returns the health status of the database
func (d *DatabaseAdapter) HealthCheck() database.HealthCheck {
	// Create a simple health check
	ctx := context.Background()
	err := d.db.HealthCheck(ctx)
	if err != nil {
		return *health.NewHealthCheck(health.HealthStatusUnhealthy, 0, err, nil)
	}
	return *health.NewHealthCheck(health.HealthStatusHealthy, 0, nil, nil)
}

// GetStats returns database connection statistics
func (d *DatabaseAdapter) GetStats() interface{} {
	// Return a simple stats object
	return map[string]interface{}{
		"connected": d.db.IsConnected(),
	}
}

// GetMonitor returns the database monitor
func (d *DatabaseAdapter) GetMonitor() interface{} {
	// Return a simple monitor object
	return map[string]interface{}{
		"type": "gorm",
	}
}

// GetHealthChecker returns the database health checker
func (d *DatabaseAdapter) GetHealthChecker() interface{} {
	// Return a simple health checker object
	return map[string]interface{}{
		"type": "gorm",
	}
}

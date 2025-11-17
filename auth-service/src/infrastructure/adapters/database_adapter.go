package adapters

import (
	"context"
	"fmt"
	"time"

	"backend-core/adapters"
	"backend-core/database/gorm"
	"backend-core/database/health"
	"backend-core/logging"
)

// DatabaseAdapter wraps the shared database adapter for auth-service
// This provides a service-specific interface while reusing backend-core implementation
type DatabaseAdapter struct {
	*adapters.DatabaseAdapter
	logger *logging.Logger
}

// NewDatabaseAdapter creates a new database adapter using the shared implementation
func NewDatabaseAdapter(db gorm.Database, logger *logging.Logger) *DatabaseAdapter {
	baseAdapter := adapters.NewDatabaseAdapter(db)
	return &DatabaseAdapter{
		DatabaseAdapter: baseAdapter,
		logger:          logger,
	}
}

// InitializeDatabase initializes the database connection with auth-service specific configuration
func (d *DatabaseAdapter) InitializeDatabase(ctx context.Context) error {
	d.logger.Info("Initializing database connection for auth-service")

	// Connect to database
	if err := d.Connect(ctx); err != nil {
		d.logger.Error("Failed to connect to database", logging.Error(err))
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := d.RunMigrations(ctx); err != nil {
		d.logger.Error("Failed to run database migrations", logging.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	d.logger.Info("Database initialized successfully for auth-service")
	return nil
}

// GetAuthServiceRepository returns a repository specifically configured for auth-service
func (d *DatabaseAdapter) GetAuthServiceRepository() interface{} {
	// Return the underlying GORM database for auth-service specific operations
	return d.DatabaseAdapter.GetRepository()
}

// HealthCheck performs a comprehensive health check for auth-service
func (d *DatabaseAdapter) HealthCheck() health.HealthCheck {
	ctx := context.Background()

	// Perform basic connectivity check
	if err := d.Ping(ctx); err != nil {
		d.logger.Error("Database health check failed", logging.Error(err))
		return *health.NewHealthCheck(health.HealthStatusUnhealthy, 0, err, nil)
	}

	// Check if database is healthy
	if !d.IsHealthy(ctx) {
		d.logger.Warn("Database health check indicates unhealthy state")
		return *health.NewHealthCheck(health.HealthStatusUnhealthy, 0, fmt.Errorf("database is not healthy"), nil)
	}

	d.logger.Debug("Database health check passed")
	return *health.NewHealthCheck(health.HealthStatusHealthy, 0, nil, nil)
}

// GetConnectionStats returns connection statistics for monitoring
func (d *DatabaseAdapter) GetConnectionStats() map[string]interface{} {
	stats := d.GetStats()

	// Add auth-service specific metrics
	authServiceStats := map[string]interface{}{
		"service":    "auth-service",
		"adapter":    "database_adapter",
		"timestamp":  time.Now(),
		"base_stats": stats,
	}

	return authServiceStats
}

// Close gracefully closes the database connection
func (d *DatabaseAdapter) Close(ctx context.Context) error {
	d.logger.Info("Closing database connection for auth-service")
	return d.Disconnect(ctx)
}

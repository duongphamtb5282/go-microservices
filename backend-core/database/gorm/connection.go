package gorm

import (
	"context"
	"errors"

	"backend-core/config"
	"backend-core/logging"

	"gorm.io/gorm"
)

// ConnectionManager handles database connection operations
type ConnectionManager struct {
	config *config.DatabaseConfig
	gormDB *gorm.DB
	logger *logging.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(config *config.DatabaseConfig, gormDB *gorm.DB, logger *logging.Logger) *ConnectionManager {
	return &ConnectionManager{
		config: config,
		gormDB: gormDB,
		logger: logger,
	}
}

// Connect establishes database connection
func (c *ConnectionManager) Connect(ctx context.Context) error {
	// Connection is already established in the factory
	c.logger.Info("database connected")
	return nil
}

// Disconnect closes database connection
func (c *ConnectionManager) Disconnect(ctx context.Context) error {
	if c.gormDB != nil {
		sqlDB, err := c.gormDB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// IsConnected checks if database is connected
func (c *ConnectionManager) IsConnected() bool {
	if c.gormDB == nil {
		return false
	}
	sqlDB, err := c.gormDB.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}

// HealthCheck performs a health check
func (c *ConnectionManager) HealthCheck(ctx context.Context) error {
	if c.gormDB == nil {
		return errors.New("database not connected")
	}

	sqlDB, err := c.gormDB.DB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

// GetGormDB returns the GORM database instance
func (c *ConnectionManager) GetGormDB() *gorm.DB {
	return c.gormDB
}

// GetConfig returns the database configuration
func (c *ConnectionManager) GetConfig() *config.DatabaseConfig {
	return c.config
}

// GetLogger returns the logger
func (c *ConnectionManager) GetLogger() *logging.Logger {
	return c.logger
}

// GetStats returns database connection statistics
func (c *ConnectionManager) GetStats() ConnectionStats {
	if c.gormDB == nil {
		return ConnectionStats{}
	}

	sqlDB, err := c.gormDB.DB()
	if err != nil {
		return ConnectionStats{}
	}

	stats := sqlDB.Stats()
	return ConnectionStats{
		OpenConnections:   stats.OpenConnections,
		IdleConnections:   stats.Idle,
		InUseConnections:  stats.InUse,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

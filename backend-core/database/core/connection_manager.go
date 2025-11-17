package core

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectionManager manages database connections with enhanced features
type ConnectionManager struct {
	db     *gorm.DB
	config *Config
	logger *zap.Logger
	mutex  sync.RWMutex
	// monitorCancel stops the background metrics collection
	monitorCancel context.CancelFunc

	// Connection pool monitoring
	stats *ConnectionStats
}

// ConnectionStats tracks connection pool metrics
type ConnectionStats struct {
	mutex             sync.RWMutex
	OpenConnections   int
	InUseConnections  int
	IdleConnections   int
	WaitCount         int64
	WaitDuration      time.Duration
	MaxIdleClosed     int64
	MaxIdleTimeClosed int64
	MaxLifetimeClosed int64
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(config *Config, logger *zap.Logger) *ConnectionManager {
	return &ConnectionManager{
		config: config,
		logger: logger,
		stats:  &ConnectionStats{},
	}
}

// Connect establishes database connection with enhanced configuration
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.db != nil {
		return nil // Already connected
	}

	// Configure GORM logger based on log level
	gormLogger := cm.getGormLogger()

	// Build connection string
	dsn := cm.config.ConnectionString()

	// Configure connection pool
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxOpenConns(cm.config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cm.config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cm.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cm.config.ConnMaxIdleTime)

	// Set connection timeouts
	sqlDB.SetConnMaxIdleTime(cm.config.ConnMaxIdleTime)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(ctx, cm.config.ConnectTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create GORM DB instance
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger:                                   gormLogger,
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: true,
		CreateBatchSize:                          1000,
	})

	if err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to create GORM instance: %w", err)
	}

	// Configure connection pool
	db, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set prepared statement cache size
	if cm.config.PreparedStatementCacheSize > 0 {
		// This is a PostgreSQL driver specific setting
		// We'll handle it through connection string parameters if needed
	}

	// Start monitoring goroutine with independent lifecycle
	monitorCtx, monitorCancel := context.WithCancel(context.Background())
	cm.monitorCancel = monitorCancel
	go cm.startMonitoring(monitorCtx, db)

	cm.db = gormDB
	cm.logger.Info("Database connected successfully",
		zap.Int("max_open_conns", cm.config.MaxOpenConns),
		zap.Int("max_idle_conns", cm.config.MaxIdleConns),
		zap.Duration("conn_max_lifetime", cm.config.ConnMaxLifetime),
		zap.Duration("conn_max_idle_time", cm.config.ConnMaxIdleTime),
	)

	return nil
}

// GetDB returns the GORM database instance
func (cm *ConnectionManager) GetDB() *gorm.DB {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.db
}

// GetStats returns current connection pool statistics
func (cm *ConnectionManager) GetStats() ConnectionStats {
	cm.stats.mutex.RLock()
	defer cm.stats.mutex.RUnlock()
	return *cm.stats
}

// Close closes the database connection
func (cm *ConnectionManager) Close() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.monitorCancel != nil {
		cm.monitorCancel()
		cm.monitorCancel = nil
	}

	if cm.db == nil {
		return nil
	}

	sqlDB, err := cm.db.DB()
	if err != nil {
		return err
	}

	cm.logger.Info("Closing database connection")
	return sqlDB.Close()
}

// getGormLogger returns configured GORM logger
func (cm *ConnectionManager) getGormLogger() logger.Interface {
	switch cm.config.LogLevel {
	case "debug":
		return logger.Default.LogMode(logger.Info)
	case "info":
		return logger.Default.LogMode(logger.Info)
	case "warn":
		return logger.Default.LogMode(logger.Warn)
	case "error":
		return logger.Default.LogMode(logger.Error)
	default:
		return logger.Default.LogMode(logger.Warn)
	}
}

// startMonitoring monitors connection pool metrics
func (cm *ConnectionManager) startMonitoring(ctx context.Context, db *sql.DB) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := db.Stats()

			cm.stats.mutex.Lock()
			cm.stats.OpenConnections = stats.OpenConnections
			cm.stats.InUseConnections = stats.InUse
			cm.stats.IdleConnections = stats.Idle
			cm.stats.WaitCount = stats.WaitCount
			cm.stats.WaitDuration = stats.WaitDuration
			cm.stats.MaxIdleClosed = stats.MaxIdleClosed
			cm.stats.MaxIdleTimeClosed = stats.MaxIdleTimeClosed
			cm.stats.MaxLifetimeClosed = stats.MaxLifetimeClosed
			cm.stats.mutex.Unlock()

			// Log warnings for high usage
			var usagePercent float64
			if stats.OpenConnections > 0 {
				usagePercent = float64(stats.InUse) / float64(stats.OpenConnections) * 100
			}
			if usagePercent > 80 {
				cm.logger.Warn("High database connection usage",
					zap.Float64("usage_percent", usagePercent),
					zap.Int("open_connections", stats.OpenConnections),
					zap.Int("in_use", stats.InUse),
					zap.Int("idle", stats.Idle),
				)
			}
		}
	}
}

// ExecuteWithTimeout executes a function with query timeout
func (cm *ConnectionManager) ExecuteWithTimeout(ctx context.Context, fn func(*gorm.DB) error) error {
	if cm.db == nil {
		return fmt.Errorf("database not connected")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, cm.config.QueryTimeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- fn(cm.db.WithContext(timeoutCtx))
	}()

	select {
	case err := <-done:
		return err
	case <-timeoutCtx.Done():
		return fmt.Errorf("query timeout after %v", cm.config.QueryTimeout)
	}
}

// HealthCheck performs database health check
func (cm *ConnectionManager) HealthCheck(ctx context.Context) error {
	if cm.db == nil {
		return fmt.Errorf("database not connected")
	}

	healthCtx, cancel := context.WithTimeout(ctx, cm.config.HealthCheckTimeout)
	defer cancel()

	return cm.ExecuteWithTimeout(healthCtx, func(db *gorm.DB) error {
		return db.Raw("SELECT 1").Error
	})
}

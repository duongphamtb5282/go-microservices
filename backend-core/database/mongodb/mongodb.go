package mongodb

import (
	"context"
	"fmt"
	"time"

	"backend-core/config"
	"backend-core/database"
	"backend-core/database/health"
	"backend-core/logging"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBDatabase implements the ExtendedDatabase interface for MongoDB
type MongoDBDatabase struct {
	config             *config.DatabaseConfig
	logger             *logging.Logger
	client             *mongo.Client
	database           *mongo.Database
	queryBuilder       *MongoDBQueryBuilder
	migrationManager   *MongoDBMigrationManager
	transactionManager *MongoDBTransactionManager
	healthChecker      *MongoDBHealthChecker
	monitor            *MongoDBMonitor
	isConnected        bool
	connectedAt        time.Time
}

// NewMongoDBDatabase creates a new MongoDB database instance
func NewMongoDBDatabase(cfg *config.DatabaseConfig) *MongoDBDatabase {
	logger, _ := logging.NewLogger(&config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "stdout",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	})

	return &MongoDBDatabase{
		config: cfg,
		logger: logger,
	}
}

// GetConfig returns the database configuration
func (m *MongoDBDatabase) GetConfig() *config.DatabaseConfig {
	return m.config
}

// GetLogger returns the logger
func (m *MongoDBDatabase) GetLogger() *logging.Logger {
	return m.logger
}

// IsConnected returns the connection status
func (m *MongoDBDatabase) IsConnected() bool {
	return m.isConnected
}

// SetConnected sets the connection status
func (m *MongoDBDatabase) SetConnected(connected bool) {
	m.isConnected = connected
	if connected {
		m.connectedAt = time.Now()
	}
}

// GetConnectedAt returns when the database was connected
func (m *MongoDBDatabase) GetConnectedAt() time.Time {
	return m.connectedAt
}

// Connect establishes connection to MongoDB
func (m *MongoDBDatabase) Connect(ctx context.Context) error {
	// Use DSN provider for connection string generation
	dsnProvider := config.NewDSNProvider(m.config.Type, m.config)
	if dsnProvider == nil {
		return fmt.Errorf("unsupported database type: %s", m.config.Type)
	}

	clientOptions := options.Client().ApplyURI(dsnProvider.Dsn())
	m.logger.Info("Connecting to MongoDB")

	// Set connection pool options
	clientOptions.SetMaxPoolSize(uint64(m.config.MaxOpenConns))
	clientOptions.SetMinPoolSize(uint64(m.config.MaxIdleConns))
	clientOptions.SetMaxConnIdleTime(m.config.ConnMaxLifetime)
	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		m.logger.Error("Failed to connect to MongoDB", logging.Error(err))
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		m.logger.Info("ping_failed", err)
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database
	database := client.Database(m.config.Database)

	// Set database components
	m.client = client
	m.database = database
	m.isConnected = true

	// Initialize components
	m.queryBuilder = NewMongoDBQueryBuilder(database)
	m.migrationManager = NewMongoDBMigrationManager(database)
	m.transactionManager = NewMongoDBTransactionManager(client)
	m.healthChecker = NewMongoDBHealthChecker(client, m.config)
	m.monitor = NewMongoDBMonitor()

	m.logger.Info("connected", nil)
	return nil
}

// Disconnect closes the database connection
func (m *MongoDBDatabase) Disconnect(ctx context.Context) error {
	if m.client != nil {
		m.logger.Info("disconnecting", nil)
		if err := m.client.Disconnect(ctx); err != nil {
			m.logger.Info("disconnect_failed", err)
			return fmt.Errorf("failed to close MongoDB connection: %w", err)
		}
		m.SetConnected(false)
		m.logger.Info("disconnected", nil)
	}
	return nil
}

// Ping tests the database connection
func (m *MongoDBDatabase) Ping(ctx context.Context) error {
	// Connection validation not needed for ping

	if err := m.client.Ping(ctx, nil); err != nil {
		m.logger.Info("ping_failed", err)
		return fmt.Errorf("MongoDB ping failed: %w", err)
	}

	return nil
}

// IsHealthy checks if the database is healthy
func (m *MongoDBDatabase) IsHealthy(ctx context.Context) bool {
	if m.healthChecker != nil {
		return m.healthChecker.GetStatus() == database.HealthStatusHealthy
	}

	// Fallback to ping
	if err := m.Ping(ctx); err != nil {
		return false
	}

	return true
}

// GetRepository returns a repository for the given entity type
func (m *MongoDBDatabase) GetRepository() interface{} {
	collectionName := "default"
	return NewMongoDBRepository[interface{}](m, collectionName)
}

// WithTransaction executes a function within a transaction
func (m *MongoDBDatabase) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	})

	return err
}

// RunMigrations runs database migrations
func (m *MongoDBDatabase) RunMigrations(ctx context.Context) error {
	if m.migrationManager != nil {
		return m.migrationManager.RunMigrations(ctx)
	}
	return fmt.Errorf("migration manager not initialized")
}

// RollbackMigration rolls back a specific migration
func (m *MongoDBDatabase) RollbackMigration(ctx context.Context, version int) error {
	if m.migrationManager != nil {
		return m.migrationManager.RollbackMigration(ctx, version)
	}
	return fmt.Errorf("migration manager not initialized")
}

// GetStats returns connection statistics
func (m *MongoDBDatabase) GetStats() database.ConnectionStats {
	// MongoDB doesn't provide the same connection stats as SQL databases
	// This is a simplified implementation
	return health.NewConnectionStats(0, 0)
}

// SetMaxOpenConns sets the maximum number of open connections
func (m *MongoDBDatabase) SetMaxOpenConns(max int) {
	// MongoDB connection pool is configured during connection
	// This is a no-op for MongoDB
}

// SetMaxIdleConns sets the maximum number of idle connections
func (m *MongoDBDatabase) SetMaxIdleConns(max int) {
	// MongoDB connection pool is configured during connection
	// This is a no-op for MongoDB
}

// SetConnMaxLifetime sets the maximum lifetime of connections
func (m *MongoDBDatabase) SetConnMaxLifetime(d time.Duration) {
	// MongoDB connection pool is configured during connection
	// This is a no-op for MongoDB
}

// ExtendedDatabase interface methods

// GetDriver returns the MongoDB client
func (m *MongoDBDatabase) GetDriver() interface{} {
	return m.client
}

// GetConnectionPool returns the MongoDB client (same as driver for MongoDB)
func (m *MongoDBDatabase) GetConnectionPool() interface{} {
	return m.client
}

// GetQueryBuilder returns the MongoDB query builder
func (m *MongoDBDatabase) GetQueryBuilder() database.QueryBuilder {
	return m.queryBuilder
}

// GetMigrationManager returns the MongoDB migration manager
func (m *MongoDBDatabase) GetMigrationManager() database.MigrationManager {
	return m.migrationManager
}

// GetTransactionManager returns the MongoDB transaction manager
func (m *MongoDBDatabase) GetTransactionManager() database.TransactionManager {
	return m.transactionManager
}

// GetHealthChecker returns the MongoDB health checker
func (m *MongoDBDatabase) GetHealthChecker() database.DatabaseHealthChecker {
	return m.healthChecker
}

// GetMonitor returns the MongoDB monitor
func (m *MongoDBDatabase) GetMonitor() database.DatabaseMonitor {
	return m.monitor
}

// Helper functions

func getCollectionName[T any]() string {
	// This is a simplified implementation
	// In a real implementation, you might use reflection or struct tags
	var t T
	return fmt.Sprintf("%T", t)
}

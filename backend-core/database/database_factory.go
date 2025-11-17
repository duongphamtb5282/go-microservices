package database

import (
	"fmt"

	"backend-core/config"
	"backend-core/database/postgresql"
	"backend-core/logging"
	// "backend-core/database/mongodb" // Temporarily disabled due to dependency issues
)

// DatabaseFactory creates database instances
type DatabaseFactory struct{}

// NewDatabaseFactory creates a new database factory
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{}
}

// CreateDatabase creates a database instance based on the type
func (f *DatabaseFactory) CreateDatabase(dbType DatabaseType, cfg *config.DatabaseConfig) (Database, error) {
	switch dbType {
	case PostgreSQL:
		db := postgresql.NewPostgreSQLDatabase(cfg)
		return &postgresql.DatabaseWrapper{PostgreSQLDatabase: db}, nil
	case MongoDB:
		return nil, fmt.Errorf("MongoDB support temporarily disabled due to dependency issues")
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// CreatePostgreSQLDatabase creates a PostgreSQL database instance
func (f *DatabaseFactory) CreatePostgreSQLDatabase(cfg *config.DatabaseConfig) (Database, error) {
	db := postgresql.NewPostgreSQLDatabase(cfg)
	return &postgresql.DatabaseWrapper{PostgreSQLDatabase: db}, nil
}

// CreateMongoDBDatabase creates a MongoDB database instance
func (f *DatabaseFactory) CreateMongoDBDatabase(cfg *config.DatabaseConfig) (Database, error) {
	return nil, fmt.Errorf("MongoDB support temporarily disabled due to dependency issues")
}

// CreateRepository creates a repository for the given entity type and database
func (f *DatabaseFactory) CreateRepository(db Database, collectionName string) interface{} {
	switch db := db.(type) {
	case *postgresql.DatabaseWrapper:
		// Create a logger for the repository
		logger, _ := logging.NewLogger(&config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		})
		return postgresql.NewPostgreSQLRepository[interface{}](db.PostgreSQLDatabase, collectionName, logger)
	// case *mongodb.MongoDBDatabase:
	//	return mongodb.NewMongoDBRepository[interface{}](db, collectionName)
	default:
		return nil
	}
}

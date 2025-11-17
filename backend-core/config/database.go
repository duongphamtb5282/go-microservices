package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

// DsnProvider defines the interface for generating database connection strings
type DsnProvider interface {
	Dsn() string
}

// DatabaseConfig holds enhanced database configuration for production
type DatabaseConfig struct {
	Type     string `mapstructure:"type" validate:"required,oneof=postgresql postgres mongodb"` // postgresql, mongodb
	Host     string `mapstructure:"host" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	Database string `mapstructure:"database" validate:"required"`
	Username string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
	SSLMode  string `mapstructure:"ssl_mode" validate:"omitempty,oneof=disable require verify-ca verify-full"`

	// Enhanced connection pooling for production
	MaxOpenConns    int           `mapstructure:"max_connections" validate:"required,min=1"`
	MaxIdleConns    int           `mapstructure:"max_idle_connections" validate:"required,min=0"`
	ConnMaxLifetime time.Duration `mapstructure:"connection_max_lifetime" validate:"required"`
	ConnMaxIdleTime time.Duration `mapstructure:"connection_max_idle_time" validate:"required"`

	// Query optimization settings
	QueryTimeout         time.Duration `mapstructure:"query_timeout" validate:"required"`
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" validate:"required"`
	AcquireTimeout       time.Duration `mapstructure:"acquire_timeout" validate:"required"`
	AcquireRetryAttempts int           `mapstructure:"acquire_retry_attempts" validate:"required,min=1,max=10"`

	// Prepared statement cache
	PreparedStatementCacheSize int `mapstructure:"prepared_statement_cache_size" validate:"required,min=0"`

	// Logging settings
	LogLevel string `mapstructure:"log_level" validate:"omitempty,oneof=silent error warn info"`

	// Database-specific configurations
	PostgreSQL *PostgreSQLConfig `mapstructure:"postgresql,omitempty" validate:"omitempty"`
	MongoDB    *MongoDBConfig    `mapstructure:"mongodb,omitempty" validate:"omitempty"`
}

// Dsn implements the DsnProvider interface
func (c *DatabaseConfig) Dsn() string {
	switch c.Type {
	case "postgres":
		if c.PostgreSQL != nil {
			return c.PostgreSQL.Dsn()
		}
		// Fallback to basic config
		postgresConfig := NewPostgreSQLConfig()
		postgresConfig.SetConnection(c.Host, c.Port, c.Database)
		postgresConfig.SetCredentials(c.Username, c.Password)
		postgresConfig.SetSSLMode(c.SSLMode)
		return postgresConfig.Dsn()
	case "mongodb":
		if c.MongoDB != nil {
			return c.MongoDB.Dsn()
		}
		// Fallback to basic config
		mongoConfig := NewMongoDBConfig()
		mongoConfig.SetDatabase(c.Database)
		mongoConfig.SetCredentials(c.Username, c.Password)
		mongoConfig.AddHost(c.Host, fmt.Sprintf("%d", c.Port))
		return mongoConfig.Dsn()
	default:
		return ""
	}
}

// GetDatabaseDSN returns database connection string (backward compatibility)
func (c *DatabaseConfig) GetDatabaseDSN() string {
	return c.Dsn()
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("database configuration validation failed: %w", err)
	}
	return nil
}

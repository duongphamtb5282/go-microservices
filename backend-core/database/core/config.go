package core

import (
	"fmt"
	"time"
)

// Config holds core database configuration
type Config struct {
	// Connection settings
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Database string `mapstructure:"database" json:"database" yaml:"database"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	SSLMode  string `mapstructure:"ssl_mode" json:"ssl_mode" yaml:"ssl_mode"`

	// Enhanced connection pooling for production
	MaxOpenConns    int           `mapstructure:"max_connections" json:"max_connections" yaml:"max_connections"`
	MaxIdleConns    int           `mapstructure:"max_idle_connections" json:"max_idle_connections" yaml:"max_idle_connections"`
	ConnMaxLifetime time.Duration `mapstructure:"connection_max_lifetime" json:"connection_max_lifetime" yaml:"connection_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"connection_max_idle_time" json:"connection_max_idle_time" yaml:"connection_max_idle_time"`

	// Query optimization settings
	SlowQueryThreshold   time.Duration `mapstructure:"slow_query_threshold" json:"slow_query_threshold" yaml:"slow_query_threshold"`
	AcquireTimeout       time.Duration `mapstructure:"acquire_timeout" json:"acquire_timeout" yaml:"acquire_timeout"`
	AcquireRetryAttempts int           `mapstructure:"acquire_retry_attempts" json:"acquire_retry_attempts" yaml:"acquire_retry_attempts"`

	// Prepared statement cache
	PreparedStatementCacheSize int `mapstructure:"prepared_statement_cache_size" json:"prepared_statement_cache_size" yaml:"prepared_statement_cache_size"`

	// Enhanced timeouts
	ConnectTimeout time.Duration `mapstructure:"connect_timeout" json:"connect_timeout" yaml:"connect_timeout"`
	QueryTimeout   time.Duration `mapstructure:"query_timeout" json:"query_timeout" yaml:"query_timeout"`

	// Logging
	LogLevel string `mapstructure:"log_level" json:"log_level" yaml:"log_level"`

	// Migration settings
	MigrationsPath string `mapstructure:"migrations_path" json:"migrations_path" yaml:"migrations_path"`
	AutoMigrate    bool   `mapstructure:"auto_migrate" json:"auto_migrate" yaml:"auto_migrate"`

	// Health check settings
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval" json:"health_check_interval" yaml:"health_check_interval"`
	HealthCheckTimeout  time.Duration `mapstructure:"health_check_timeout" json:"health_check_timeout" yaml:"health_check_timeout"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",

		// Enhanced connection pooling defaults
		MaxOpenConns:    100,
		MaxIdleConns:    25,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,

		// Query optimization defaults
		SlowQueryThreshold:         1 * time.Second,
		AcquireTimeout:             30 * time.Second,
		AcquireRetryAttempts:       3,
		PreparedStatementCacheSize: 100,

		// Enhanced timeouts
		ConnectTimeout: 30 * time.Second,
		QueryTimeout:   30 * time.Second,

		// Logging
		LogLevel: "warn",

		// Migration settings
		MigrationsPath: "./migrations",
		AutoMigrate:    true,

		// Health check settings
		HealthCheckInterval: 30 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
	}
}

// ProductionConfig returns a production-ready configuration
func ProductionConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     5432,
		Database: "postgres",
		Username: "postgres",
		Password: "", // Must be set via environment
		SSLMode:  "require",

		// High-performance connection pooling
		MaxOpenConns:    200,
		MaxIdleConns:    50,
		ConnMaxLifetime: 1 * time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,

		// Production query optimization
		SlowQueryThreshold:         500 * time.Millisecond,
		AcquireTimeout:             30 * time.Second,
		AcquireRetryAttempts:       5,
		PreparedStatementCacheSize: 200,

		// Production timeouts
		ConnectTimeout: 30 * time.Second,
		QueryTimeout:   10 * time.Second,

		// Production logging
		LogLevel: "error",

		// Migration settings
		MigrationsPath: "./migrations",
		AutoMigrate:    false, // Manual migrations in production

		// Health check settings
		HealthCheckInterval: 15 * time.Second,
		HealthCheckTimeout:  3 * time.Second,
	}
}

// ConnectionString builds the database connection string
func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

// Validate validates the database configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}
	if c.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections cannot be negative")
	}
	if c.MaxIdleConns > c.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}
	return nil
}

// GetEffectiveMaxIdleConns returns the effective max idle connections
func (c *Config) GetEffectiveMaxIdleConns() int {
	if c.MaxIdleConns == 0 {
		// Default to 25% of max open connections
		return c.MaxOpenConns / 4
	}
	return c.MaxIdleConns
}

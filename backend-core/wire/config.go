package wire

import (
	"backend-core/config"
	"backend-core/monitoring"
	"errors"
	"fmt"
)

// ServiceConfig represents configuration for a specific service
type ServiceConfig struct {
	ServiceName string
	Strategy    string
	Port        string
	Database    config.DatabaseConfig
	Cache       config.RedisConfig
	Logging     config.LoggingConfig
	Security    config.SecurityConfig
	Monitoring  monitoring.Config
}

// NewServiceConfig creates a new service configuration
func NewServiceConfig(serviceName, strategy string, baseConfig *config.Config) *ServiceConfig {
	return &ServiceConfig{
		ServiceName: serviceName,
		Strategy:    strategy,
		Port:        "8080", // Default port
		Database:    baseConfig.Database,
		Cache:       baseConfig.Redis,
		Logging:     baseConfig.Logging,
		Security:    baseConfig.Security,
		Monitoring:  baseConfig.Monitoring,
	}
}

// GetServiceName returns the service name
func (c *ServiceConfig) GetServiceName() string {
	return c.ServiceName
}

// GetStrategy returns the cache strategy
func (c *ServiceConfig) GetStrategy() string {
	return c.Strategy
}

// GetPort returns the service port
func (c *ServiceConfig) GetPort() string {
	return c.Port
}

// GetDatabaseConfig returns the database configuration
func (c *ServiceConfig) GetDatabaseConfig() config.DatabaseConfig {
	return c.Database
}

// GetCacheConfig returns the cache configuration
func (c *ServiceConfig) GetCacheConfig() config.RedisConfig {
	return c.Cache
}

// GetLoggingConfig returns the logging configuration
func (c *ServiceConfig) GetLoggingConfig() config.LoggingConfig {
	return c.Logging
}

// GetSecurityConfig returns the security configuration
func (c *ServiceConfig) GetSecurityConfig() config.SecurityConfig {
	return c.Security
}

// GetMonitoringConfig returns the monitoring configuration
func (c *ServiceConfig) GetMonitoringConfig() monitoring.Config {
	return c.Monitoring
}

// Validate validates the service configuration
func (c *ServiceConfig) Validate() error {
	// Validate service name
	if c.ServiceName == "" {
		return errors.New("service name is required")
	}

	// Validate strategy
	if c.Strategy == "" {
		return errors.New("cache strategy is required")
	}

	// Validate port
	if c.Port == "" {
		return errors.New("service port is required")
	}

	// Validate database configuration
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration is invalid: %w", err)
	}

	// Validate cache configuration
	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache configuration is invalid: %w", err)
	}

	// Validate logging configuration
	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging configuration is invalid: %w", err)
	}

	// Validate security configuration
	if err := c.Security.Validate(); err != nil {
		return fmt.Errorf("security configuration is invalid: %w", err)
	}

	// Validate monitoring configuration
	// Note: MonitoringConfig doesn't have Validate method, skipping validation

	return nil
}

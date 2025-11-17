package config

import (
	"fmt"

	"backend-core/monitoring"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Database   DatabaseConfig    `mapstructure:"database"`
	Redis      RedisConfig       `mapstructure:"redis"`
	Logging    LoggingConfig     `mapstructure:"logging"`
	Zap        Zap               `mapstructure:"zap"`
	Security   SecurityConfig    `mapstructure:"security"`
	Masking    MaskingConfig     `mapstructure:"masking"`
	Monitoring monitoring.Config `mapstructure:"monitoring"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// All configuration values are loaded from config.yml
	// No defaults are applied - all values must be explicitly configured

	// Apply default masking config if not provided
	if config.Masking.Environment == "" {
		config.Masking = *GetDefaultMaskingConfig()
	}

	// Validate that all required configuration sections are present
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// validate validates that all required configuration sections are present
func (c *Config) validate() error {
	// Validate database configuration
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database configuration: %w", err)
	}

	// Validate Redis configuration
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis configuration: %w", err)
	}

	// Validate logging configuration
	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging configuration: %w", err)
	}

	// Validate Zap configuration
	if err := c.Zap.Validate(); err != nil {
		return fmt.Errorf("zap configuration: %w", err)
	}

	// Validate security configuration
	if err := c.Security.Validate(); err != nil {
		return fmt.Errorf("security configuration: %w", err)
	}

	// Validate masking configuration
	if err := c.Masking.Validate(); err != nil {
		return fmt.Errorf("masking configuration: %w", err)
	}

	return nil
}

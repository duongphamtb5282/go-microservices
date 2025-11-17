package monitoring

import "fmt"

// Config holds monitoring configuration
type Config struct {
	Enabled     bool   `mapstructure:"enabled"`
	MetricsPort int    `mapstructure:"metrics_port"`
	HealthPort  int    `mapstructure:"health_port"`
	LogLevel    string `mapstructure:"log_level"`
}

// GetDefaultConfig returns default monitoring configuration
func GetDefaultConfig() *Config {
	return &Config{
		Enabled:     false,
		MetricsPort: 9090,
		HealthPort:  8080,
		LogLevel:    "info",
	}
}

// IsEnabled returns whether monitoring is enabled
func (c *Config) IsEnabled() bool {
	return c.Enabled
}

// GetMetricsPort returns the metrics port
func (c *Config) GetMetricsPort() int {
	return c.MetricsPort
}

// GetHealthPort returns the health check port
func (c *Config) GetHealthPort() int {
	return c.HealthPort
}

// GetLogLevel returns the log level
func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

// Validate validates the monitoring configuration
func (c *Config) Validate() error {
	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d", c.MetricsPort)
	}

	if c.HealthPort <= 0 || c.HealthPort > 65535 {
		return fmt.Errorf("invalid health port: %d", c.HealthPort)
	}

	if c.MetricsPort == c.HealthPort {
		return fmt.Errorf("metrics port and health port cannot be the same: %d", c.MetricsPort)
	}

	validLogLevels := []string{"debug", "info", "warn", "error"}
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			return nil
		}
	}
	return fmt.Errorf("invalid log level: %s", c.LogLevel)
}

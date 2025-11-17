package health

import (
	"time"
)

// Config holds health monitoring configuration
type Config struct {
	// Health check settings
	Enabled       bool          `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	CheckInterval time.Duration `mapstructure:"check_interval" json:"check_interval" yaml:"check_interval"`
	CheckTimeout  time.Duration `mapstructure:"check_timeout" json:"check_timeout" yaml:"check_timeout"`
	RetryCount    int           `mapstructure:"retry_count" json:"retry_count" yaml:"retry_count"`
	RetryInterval time.Duration `mapstructure:"retry_interval" json:"retry_interval" yaml:"retry_interval"`

	// Performance monitoring
	PerformanceEnabled bool          `mapstructure:"performance_enabled" json:"performance_enabled" yaml:"performance_enabled"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold" json:"slow_query_threshold" yaml:"slow_query_threshold"`
	MetricsRetention   time.Duration `mapstructure:"metrics_retention" json:"metrics_retention" yaml:"metrics_retention"`

	// Alerting settings
	AlertEnabled   bool          `mapstructure:"alert_enabled" json:"alert_enabled" yaml:"alert_enabled"`
	AlertThreshold time.Duration `mapstructure:"alert_threshold" json:"alert_threshold" yaml:"alert_threshold"`
	AlertCooldown  time.Duration `mapstructure:"alert_cooldown" json:"alert_cooldown" yaml:"alert_cooldown"`

	// Status reporting
	StatusEndpoint  string `mapstructure:"status_endpoint" json:"status_endpoint" yaml:"status_endpoint"`
	MetricsEndpoint string `mapstructure:"metrics_endpoint" json:"metrics_endpoint" yaml:"metrics_endpoint"`
}

// DefaultConfig returns a default health configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:            true,
		CheckInterval:      30 * time.Second,
		CheckTimeout:       5 * time.Second,
		RetryCount:         3,
		RetryInterval:      1 * time.Second,
		PerformanceEnabled: true,
		SlowQueryThreshold: 1 * time.Second,
		MetricsRetention:   24 * time.Hour,
		AlertEnabled:       true,
		AlertThreshold:     5 * time.Second,
		AlertCooldown:      5 * time.Minute,
		StatusEndpoint:     "/health",
		MetricsEndpoint:    "/metrics",
	}
}

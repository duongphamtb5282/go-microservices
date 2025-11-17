package factory

import (
	"time"

	cacheconfig "backend-core/cache/config"
)

// Config represents middleware configuration
type Config struct {
	Cache      cacheconfig.Config `yaml:"cache" json:"cache"`
	Security   SecurityConfig     `yaml:"security" json:"security"`
	Monitoring MonitoringConfig   `yaml:"monitoring" json:"monitoring"`
	HTTP       HTTPConfig         `yaml:"http" json:"http"`
	Logging    LoggingConfig      `yaml:"logging" json:"logging"`
	Validation ValidationConfig   `yaml:"validation" json:"validation"`
}

// SecurityConfig represents security middleware configuration
type SecurityConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Strategy    string `yaml:"strategy" json:"strategy"`
	JWTSecret   string `yaml:"jwt_secret" json:"jwt_secret"`
	TokenExpiry string `yaml:"token_expiry" json:"token_expiry"`
	APIKey      string `yaml:"api_key" json:"api_key"`
}

// MonitoringConfig represents monitoring middleware configuration
type MonitoringConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	Collector      string `yaml:"collector" json:"collector"`
	Port           int    `yaml:"port" json:"port"`
	MetricsPath    string `yaml:"metrics_path" json:"metrics_path"`
	HealthPath     string `yaml:"health_path" json:"health_path"`
	TracingEnabled bool   `yaml:"tracing_enabled" json:"tracing_enabled"`
}

// HTTPConfig represents HTTP middleware configuration
type HTTPConfig struct {
	CORS        CORSConfig        `yaml:"cors" json:"cors"`
	RateLimit   RateLimitConfig   `yaml:"rate_limit" json:"rate_limit"`
	Compression CompressionConfig `yaml:"compression" json:"compression"`
	Timeout     TimeoutConfig     `yaml:"timeout" json:"timeout"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Origins     []string `yaml:"origins" json:"origins"`
	Methods     []string `yaml:"methods" json:"methods"`
	Headers     []string `yaml:"headers" json:"headers"`
	Credentials bool     `yaml:"credentials" json:"credentials"`
	MaxAge      int      `yaml:"max_age" json:"max_age"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	RequestsPerMinute int           `yaml:"requests_per_minute" json:"requests_per_minute"`
	BurstSize         int           `yaml:"burst_size" json:"burst_size"`
	WindowSize        time.Duration `yaml:"window_size" json:"window_size"`
}

// CompressionConfig represents compression configuration
type CompressionConfig struct {
	Enabled bool     `yaml:"enabled" json:"enabled"`
	Level   int      `yaml:"level" json:"level"`
	Types   []string `yaml:"types" json:"types"`
	MinSize int      `yaml:"min_size" json:"min_size"`
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	Enabled      bool          `yaml:"enabled" json:"enabled"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
}

// LoggingConfig represents logging middleware configuration
type LoggingConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	Level          string `yaml:"level" json:"level"`
	Format         string `yaml:"format" json:"format"`
	IncludeBody    bool   `yaml:"include_body" json:"include_body"`
	IncludeHeaders bool   `yaml:"include_headers" json:"include_headers"`
	MaxBodySize    int64  `yaml:"max_body_size" json:"max_body_size"`
}

// ValidationConfig represents validation middleware configuration
type ValidationConfig struct {
	Enabled        bool     `yaml:"enabled" json:"enabled"`
	ValidateInput  bool     `yaml:"validate_input" json:"validate_input"`
	ValidateOutput bool     `yaml:"validate_output" json:"validate_output"`
	AllowedTypes   []string `yaml:"allowed_types" json:"allowed_types"`
	MaxSize        int64    `yaml:"max_size" json:"max_size"`
}

// DefaultConfig returns default middleware configuration
func DefaultConfig() *Config {
	return &Config{
		Cache: cacheconfig.Config{
			Enabled: true,
			TTL:     30 * time.Minute,
			Size:    100 * 1024 * 1024, // 100MB
		},
		Security: SecurityConfig{
			Enabled:     true,
			Strategy:    "jwt",
			JWTSecret:   "your-secret-key",
			TokenExpiry: "24h",
		},
		Monitoring: MonitoringConfig{
			Enabled:        true,
			Collector:      "prometheus",
			Port:           9090,
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			TracingEnabled: true,
		},
		HTTP: HTTPConfig{
			CORS: CORSConfig{
				Enabled:     true,
				Origins:     []string{"*"},
				Methods:     []string{"GET", "POST", "PUT", "DELETE"},
				Headers:     []string{"*"},
				Credentials: false,
				MaxAge:      86400,
			},
			RateLimit: RateLimitConfig{
				Enabled:           true,
				RequestsPerMinute: 100,
				BurstSize:         10,
				WindowSize:        time.Minute,
			},
			Compression: CompressionConfig{
				Enabled: true,
				Level:   6,
				Types:   []string{"text/plain", "text/html", "application/json"},
				MinSize: 1024,
			},
			Timeout: TimeoutConfig{
				Enabled:      true,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  120 * time.Second,
			},
		},
		Logging: LoggingConfig{
			Enabled:        true,
			Level:          "info",
			Format:         "json",
			IncludeBody:    false,
			IncludeHeaders: true,
			MaxBodySize:    1024,
		},
		Validation: ValidationConfig{
			Enabled:        true,
			ValidateInput:  true,
			ValidateOutput: false,
			AllowedTypes:   []string{"application/json", "application/xml"},
			MaxSize:        10 * 1024 * 1024, // 10MB
		},
	}
}

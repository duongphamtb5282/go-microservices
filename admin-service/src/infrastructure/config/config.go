package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string `yaml:"type"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"ssl_mode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime string `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime string `yaml:"conn_max_idle_time"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port string     `yaml:"port"`
	Host string     `yaml:"host"`
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled   bool   `yaml:"enabled"`
	JWTSecret string `yaml:"jwt_secret"`
	JWTIssuer string `yaml:"jwt_issuer"`
	APIKey    string `yaml:"api_key"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Determine environment
	env := getEnvironment()

	// Configure Viper
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./src/infrastructure/config")

	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v, env)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Try default config
		v.SetConfigName("config")
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading default config file: %w", err)
			}
		}
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate config
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	return env
}

func setDefaults(v *viper.Viper, environment string) {
	// Server defaults
	v.SetDefault("server.port", getEnvOrDefault("SERVER_PORT", "8086"))
	v.SetDefault("server.host", getEnvOrDefault("SERVER_HOST", "0.0.0.0"))

	// gRPC defaults
	v.SetDefault("grpc.port", getEnvOrDefault("GRPC_PORT", "50051"))
	v.SetDefault("grpc.host", getEnvOrDefault("GRPC_HOST", "0.0.0.0"))
	v.SetDefault("grpc.auth.enabled", getEnvBoolOrDefault("GRPC_AUTH_ENABLED", false))
	v.SetDefault("grpc.auth.jwt_secret", getEnvOrDefault("JWT_SECRET", ""))
	v.SetDefault("grpc.auth.jwt_issuer", getEnvOrDefault("JWT_ISSUER", "microservices"))
	v.SetDefault("grpc.auth.api_key", getEnvOrDefault("GRPC_API_KEY", ""))

	// Database defaults
	v.SetDefault("database.host", getEnvOrDefault("DATABASE_HOST", "localhost"))
	v.SetDefault("database.port", getEnvIntOrDefault("DATABASE_PORT", 5432))
	v.SetDefault("database.username", getEnvOrDefault("DATABASE_USERNAME", ""))
	v.SetDefault("database.password", getEnvOrDefault("DATABASE_PASSWORD", ""))
	v.SetDefault("database.name", getEnvOrDefault("DATABASE_NAME", "admin_service"))
	v.SetDefault("database.ssl_mode", getSSLModeForEnv(environment))
	v.SetDefault("database.max_open_conns", getEnvIntOrDefault("DATABASE_MAX_OPEN_CONNS", 25))
	v.SetDefault("database.max_idle_conns", getEnvIntOrDefault("DATABASE_MAX_IDLE_CONNS", 5))
	v.SetDefault("database.conn_max_lifetime", "1h")
	v.SetDefault("database.conn_max_idle_time", "30m")

	// Logging defaults
	v.SetDefault("logging.level", getEnvOrDefault("LOG_LEVEL", getLogLevelForEnv(environment)))
	v.SetDefault("logging.format", getEnvOrDefault("LOG_FORMAT", "json"))
	v.SetDefault("logging.output", getEnvOrDefault("LOG_OUTPUT", "stdout"))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getSSLModeForEnv(env string) string {
	switch env {
	case "production", "staging":
		return "require"
	default:
		return "disable"
	}
}

func getLogLevelForEnv(env string) string {
	switch env {
	case "production":
		return "info"
	case "staging":
		return "info"
	case "development":
		return "debug"
	default:
		return "info"
	}
}

func validateConfig(cfg *Config) error {
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.GRPC.Port == "" {
		return fmt.Errorf("grpc port is required")
	}
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	return nil
}

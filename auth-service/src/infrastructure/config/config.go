package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server        ServerConfig        `yaml:"server" mapstructure:"server"`
	Database      DatabaseConfig      `yaml:"database" mapstructure:"database"`
	Kafka         KafkaConfig         `yaml:"kafka" mapstructure:"kafka"`
	Logging       LoggingConfig       `yaml:"logging" mapstructure:"logging"`
	JWT           JWTConfig           `yaml:"jwt" mapstructure:"jwt"`
	Keycloak      KeycloakConfig      `yaml:"keycloak" mapstructure:"keycloak"`
	Authorization AuthorizationConfig `yaml:"authorization" mapstructure:"authorization"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `yaml:"port" mapstructure:"port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type                       string `yaml:"type" mapstructure:"type"`
	Host                       string `yaml:"host" mapstructure:"host"`
	Port                       int    `yaml:"port" mapstructure:"port"`
	Username                   string `yaml:"username" mapstructure:"username"`
	Password                   string `yaml:"password" mapstructure:"password"`
	Name                       string `yaml:"name" mapstructure:"name"`
	SSLMode                    string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	MaxOpenConns               int    `yaml:"max_connections" mapstructure:"max_connections"`
	MaxIdleConns               int    `yaml:"max_idle_connections" mapstructure:"max_idle_connections"`
	ConnMaxLifetime            string `yaml:"connection_max_lifetime" mapstructure:"connection_max_lifetime"`
	ConnMaxIdleTime            string `yaml:"connection_max_idle_time" mapstructure:"connection_max_idle_time"`
	QueryTimeout               string `yaml:"query_timeout" mapstructure:"query_timeout"`
	AcquireTimeout             string `yaml:"acquire_timeout" mapstructure:"acquire_timeout"`
	AcquireRetryAttempts       int    `yaml:"acquire_retry_attempts" mapstructure:"acquire_retry_attempts"`
	PreparedStatementCacheSize int    `yaml:"prepared_statement_cache_size" mapstructure:"prepared_statement_cache_size"`
	LogLevel                   string `yaml:"log_level" mapstructure:"log_level"`
	SlowThreshold              string `yaml:"slow_threshold" mapstructure:"slow_threshold"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string `yaml:"brokers" mapstructure:"brokers"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level" mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
	Output string `yaml:"output" mapstructure:"output"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret        string `yaml:"secret" mapstructure:"secret"`
	Expiry        string `yaml:"expiry" mapstructure:"expiry"`               // e.g., "24h", "15m"
	RefreshExpiry string `yaml:"refresh_expiry" mapstructure:"refresh_expiry"` // e.g., "168h" (7 days)
	Issuer        string `yaml:"issuer" mapstructure:"issuer"`
	Audience      string `yaml:"audience" mapstructure:"audience"`
	BcryptCost    int    `yaml:"bcrypt_cost" mapstructure:"bcrypt_cost"`
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	BaseURL         string       `yaml:"base_url" mapstructure:"base_url"`
	Realm           string       `yaml:"realm" mapstructure:"realm"`
	ClientID        string       `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret    string       `yaml:"client_secret" mapstructure:"client_secret"`
	RedirectURI     string       `yaml:"redirect_uri" mapstructure:"redirect_uri"`
	Scopes          []string     `yaml:"scopes" mapstructure:"scopes"`
	Timeout         string       `yaml:"timeout" mapstructure:"timeout"`
	RetryAttempts   int          `yaml:"retry_attempts" mapstructure:"retry_attempts"`
	CacheTTL        string       `yaml:"cache_ttl" mapstructure:"cache_ttl"`
	EnableSSO       bool         `yaml:"enable_sso" mapstructure:"enable_sso"`
	EnableMFA       bool         `yaml:"enable_mfa" mapstructure:"enable_mfa"`
	EnableRiskBased bool         `yaml:"enable_risk_based" mapstructure:"enable_risk_based"`
	SAML            SAMLConfig   `yaml:"saml" mapstructure:"saml"`
	OAuth           OAuthConfig  `yaml:"oauth" mapstructure:"oauth"`
	Policy          PolicyConfig `yaml:"policy" mapstructure:"policy"`
	Admin           AdminConfig  `yaml:"admin" mapstructure:"admin"`
}

// SAMLConfig holds SAML-specific configuration
type SAMLConfig struct {
	EntityID    string `yaml:"entity_id" mapstructure:"entity_id"`
	SSOURL      string `yaml:"sso_url" mapstructure:"sso_url"`
	SLOURL      string `yaml:"slo_url" mapstructure:"slo_url"`
	Certificate string `yaml:"certificate" mapstructure:"certificate"`
	PrivateKey  string `yaml:"private_key" mapstructure:"private_key"`
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
}

// OAuthConfig holds OAuth-specific configuration
type OAuthConfig struct {
	AuthorizationURL string   `yaml:"authorization_url" mapstructure:"authorization_url"`
	TokenURL         string   `yaml:"token_url" mapstructure:"token_url"`
	UserInfoURL      string   `yaml:"userinfo_url" mapstructure:"userinfo_url"`
	LogoutURL        string   `yaml:"logout_url" mapstructure:"logout_url"`
	JWKSURL          string   `yaml:"jwks_url" mapstructure:"jwks_url"`
	Scopes           []string `yaml:"scopes" mapstructure:"scopes"`
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
}

// PolicyConfig holds policy engine configuration
type PolicyConfig struct {
	PolicyURL   string `yaml:"policy_url" mapstructure:"policy_url"`
	DecisionURL string `yaml:"decision_url" mapstructure:"decision_url"`
	EnableRBAC  bool   `yaml:"enable_rbac" mapstructure:"enable_rbac"`
	Enabled     bool   `yaml:"enabled" mapstructure:"enabled"`
}

// AdminConfig holds Keycloak Admin API configuration
type AdminConfig struct {
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
}

// expandEnvInYAML expands environment variables in YAML content
// Supports ${VAR_NAME:default_value} syntax
func expandEnvInYAML(yamlContent []byte) []byte {
	content := string(yamlContent)

	// First, handle ${VAR:default} syntax
	re := regexp.MustCompile(`\$\{([^:}]+):([^}]+)\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			envVar := parts[1]
			defaultVal := parts[2]
			if value := os.Getenv(envVar); value != "" {
				return value
			}
			return defaultVal
		}
		return match
	})

	// Then handle regular ${VAR} syntax
	content = os.ExpandEnv(content)

	return []byte(content)
}

// readYAMLFileWithEnvExpansion reads a YAML file and expands environment variables
func readYAMLFileWithEnvExpansion(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	expanded := expandEnvInYAML(content)
	return expanded, nil
}

// Load loads configuration from YAML file and environment variables using Viper
func Load() (*Config, error) {
	// Initialize Viper
	v := viper.New()

	// Determine environment
	env := getEnvironment()

	// Configure Viper to read from environment-specific YAML file first
	v.SetConfigName(fmt.Sprintf("config.%s", env)) // e.g., config.development.yaml
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./src/infrastructure/config")

	// Enable reading from environment variables (these will override YAML values)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read environment-specific config file with environment variable expansion
	envConfigPath := fmt.Sprintf("./config/config.%s.yaml", env)
	expandedContent, err := readYAMLFileWithEnvExpansion(envConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try default config file
			expandedContent, err = readYAMLFileWithEnvExpansion("./config/config.yaml")
			if err != nil {
				return nil, fmt.Errorf("no config file found: neither %s nor config.yaml exist. Please ensure config files are present", envConfigPath)
			}
		} else {
			return nil, fmt.Errorf("error reading config file %s: %w", envConfigPath, err)
		}
	}

	// Read expanded YAML content into Viper
	if err := v.ReadConfig(strings.NewReader(string(expandedContent))); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Unmarshal into struct with custom decoder hooks for custom types
	var cfg Config
	
	// Create custom decoder with hooks for AuthorizationMode and IdentityProviderMode
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			// Custom hook for AuthorizationMode and IdentityProviderMode
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if t == reflect.TypeOf(AuthorizationMode("")) {
					if str, ok := data.(string); ok {
						return AuthorizationMode(str), nil
					}
				}
				if t == reflect.TypeOf(IdentityProviderMode("")) {
					if str, ok := data.(string); ok {
						return IdentityProviderMode(str), nil
					}
				}
				return data, nil
			},
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
		Result:           &cfg,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating decoder: %w", err)
	}

	// Decode from Viper's settings
	if err := decoder.Decode(v.AllSettings()); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Set defaults for authorization if not set
	if cfg.Authorization.Mode == "" {
		cfg.Authorization.Mode = AuthorizationModeJWT
	}
	if cfg.Authorization.IdentityProvider == "" {
		cfg.Authorization.IdentityProvider = IdentityProviderDatabase
	}

	// Validate JWT secret in production
	if err := validateProductionSecrets(env, &cfg); err != nil {
		return nil, err
	}

	// Validate configuration
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// validateProductionSecrets validates that required secrets are set in production
func validateProductionSecrets(env string, cfg *Config) error {
	if env == "production" {
		if cfg.JWT.Secret == "" || cfg.JWT.Secret == "your-secret-key-here-change-in-production" {
			return fmt.Errorf("JWT_SECRET must be set to a secure value in production")
		}
	}
	return nil
}

// getEnvironment determines the current environment
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Default to development
	}
	return env
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	// Validate server configuration
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Validate database configuration
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if cfg.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if cfg.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	// Validate logging configuration
	if cfg.Logging.Level == "" {
		return fmt.Errorf("logging level is required")
	}
	if cfg.Logging.Format == "" {
		return fmt.Errorf("logging format is required")
	}
	if cfg.Logging.Output == "" {
		return fmt.Errorf("logging output is required")
	}

	return nil
}

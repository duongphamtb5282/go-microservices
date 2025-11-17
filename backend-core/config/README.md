# Config Package

This package provides a comprehensive configuration management system with support for multiple configuration sources, validation, and environment-specific settings.

## Package Structure

```
config/
├── config.go           # Core configuration interface
├── database.go         # Database configuration
├── dsn_providers.go    # Data source name providers
├── logging.go          # Logging configuration
├── masking.go          # Sensitive data masking
├── mongo.go           # MongoDB configuration
├── postgres.go         # PostgreSQL configuration
├── redis.go            # Redis configuration
├── security.go         # Security configuration
└── zap.go              # Zap logger configuration
```

## Core Components

### 1. Configuration Interface (`config.go`)

The core configuration interface provides a unified way to access configuration values:

```go
type Config interface {
    // Basic operations
    Get(key string) interface{}
    GetString(key string) string
    GetInt(key string) int
    GetBool(key string) bool
    GetFloat64(key string) float64

    // Nested operations
    GetNested(key string) Config
    GetStringSlice(key string) []string
    GetStringMap(key string) map[string]interface{}

    // Default values
    GetStringWithDefault(key string, defaultValue string) string
    GetIntWithDefault(key string, defaultValue int) int
    GetBoolWithDefault(key string, defaultValue bool) bool

    // Validation
    Validate() error
    IsSet(key string) bool
}
```

### 2. Database Configuration (`database.go`)

Database configuration provides settings for database connections:

```go
type DatabaseConfig struct {
    Host            string        `yaml:"host" json:"host"`
    Port            int           `yaml:"port" json:"port"`
    User            string        `yaml:"user" json:"user"`
    Password        string        `yaml:"password" json:"password"`
    Name            string        `yaml:"name" json:"name"`
    SSLMode         string        `yaml:"ssl_mode" json:"ssl_mode"`
    MaxConns        int           `yaml:"max_conns" json:"max_conns"`
    MinConns        int           `yaml:"min_conns" json:"min_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`
    ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"`
}
```

### 3. Redis Configuration (`redis.go`)

Redis configuration provides settings for Redis connections:

```go
type RedisConfig struct {
    Host         string        `yaml:"host" json:"host"`
    Port         int           `yaml:"port" json:"port"`
    Password     string        `yaml:"password" json:"password"`
    DB           int           `yaml:"db" json:"db"`
    PoolSize     int           `yaml:"pool_size" json:"pool_size"`
    MinIdleConns int           `yaml:"min_idle_conns" json:"min_idle_conns"`
    MaxRetries   int           `yaml:"max_retries" json:"max_retries"`
    DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout"`
    ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
}
```

### 4. Logging Configuration (`logging.go`)

Logging configuration provides settings for structured logging:

```go
type LoggingConfig struct {
    Level      string `yaml:"level" json:"level"`
    Format     string `yaml:"format" json:"format"`
    Output     string `yaml:"output" json:"output"`
    File       string `yaml:"file" json:"file"`
    MaxSize    int    `yaml:"max_size" json:"max_size"`
    MaxBackups int    `yaml:"max_backups" json:"max_backups"`
    MaxAge     int    `yaml:"max_age" json:"max_age"`
    Compress   bool   `yaml:"compress" json:"compress"`
}
```

## Configuration Sources

### 1. Environment Variables

```go
// Load configuration from environment variables
config, err := NewConfigFromEnv()

// Access configuration values
dbHost := config.GetString("DB_HOST")
dbPort := config.GetInt("DB_PORT")
debug := config.GetBool("DEBUG")
```

### 2. YAML Files

```go
// Load configuration from YAML file
config, err := NewConfigFromFile("config.yaml")

// Access nested configuration
database := config.GetNested("database")
dbHost := database.GetString("host")
dbPort := database.GetInt("port")
```

### 3. JSON Files

```go
// Load configuration from JSON file
config, err := NewConfigFromFile("config.json")

// Access configuration values
apiKey := config.GetString("api.key")
timeout := config.GetInt("api.timeout")
```

### 4. Multiple Sources

```go
// Load configuration from multiple sources
config, err := NewConfig(
    WithFile("config.yaml"),
    WithEnv(),
    WithDefaults(defaultConfig),
)

// Sources are merged in order (later sources override earlier ones)
```

## Configuration Validation

### Basic Validation

```go
// Validate configuration
err := config.Validate()
if err != nil {
    log.Fatal("Configuration validation failed", err)
}
```

### Custom Validation

```go
// Implement custom validation
type CustomConfig struct {
    Database DatabaseConfig `yaml:"database"`
    Redis    RedisConfig    `yaml:"redis"`
}

func (c *CustomConfig) Validate() error {
    // Validate database configuration
    if c.Database.Host == "" {
        return errors.New("database host is required")
    }

    if c.Database.Port <= 0 || c.Database.Port > 65535 {
        return errors.New("database port must be between 1 and 65535")
    }

    // Validate Redis configuration
    if c.Redis.Host == "" {
        return errors.New("redis host is required")
    }

    return nil
}
```

### Validation Tags

```go
type DatabaseConfig struct {
    Host     string `yaml:"host" validate:"required"`
    Port     int    `yaml:"port" validate:"required,min=1,max=65535"`
    User     string `yaml:"user" validate:"required"`
    Password string `yaml:"password" validate:"required"`
    Name     string `yaml:"name" validate:"required"`
}
```

## Environment-Specific Configuration

### Development Configuration

```yaml
# config.dev.yaml
database:
  host: localhost
  port: 5432
  user: dev_user
  password: dev_password
  name: dev_db

redis:
  host: localhost
  port: 6379
  password: ""

logging:
  level: debug
  format: console
```

### Production Configuration

```yaml
# config.prod.yaml
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  name: ${DB_NAME}
  ssl_mode: require

redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  password: ${REDIS_PASSWORD}

logging:
  level: info
  format: json
  file: /var/log/app.log
```

### Loading Environment-Specific Config

```go
// Load environment-specific configuration
env := os.Getenv("ENVIRONMENT")
if env == "" {
    env = "development"
}

configFile := fmt.Sprintf("config.%s.yaml", env)
config, err := NewConfigFromFile(configFile)
```

## Sensitive Data Masking

### Configuration Masking

```go
// Create configuration with masking
config := NewConfigWithMasking(
    WithFile("config.yaml"),
    WithMaskedFields("password", "secret", "key"),
)

// Sensitive values are masked in logs
log.Info("Database config", "config", config.GetNested("database"))
```

### Custom Masking

```go
// Implement custom masking
type MaskedConfig struct {
    *BaseConfig
    maskedFields []string
}

func (c *MaskedConfig) GetString(key string) string {
    value := c.BaseConfig.GetString(key)

    // Check if field should be masked
    for _, field := range c.maskedFields {
        if strings.Contains(key, field) {
            return "***"
        }
    }

    return value
}
```

## DSN (Data Source Name) Providers

### PostgreSQL DSN

```go
// Create PostgreSQL DSN
dsnProvider := NewPostgreSQLDSNProvider()
dsn, err := dsnProvider.GetDSN(&DatabaseConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Name:     "mydb",
    SSLMode:  "disable",
})
// Result: "postgres://postgres:password@localhost:5432/mydb?sslmode=disable"
```

### MongoDB DSN

```go
// Create MongoDB DSN
dsnProvider := NewMongoDSNProvider()
dsn, err := dsnProvider.GetDSN(&MongoConfig{
    Host:     "localhost",
    Port:     27017,
    User:     "mongo",
    Password: "password",
    Database: "mydb",
})
// Result: "mongodb://mongo:password@localhost:27017/mydb"
```

### Redis DSN

```go
// Create Redis DSN
dsnProvider := NewRedisDSNProvider()
dsn, err := dsnProvider.GetDSN(&RedisConfig{
    Host:     "localhost",
    Port:     6379,
    Password: "password",
    DB:       0,
})
// Result: "redis://:password@localhost:6379/0"
```

## Configuration Hot Reloading

### File Watching

```go
// Create configuration with file watching
config, err := NewConfigWithWatcher(
    WithFile("config.yaml"),
    WithWatchInterval(5*time.Second),
    WithReloadCallback(func(newConfig Config) {
        log.Info("Configuration reloaded")
        // Update application settings
    }),
)
```

### Manual Reload

```go
// Reload configuration manually
err := config.Reload()
if err != nil {
    log.Error("Failed to reload configuration", err)
}
```

## Configuration Examples

### Complete Application Configuration

```yaml
# config.yaml
app:
  name: "My Application"
  version: "1.0.0"
  environment: "production"

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "mydb"
  ssl_mode: "disable"
  max_conns: 25
  min_conns: 5

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file: "/var/log/app.log"

security:
  jwt_secret: "your-secret-key"
  token_expiry: "24h"
  bcrypt_cost: 12
```

### Using Configuration in Application

```go
// Load configuration
config, err := NewConfigFromFile("config.yaml")
if err != nil {
    log.Fatal("Failed to load configuration", err)
}

// Validate configuration
err = config.Validate()
if err != nil {
    log.Fatal("Configuration validation failed", err)
}

// Access configuration values
appName := config.GetString("app.name")
serverPort := config.GetInt("server.port")
dbHost := config.GetString("database.host")
redisHost := config.GetString("redis.host")

// Create database configuration
dbConfig := &DatabaseConfig{
    Host:     config.GetString("database.host"),
    Port:     config.GetInt("database.port"),
    User:     config.GetString("database.user"),
    Password: config.GetString("database.password"),
    Name:     config.GetString("database.name"),
    SSLMode:  config.GetString("database.ssl_mode"),
    MaxConns: config.GetInt("database.max_conns"),
    MinConns: config.GetInt("database.min_conns"),
}

// Create Redis configuration
redisConfig := &RedisConfig{
    Host:     config.GetString("redis.host"),
    Port:     config.GetInt("redis.port"),
    Password: config.GetString("redis.password"),
    DB:       config.GetInt("redis.db"),
    PoolSize: config.GetInt("redis.pool_size"),
}
```

## Best Practices

### 1. Configuration Structure

- Use nested configuration for related settings
- Group settings by functionality
- Use consistent naming conventions
- Document configuration options

### 2. Environment Variables

- Use environment variables for sensitive data
- Provide sensible defaults
- Use consistent naming (e.g., `DB_HOST`, `REDIS_PORT`)
- Document required environment variables

### 3. Validation

- Validate all configuration values
- Provide meaningful error messages
- Check for required fields
- Validate ranges and formats

### 4. Security

- Mask sensitive values in logs
- Use secure defaults
- Validate input from external sources
- Keep secrets out of version control

### 5. Testing

- Test configuration loading
- Test validation logic
- Test environment-specific configs
- Mock configuration for unit tests

## Migration Guide

When upgrading configuration:

1. **Check breaking changes** in the changelog
2. **Update configuration files** if needed
3. **Test configuration loading** in all environments
4. **Update documentation** for new options
5. **Monitor application startup** for configuration errors

## Future Enhancements

- **Configuration encryption** - Encrypt sensitive configuration values
- **Remote configuration** - Load configuration from remote sources
- **Configuration templates** - Template-based configuration generation
- **Configuration diffing** - Compare configuration changes
- **Configuration backup** - Backup and restore configuration

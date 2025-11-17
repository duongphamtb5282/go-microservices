# Configuration System

This directory contains configuration files for the auth-service. The configuration system uses a hierarchical approach with environment-specific files and environment variable overrides.

## 📁 Configuration Files

### Available Configurations

- **`config.yaml`** - Base configuration with sensible defaults
- **`config.development.yaml`** - Development environment overrides
- **`config.prod.yaml`** - Production environment overrides

### Configuration Structure

```
config/
├── config.yaml              # Base configuration
├── config.development.yaml  # Dev overrides
├── config.prod.yaml         # Production overrides
├── README.md               # This file
└── pingam-mock/            # PingAM mock data for testing
    ├── initializerJson.json
    └── mockserver.properties
```

## 🔧 Configuration Loading Order

The configuration system loads settings in the following priority order (highest to lowest):

1. **Environment Variables** (highest priority)
2. **Environment-specific YAML file** (e.g., `config.development.yaml`)
3. **Default YAML file** (`config.yaml`)
4. **Hardcoded defaults** (lowest priority)

## 🌍 Environment Detection

The system automatically detects the environment using the `APP_ENV` environment variable:

```bash
# Development (default)
export APP_ENV=development
# or
export APP_ENV=dev

# Production
export APP_ENV=production
# or
export APP_ENV=prod
```

**Supported values:**

- `development` or `dev` → Uses `config.development.yaml`
- `production` or `prod` → Uses `config.prod.yaml`
- If not set → Defaults to `development`

## 📝 Configuration Structure

### Server Configuration

```yaml
server:
  port: "8080"
  host: "localhost"
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
```

### Database Configuration

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "P$ssword9"
  database: "auth_db"
  ssl_mode: "disable"
  max_connections: 10
  max_idle_connections: 5
  connection_max_lifetime: "1h"
```

### Cache Configuration

```yaml
cache:
  name: "auth-cache"
  addr: "localhost:6379"
  password: ""
  db: 0
  use_cluster: false
  cluster_addrs: []
  pool_size: 10
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: "5s"
  read_timeout: "3s"
  write_timeout: "3s"
```

### Logging Configuration

```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file:
    enabled: false
    path: "logs/auth-service.log"
    max_size: 100
    max_backups: 3
    max_age: 28
    compress: true
```

## 🔄 Environment Variable Overrides

You can override any configuration value using environment variables. The system automatically converts YAML keys to environment variable names:

### Examples

```bash
# Override server port
export SERVER_PORT=9090

# Override database host
export DATABASE_HOST=prod-db.example.com

# Override cache address
export CACHE_ADDR=prod-cache.example.com:6379

# Override logging level
export LOGGING_LEVEL=debug
```

### Nested Configuration

For nested configuration, use underscores:

```bash
# Override nested logging file settings
export LOGGING_FILE_ENABLED=true
export LOGGING_FILE_PATH=/var/log/auth-service.log
export LOGGING_FILE_MAX_SIZE=200
```

## 🚀 Usage Examples

### Development

```bash
# Uses config.development.yaml
export APP_ENV=development
go run main.go
```

### Production

```bash
# Uses config.prod.yaml with environment overrides
export APP_ENV=production
export DATABASE_HOST=prod-db.example.com
export DATABASE_PASSWORD=secure-password
export JWT_SECRET=your-very-secure-jwt-secret
export CACHE_ADDR=prod-cache.example.com:6379
go run main.go
```

## 🔒 Security Considerations

### Production Configuration

- **Never commit sensitive data** to YAML files
- **Use environment variables** for secrets in production
- **Use placeholder values** in YAML files (e.g., `${DB_PASSWORD}`)

### Example Production Setup

```yaml
# config.prod.yaml
database:
  password: "${DATABASE_PASSWORD}" # Use environment variable
  ssl_mode: "require" # Require SSL in production

jwt:
  secret: "${JWT_SECRET}" # Use environment variable
```

```bash
# Set secrets via environment variables
export DATABASE_PASSWORD=secure-database-password
export JWT_SECRET=very-secure-jwt-secret
export CACHE_PASSWORD=redis-password
```

## 🔐 Required Production Environment Variables

When running in production, the following environment variables **MUST** be set:

```bash
# Database
export DATABASE_PASSWORD="your-secure-password"

# JWT
export JWT_SECRET="your-very-secure-jwt-secret-min-32-chars"

# Optional: Override other settings
export DATABASE_HOST="your-db-host"
export DATABASE_NAME="auth_service"
export CACHE_ADDR="your-redis:6379"
export CACHE_PASSWORD="redis-password"
```

## 📊 Configuration Validation

The configuration system includes validation:

- **Required fields**: Validates that required fields are present
- **Type validation**: Ensures correct data types
- **Range validation**: Validates numeric ranges
- **Format validation**: Validates email formats, URLs, etc.

## 🔧 Custom Configuration

### Adding New Configuration

1. **Update the Config struct** in `src/infrastructure/config/config.go`
2. **Add to base config** in `config/config.yaml`
3. **Add environment-specific overrides** if needed
4. **Update documentation** in this README
5. **Add validation** if required

### Example: Adding Email Configuration

```go
// src/infrastructure/config/config.go
type Config struct {
    // ... existing fields
    Email EmailConfig `yaml:"email"`
}

type EmailConfig struct {
    SMTPHost     string `yaml:"smtp_host"`
    SMTPPort     int    `yaml:"smtp_port"`
    SMTPUsername string `yaml:"smtp_username"`
    SMTPPassword string `yaml:"smtp_password"`
}
```

```yaml
# config.yaml
email:
  smtp_host: "localhost"
  smtp_port: 587
  smtp_username: "noreply@example.com"
  smtp_password: "${EMAIL_PASSWORD}" # Use env var for password
```

## 🐛 Troubleshooting

### Common Issues

1. **Config file not found**

   - Check file paths in `AddConfigPath()`
   - Verify file exists in the correct directory

2. **Environment variables not working**

   - Check `SetEnvKeyReplacer()` configuration
   - Verify environment variable names match the pattern

3. **Type conversion errors**
   - Check struct tags (`mapstructure`, `yaml`)
   - Verify data types in YAML files

### Debug Configuration

```go
// Enable debug logging to see configuration loading
v.SetDebug(true)
```

## 📚 Best Practices

1. **Use environment-specific files** for different environments
2. **Use environment variables** for ALL sensitive data (passwords, secrets, tokens)
3. **Never commit secrets** to version control
4. **Validate configuration** on startup
5. **Use meaningful default values** in base config
6. **Keep production config minimal** - only override what's necessary
7. **Use consistent naming conventions** (snake_case for YAML keys)
8. **Test configuration loading** in different environments
9. **Document all configuration options** in this README
10. **Use `${ENV_VAR}` syntax** for environment variable substitution in YAML

## 🎯 Configuration Priority

The configuration system merges settings in this order (last wins):

1. **Base config** (`config.yaml`)
2. **Environment config** (`config.development.yaml` or `config.prod.yaml`)
3. **Environment variables** (highest priority)

## 📝 Example: Complete Setup

### Development

```bash
# .env.development
APP_ENV=development
DATABASE_PASSWORD=dev_password
JWT_SECRET=dev-secret-key-for-testing-only
```

### Production

```bash
# Set via environment or secrets manager
export APP_ENV=production
export DATABASE_HOST=prod-db.example.com
export DATABASE_PASSWORD=secure-prod-password
export JWT_SECRET=very-secure-jwt-secret-min-32-chars
export CACHE_ADDR=prod-redis:6379
export CACHE_PASSWORD=redis-secure-password
```

# Wire Package

This package provides a comprehensive dependency injection system using Google Wire, with support for automatic dependency resolution, provider registration, and service lifecycle management.

## Package Structure

```
wire/
├── application.go          # Application-level providers
├── config.go               # Configuration providers
├── generic_providers.go    # Generic service providers
├── injectors.go            # Wire injectors
├── middleware.go           # Middleware providers
├── providers.go            # Core providers
├── service_providers.go    # Service-specific providers
├── examples/               # Usage examples
│   ├── auth_service.go    # Authentication service example
│   ├── generic_service.go # Generic service example
│   ├── payment_service.go  # Payment service example
│   └── user_service.go    # User service example
├── README.md              # Main documentation
├── USAGE_GUIDE.md         # Detailed usage guide
└── abstract_solution.md   # Abstract solution documentation
```

## Core Components

### 1. Provider Functions

Provider functions are the building blocks of dependency injection:

```go
// Database provider
func ProvideDatabase(config *config.DatabaseConfig) (Database, error) {
    return database.NewDatabase(config)
}

// Cache provider
func ProvideCache(config *config.CacheConfig) (Cache, error) {
    return cache.NewCache(config)
}

// Logger provider
func ProvideLogger(config *config.LoggingConfig) (Logger, error) {
    return logging.NewLogger(config)
}

// Service provider
func ProvideUserService(db Database, cache Cache, logger Logger) *UserService {
    return &UserService{
        db:     db,
        cache:  cache,
        logger: logger,
    }
}
```

### 2. Wire Sets

Wire sets group related providers together:

```go
// Database providers set
var DatabaseProviders = wire.NewSet(
    ProvideDatabase,
    ProvideDatabaseConfig,
)

// Cache providers set
var CacheProviders = wire.NewSet(
    ProvideCache,
    ProvideCacheConfig,
)

// Logging providers set
var LoggingProviders = wire.NewSet(
    ProvideLogger,
    ProvideLoggingConfig,
)

// Core providers set
var CoreProviders = wire.NewSet(
    DatabaseProviders,
    CacheProviders,
    LoggingProviders,
)
```

### 3. Injectors

Injectors define the dependency graph and provide entry points:

```go
// Application injector
func InitializeApplication() (*Application, error) {
    wire.Build(
        CoreProviders,
        ProvideApplication,
    )
    return &Application{}, nil
}

// Service injector
func InitializeUserService() (*UserService, error) {
    wire.Build(
        CoreProviders,
        ProvideUserService,
    )
    return &UserService{}, nil
}
```

## Configuration Providers

### Database Configuration

```go
// Database configuration provider
func ProvideDatabaseConfig() *config.DatabaseConfig {
    return &config.DatabaseConfig{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnvInt("DB_PORT", 5432),
        User:     getEnv("DB_USER", "postgres"),
        Password: getEnv("DB_PASSWORD", ""),
        Name:     getEnv("DB_NAME", "mydb"),
        SSLMode:  getEnv("DB_SSL_MODE", "disable"),
    }
}

// Database provider
func ProvideDatabase(config *config.DatabaseConfig) (Database, error) {
    return database.NewDatabase(config)
}
```

### Cache Configuration

```go
// Cache configuration provider
func ProvideCacheConfig() *config.CacheConfig {
    return &config.CacheConfig{
        Host:     getEnv("CACHE_HOST", "localhost"),
        Port:     getEnvInt("CACHE_PORT", 6379),
        Password: getEnv("CACHE_PASSWORD", ""),
        DB:       getEnvInt("CACHE_DB", 0),
    }
}

// Cache provider
func ProvideCache(config *config.CacheConfig) (Cache, error) {
    return cache.NewCache(config)
}
```

### Logging Configuration

```go
// Logging configuration provider
func ProvideLoggingConfig() *config.LoggingConfig {
    return &config.LoggingConfig{
        Level:  getEnv("LOG_LEVEL", "info"),
        Format: getEnv("LOG_FORMAT", "json"),
        Output: getEnv("LOG_OUTPUT", "stdout"),
    }
}

// Logger provider
func ProvideLogger(config *config.LoggingConfig) (Logger, error) {
    return logging.NewLogger(config)
}
```

## Service Providers

### User Service

```go
// User repository provider
func ProvideUserRepository(db Database) UserRepository {
    return &UserRepositoryImpl{db: db}
}

// User service provider
func ProvideUserService(
    repo UserRepository,
    cache Cache,
    logger Logger,
) *UserService {
    return &UserService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}

// User service providers set
var UserServiceProviders = wire.NewSet(
    ProvideUserRepository,
    ProvideUserService,
)
```

### Authentication Service

```go
// JWT service provider
func ProvideJWTService(config *config.SecurityConfig) JWTService {
    return &JWTServiceImpl{
        secret: config.JWTSecret,
        expiry: config.JWTExpiry,
    }
}

// Auth service provider
func ProvideAuthService(
    userService *UserService,
    jwtService JWTService,
    logger Logger,
) *AuthService {
    return &AuthService{
        userService: userService,
        jwtService:  jwtService,
        logger:      logger,
    }
}

// Auth service providers set
var AuthServiceProviders = wire.NewSet(
    ProvideJWTService,
    ProvideAuthService,
)
```

## Application Providers

### Application Structure

```go
// Application struct
type Application struct {
    Config      *config.Config
    Database    Database
    Cache       Cache
    Logger      Logger
    UserService *UserService
    AuthService *AuthService
}

// Application provider
func ProvideApplication(
    config *config.Config,
    db Database,
    cache Cache,
    logger Logger,
    userService *UserService,
    authService *AuthService,
) *Application {
    return &Application{
        Config:      config,
        Database:    db,
        Cache:       cache,
        Logger:      logger,
        UserService: userService,
        AuthService: authService,
    }
}
```

### Application Initialization

```go
// Application injector
func InitializeApplication() (*Application, error) {
    wire.Build(
        // Configuration providers
        ProvideConfig,
        ProvideDatabaseConfig,
        ProvideCacheConfig,
        ProvideLoggingConfig,

        // Core providers
        ProvideDatabase,
        ProvideCache,
        ProvideLogger,

        // Service providers
        UserServiceProviders,
        AuthServiceProviders,

        // Application provider
        ProvideApplication,
    )
    return &Application{}, nil
}
```

## Middleware Providers

### HTTP Middleware

```go
// Logging middleware provider
func ProvideLoggingMiddleware(logger Logger) *LoggingMiddleware {
    return &LoggingMiddleware{logger: logger}
}

// Cache middleware provider
func ProvideCacheMiddleware(cache Cache) *CacheMiddleware {
    return &CacheMiddleware{cache: cache}
}

// Security middleware provider
func ProvideSecurityMiddleware(config *config.SecurityConfig) *SecurityMiddleware {
    return &SecurityMiddleware{config: config}
}

// Middleware providers set
var MiddlewareProviders = wire.NewSet(
    ProvideLoggingMiddleware,
    ProvideCacheMiddleware,
    ProvideSecurityMiddleware,
)
```

### Middleware Chain

```go
// Middleware chain provider
func ProvideMiddlewareChain(
    logging *LoggingMiddleware,
    cache *CacheMiddleware,
    security *SecurityMiddleware,
) *MiddlewareChain {
    return &MiddlewareChain{
        middlewares: []Middleware{
            security,
            logging,
            cache,
        },
    }
}
```

## Generic Providers

### Generic Service Provider

```go
// Generic service provider
func ProvideGenericService[T any](
    repo Repository[T],
    cache Cache,
    logger Logger,
) *GenericService[T] {
    return &GenericService[T]{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}

// Generic repository provider
func ProvideGenericRepository[T any](db Database) Repository[T] {
    return &GenericRepository[T]{db: db}
}
```

### Generic Service Set

```go
// Generic service providers set
func GenericServiceProviders[T any]() wire.ProviderSet {
    return wire.NewSet(
        ProvideGenericRepository[T],
        ProvideGenericService[T],
    )
}
```

## Error Handling

### Provider Error Handling

```go
// Provider with error handling
func ProvideDatabase(config *config.DatabaseConfig) (Database, error) {
    db, err := database.NewDatabase(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create database: %w", err)
    }

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

### Injector Error Handling

```go
// Injector with error handling
func InitializeApplication() (*Application, error) {
    wire.Build(
        CoreProviders,
        ServiceProviders,
        ProvideApplication,
    )
    return &Application{}, nil
}

// Usage with error handling
func main() {
    app, err := InitializeApplication()
    if err != nil {
        log.Fatal("Failed to initialize application", err)
    }

    // Use application
    app.Start()
}
```

## Testing

### Unit Testing

```go
// Test provider
func TestProvideUserService(t *testing.T) {
    // Create test dependencies
    mockRepo := &MockUserRepository{}
    mockCache := &MockCache{}
    mockLogger := &MockLogger{}

    // Create service
    service := ProvideUserService(mockRepo, mockCache, mockLogger)

    // Test service
    assert.NotNil(t, service)
    assert.Equal(t, mockRepo, service.repo)
    assert.Equal(t, mockCache, service.cache)
    assert.Equal(t, mockLogger, service.logger)
}
```

### Integration Testing

```go
// Test injector
func TestInitializeApplication(t *testing.T) {
    // Set test environment variables
    os.Setenv("DB_HOST", "localhost")
    os.Setenv("DB_PORT", "5432")
    os.Setenv("CACHE_HOST", "localhost")
    os.Setenv("CACHE_PORT", "6379")

    // Initialize application
    app, err := InitializeApplication()
    require.NoError(t, err)
    require.NotNil(t, app)

    // Test application components
    assert.NotNil(t, app.Database)
    assert.NotNil(t, app.Cache)
    assert.NotNil(t, app.Logger)
    assert.NotNil(t, app.UserService)
    assert.NotNil(t, app.AuthService)
}
```

## Configuration Management

### Environment-Based Configuration

```go
// Environment configuration provider
func ProvideConfig() *config.Config {
    return &config.Config{
        Environment: getEnv("ENVIRONMENT", "development"),
        Debug:       getEnvBool("DEBUG", false),
        Database:    ProvideDatabaseConfig(),
        Cache:       ProvideCacheConfig(),
        Logging:     ProvideLoggingConfig(),
    }
}

// Environment variable helpers
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}
```

### File-Based Configuration

```go
// File configuration provider
func ProvideConfigFromFile(filename string) (*config.Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }

    var cfg config.Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return &cfg, nil
}
```

## Service Lifecycle

### Service Initialization

```go
// Service with initialization
func ProvideUserService(
    repo UserRepository,
    cache Cache,
    logger Logger,
) *UserService {
    service := &UserService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }

    // Initialize service
    if err := service.Initialize(); err != nil {
        logger.Error("Failed to initialize user service", err)
        return nil
    }

    return service
}
```

### Service Cleanup

```go
// Service with cleanup
func ProvideDatabase(config *config.DatabaseConfig) (Database, func(), error) {
    db, err := database.NewDatabase(config)
    if err != nil {
        return nil, nil, err
    }

    cleanup := func() {
        if err := db.Close(); err != nil {
            log.Error("Failed to close database", err)
        }
    }

    return db, cleanup, nil
}
```

## Best Practices

### 1. Provider Design

- Keep providers simple and focused
- Handle errors appropriately
- Use meaningful provider names
- Document provider dependencies

### 2. Wire Sets

- Group related providers together
- Use descriptive set names
- Keep sets focused and cohesive
- Avoid circular dependencies

### 3. Error Handling

- Always handle provider errors
- Provide meaningful error messages
- Use error wrapping for context
- Log errors appropriately

### 4. Testing

- Test providers in isolation
- Mock external dependencies
- Test error scenarios
- Use table-driven tests

### 5. Configuration

- Use environment variables for configuration
- Provide sensible defaults
- Validate configuration values
- Document configuration options

## Examples

### Complete Application Example

```go
// main.go
func main() {
    // Initialize application
    app, err := InitializeApplication()
    if err != nil {
        log.Fatal("Failed to initialize application", err)
    }

    // Start application
    if err := app.Start(); err != nil {
        log.Fatal("Failed to start application", err)
    }

    // Wait for shutdown signal
    <-app.Shutdown()
}

// wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "backend-core/wire"
)

func InitializeApplication() (*Application, error) {
    wire.Build(
        wire.CoreProviders,
        wire.ServiceProviders,
        wire.MiddlewareProviders,
        ProvideApplication,
    )
    return &Application{}, nil
}
```

### Service Example

```go
// user_service.go
type UserService struct {
    repo   UserRepository
    cache  Cache
    logger Logger
}

func (s *UserService) GetUser(id string) (*User, error) {
    // Check cache first
    var user User
    if err := s.cache.Get("user:"+id, &user); err == nil {
        return &user, nil
    }

    // Get from repository
    user, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }

    // Cache result
    s.cache.Set("user:"+id, user, 5*time.Minute)

    return &user, nil
}

// wire.go
func ProvideUserService(
    repo UserRepository,
    cache Cache,
    logger Logger,
) *UserService {
    return &UserService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

## Migration Guide

When upgrading Wire:

1. **Check breaking changes** in the changelog
2. **Update provider signatures** if needed
3. **Test dependency injection** in all environments
4. **Update wire sets** if using new features
5. **Monitor performance** after upgrade

## Future Enhancements

- **Dynamic providers** - Runtime provider registration
- **Provider lifecycle** - Advanced lifecycle management
- **Provider monitoring** - Provider performance monitoring
- **Provider validation** - Automatic provider validation
- **Provider documentation** - Automatic provider documentation generation

# Backend Core

A comprehensive shared module for Go microservices that provides database abstraction, caching, logging, security, and messaging capabilities.

## Features

- **Database Abstraction**: Support for multiple SQL and NoSQL databases with GORM
- **Connection Pooling**: Optimized database connection management
- **Migration System**: Database schema migration management
- **Transaction Support**: ACID-compliant transaction handling
- **Caching**: Redis and in-memory caching with advanced operations
- **Logging**: Structured logging with correlation IDs
- **Security**: JWT authentication and password hashing
- **Type Safety**: Generic repository pattern with compile-time type checking

## Supported Databases

### SQL Databases (via GORM)

- MySQL
- PostgreSQL
- SQLite
- SQL Server

### NoSQL Databases

- MongoDB

## Installation

```bash
go get backend-core
```

## Quick Start

### 1. Configuration

Create a `config.yaml` file:

```yaml
database:
  type: mysql
  host: localhost
  port: 3306
  database: microservices
  username: root
  password: password
  max_open_conns: 100
  max_idle_conns: 10

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

logging:
  level: info
  format: json
  output: stdout

security:
  jwt_secret: your-secret-key
  jwt_expiration: 24h
  issuer: microservices
  audience: microservices-clients
  bcrypt_cost: 12
```

### 2. Basic Usage

```go
package main

import (
    "context"
    "log"

    "backend-core/config"
    "backend-core/database"
    "backend-core/logging"
    "backend-core/security"
)

type User struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"size:100;not null;uniqueIndex" json:"username"`
    Email    string `gorm:"size:100;not null;uniqueIndex" json:"email"`
    Password string `gorm:"size:255;not null" json:"-"`
}

func main() {
    // Load configuration
    cfg, err := config.LoadConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Initialize logger
    logger := logging.NewLogger(cfg.Logging.Level, cfg.Logging.Format)

    // Create database
    factory := database.NewDatabaseFactory()
    db, err := factory.CreateDatabase(database.MySQL, &cfg.Database)
    if err != nil {
        log.Fatal(err)
    }

    // Connect
    ctx := context.Background()
    if err := db.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer db.Disconnect(ctx)

    // Get repository
    userRepo := db.GetRepository[User]()

    // Create user
    user := &User{
        Username: "john_doe",
        Email:    "john@example.com",
        Password: "hashed_password",
    }

    if err := userRepo.Create(ctx, user); err != nil {
        log.Fatal(err)
    }

    logger.Info("User created successfully", "user_id", user.ID)
}
```

## DSN Provider Interface

The `DsnProvider` interface provides a generic way to generate database connection strings:

```go
type DsnProvider interface {
    Dsn() string
}
```

### Built-in DSN Providers

The library includes built-in DSN providers for supported databases:

```go
// PostgreSQL DSN Provider
postgresProvider := config.NewDSNProvider("postgresql", &cfg.Database)
dsn := postgresProvider.Dsn()

// MongoDB DSN Provider
mongoProvider := config.NewDSNProvider("mongodb", &cfg.Database)
dsn := mongoProvider.Dsn()
```

### Custom DSN Providers

You can create custom DSN providers for specialized connection requirements:

```go
type CustomDSNProvider struct {
    Host     string
    Port     int
    Database string
    Username string
    Password string
    CustomParam string
}

func (c *CustomDSNProvider) Dsn() string {
    return fmt.Sprintf("custom://%s:%s@%s:%d/%s?custom_param=%s",
        c.Username, c.Password, c.Host, c.Port, c.Database, c.CustomParam)
}

// Use custom provider
customProvider := &CustomDSNProvider{...}
dsn := customProvider.Dsn()
```

### Advanced DSN Configuration

DSN providers support additional parameters:

```go
// PostgreSQL with custom parameters
postgresProvider := &config.PostgreSQLDSNProvider{
    Host:     "localhost",
    Port:     5432,
    Database: "test_db",
    Username: "user",
    Password: "pass",
    SSLMode:  "disable",
    Params: map[string]string{
        "connect_timeout": "30",
        "statement_timeout": "30000",
        "idle_in_transaction_session_timeout": "300000",
    },
}

// PostgreSQL with SSL
postgresProvider := &config.PostgreSQLDSNProvider{
    Host:     "localhost",
    Port:     5432,
    Database: "test_db",
    Username: "user",
    Password: "pass",
    SSLMode:  "require",
    Params: map[string]string{
        "sslmode":     "require",
        "sslcert":     "/path/to/client-cert.pem",
        "sslkey":      "/path/to/client-key.pem",
        "sslrootcert": "/path/to/ca-cert.pem",
    },
}
```

## Database Operations

### Repository Pattern

```go
// Get repository for any entity type
userRepo := db.GetRepository[User]()

// Create
user := &User{Username: "john", Email: "john@example.com"}
err := userRepo.Create(ctx, user)

// Read
user, err := userRepo.GetByID(ctx, 1)
user, err := userRepo.GetByField(ctx, "email", "john@example.com")

// Update
user.Username = "john_updated"
err := userRepo.Update(ctx, user)

// Delete
err := userRepo.Delete(ctx, 1)

// Query with pagination
users, err := userRepo.GetAll(ctx, database.Filter{}, database.Pagination{
    Page:     1,
    PageSize: 10,
})

// Complex queries
query := database.Query{
    Filter: database.Filter{"role": "admin"},
    Pagination: database.Pagination{Page: 1, PageSize: 10},
    OrderBy: "created_at",
    Order:   "desc",
}
users, err := userRepo.Find(ctx, query)
```

### Transactions

```go
err := userRepo.WithTransaction(ctx, func(txRepo database.Repository[User]) error {
    // Create user
    user := &User{Username: "john", Email: "john@example.com"}
    if err := txRepo.Create(ctx, user); err != nil {
        return err
    }

    // Update user
    user.Role = "admin"
    if err := txRepo.Update(ctx, user); err != nil {
        return err
    }

    return nil
})
```

## Caching

### Redis Cache

```go
import "backend-core/cache"

// Create Redis cache
redisCache := cache.NewRedisCache("localhost", 6379, "", 0)
cacheManager := cache.NewCacheManager(redisCache)

// Basic operations
cacheManager.Put(ctx, "key", "value", 1*time.Hour)
var value string
cacheManager.Get(ctx, "key", &value)

// Counter operations
cacheManager.Increment(ctx, "counter")
counter, _ := cacheManager.Get(ctx, "counter", &counter)

// List operations
cacheManager.ListPush(ctx, "list", "item1")
cacheManager.ListPush(ctx, "list", "item2")

// Set operations
cacheManager.SetAdd(ctx, "set", "member1")
cacheManager.SetAdd(ctx, "set", "member2")

// Hash operations
cacheManager.HashSet(ctx, "hash", "field1", "value1")
cacheManager.HashGet(ctx, "hash", "field1", &value)
```

## Security

### JWT Authentication

```go
import "backend-core/security"

// Initialize JWT manager
jwtManager := security.NewJWTManager(
    "your-secret-key",
    24*time.Hour,
    "microservices",
    "microservices-clients",
)

// Generate token
token, err := jwtManager.GenerateToken("user123", "john_doe", "user")

// Validate token
claims, err := jwtManager.ValidateToken(token)
```

### Password Hashing

```go
// Initialize auth manager
authManager := security.NewAuthManager(jwtManager, 12)

// Hash password
hashedPassword, err := authManager.HashPassword("password123")

// Verify password
err := authManager.VerifyPassword(hashedPassword, "password123")
```

## Logging

### Structured Logging

```go
import "backend-core/logging"

// Create logger
logger := logging.NewLogger("info", "json")

// Basic logging
logger.Info("Application started")
logger.Error("Something went wrong", "error", err)

// With context
ctx := context.WithValue(context.Background(), "user_id", "123")
logger.WithContext(ctx).Info("User action", "action", "login")

// With fields
logger.WithFields(logrus.Fields{
    "user_id": "123",
    "action":  "login",
}).Info("User logged in")
```

## Migration System

### Creating Migrations

```go
import "backend-core/database/gorm"

// Create migration
migration := gorm.NewMigration(
    1,
    "Create users table",
    func(ctx context.Context, db *gorm.DB) error {
        return db.AutoMigrate(&User{})
    },
    func(ctx context.Context, db *gorm.DB) error {
        return db.Migrator().DropTable(&User{})
    },
)

// Add to migration manager
migrationManager := gorm.NewMigrationManager(db)
migrationManager.AddMigration(migration)

// Run migrations
err := migrationManager.RunMigrations(ctx)
```

## Connection Pooling

### Database Statistics

```go
// Get connection statistics
stats := db.GetStats()
fmt.Printf("Open connections: %d\n", stats.OpenConnections)
fmt.Printf("Idle connections: %d\n", stats.IdleConnections)
fmt.Printf("In use connections: %d\n", stats.InUseConnections)

// Configure connection pool
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(1 * time.Hour)
```

## Multi-Database Support

### Database Factory

```go
// Create different database types
mysqlDB, _ := factory.CreateDatabase(database.MySQL, mysqlConfig)
postgresDB, _ := factory.CreateDatabase(database.PostgreSQL, postgresConfig)
mongoDB, _ := factory.CreateDatabase(database.MongoDB, mongoConfig)

// All implement the same Database interface
repositories := []database.Database{mysqlDB, postgresDB, mongoDB}
for _, db := range repositories {
    userRepo := db.GetRepository[User]()
    // Use repository...
}
```

## Error Handling

The library provides consistent error handling across all operations:

```go
// Database errors
if err := userRepo.Create(ctx, user); err != nil {
    if errors.Is(err, gorm.ErrDuplicatedKey) {
        // Handle duplicate key error
    }
    // Handle other errors
}

// Cache errors
if err := cacheManager.Get(ctx, "key", &value); err != nil {
    if errors.Is(err, cache.ErrCacheMiss) {
        // Handle cache miss
    }
    // Handle other errors
}
```

## Testing

### Mock Support

All interfaces can be easily mocked for testing:

```go
type MockRepository struct{}

func (m *MockRepository) Create(ctx context.Context, entity *User) error {
    // Mock implementation
    return nil
}

// Use in tests
mockRepo := &MockRepository{}
// Test with mock...
```

## Performance Considerations

1. **Connection Pooling**: Configure appropriate pool sizes based on your load
2. **Caching**: Use Redis for frequently accessed data
3. **Query Optimization**: Use indexes and optimize queries
4. **Batch Operations**: Use batch operations for bulk inserts/updates
5. **Transaction Scope**: Keep transactions as short as possible

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

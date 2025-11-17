# Database Package

This package provides a comprehensive database abstraction layer with support for multiple database backends, repository patterns, and transaction management.

## Package Structure

```
database/
├── common/                    # Common database utilities
│   ├── base_database.go      # Base database interface
│   ├── base_repository.go     # Base repository implementation
│   └── types.go              # Common types and interfaces
├── gorm/                      # GORM implementation
│   ├── gorm.go               # GORM database implementation
│   ├── migration.go          # GORM migration utilities
│   └── repository.go         # GORM repository implementation
├── mongodb/                   # MongoDB implementation
│   ├── mongodb.go            # MongoDB database implementation
│   ├── health.go             # MongoDB health checks
│   ├── migration.go          # MongoDB migration utilities
│   ├── query_builder.go      # MongoDB query builder
│   ├── repository.go         # MongoDB repository implementation
│   └── transaction.go        # MongoDB transaction support
├── postgresql/                # PostgreSQL implementation
│   ├── postgresql.go         # PostgreSQL database implementation
│   ├── adapter.go            # PostgreSQL adapter
│   ├── health.go             # PostgreSQL health checks
│   ├── migration.go          # PostgreSQL migration utilities
│   ├── query_builder.go      # PostgreSQL query builder
│   ├── repository.go         # PostgreSQL repository implementation
│   └── transaction.go        # PostgreSQL transaction support
├── interfaces/                # Database interfaces
│   ├── database.go           # Core database interface
│   ├── health.go             # Health check interface
│   ├── migration.go          # Migration interface
│   ├── monitoring.go         # Monitoring interface
│   ├── query_builder.go      # Query builder interface
│   ├── repository.go         # Repository interface
│   └── transaction.go        # Transaction interface
├── repository/                # Repository implementations
│   └── base_repository.go    # Base repository
├── database_factory.go        # Database factory
└── types.go                  # Database types
```

## Core Components

### 1. Database Interface (`interfaces/database.go`)

The core database interface defines the fundamental database operations:

```go
type Database interface {
    // Connection management
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Ping(ctx context.Context) error

    // Transaction support
    BeginTransaction(ctx context.Context) (Transaction, error)
    WithTransaction(ctx context.Context, fn func(Transaction) error) error

    // Repository access
    GetRepository() Repository
    GetQueryBuilder() QueryBuilder

    // Health and monitoring
    GetHealthChecker() HealthChecker
    GetMonitor() Monitor
}
```

### 2. Repository Interface (`interfaces/repository.go`)

The repository interface provides data access operations:

```go
type Repository interface {
    // CRUD operations
    Create(ctx context.Context, entity interface{}) error
    Read(ctx context.Context, id interface{}, dest interface{}) error
    Update(ctx context.Context, entity interface{}) error
    Delete(ctx context.Context, id interface{}) error

    // Query operations
    Find(ctx context.Context, query interface{}, dest interface{}) error
    FindAll(ctx context.Context, dest interface{}) error
    Count(ctx context.Context, query interface{}) (int64, error)

    // Batch operations
    CreateBatch(ctx context.Context, entities []interface{}) error
    UpdateBatch(ctx context.Context, entities []interface{}) error
    DeleteBatch(ctx context.Context, ids []interface{}) error
}
```

### 3. Transaction Interface (`interfaces/transaction.go`)

The transaction interface provides transaction management:

```go
type Transaction interface {
    // Transaction control
    Commit() error
    Rollback() error

    // Repository access within transaction
    GetRepository() Repository
    GetQueryBuilder() QueryBuilder
}
```

### 4. Query Builder Interface (`interfaces/query_builder.go`)

The query builder interface provides query construction:

```go
type QueryBuilder interface {
    // Query construction
    Select(fields ...string) QueryBuilder
    From(table string) QueryBuilder
    Where(condition string, args ...interface{}) QueryBuilder
    OrderBy(field string, direction string) QueryBuilder
    Limit(limit int) QueryBuilder
    Offset(offset int) QueryBuilder

    // Query execution
    Execute(ctx context.Context, dest interface{}) error
    Count(ctx context.Context) (int64, error)
}
```

## Database Implementations

### PostgreSQL Implementation

```go
// Create PostgreSQL database
db, err := postgresql.NewPostgreSQLDatabase(&config.PostgreSQLConfig{
    Host:     "localhost",
    Port:     5432,
    User:     "postgres",
    Password: "password",
    Database: "mydb",
    SSLMode:  "disable",
})

// Connect to database
err = db.Connect(ctx)

// Get repository
repo := db.GetRepository()

// Create entity
user := &User{Name: "John", Email: "john@example.com"}
err = repo.Create(ctx, user)

// Read entity
var foundUser User
err = repo.Read(ctx, user.ID, &foundUser)

// Update entity
foundUser.Name = "Jane"
err = repo.Update(ctx, &foundUser)

// Delete entity
err = repo.Delete(ctx, user.ID)
```

### MongoDB Implementation

```go
// Create MongoDB database
db, err := mongodb.NewMongoDatabase(&config.MongoConfig{
    URI:      "mongodb://localhost:27017",
    Database: "mydb",
})

// Connect to database
err = db.Connect(ctx)

// Get repository
repo := db.GetRepository()

// Create document
user := &User{Name: "John", Email: "john@example.com"}
err = repo.Create(ctx, user)

// Find documents
var users []User
err = repo.Find(ctx, bson.M{"name": "John"}, &users)
```

### GORM Implementation

```go
// Create GORM database
db, err := gorm.NewGormDatabase(&config.GormConfig{
    DSN: "postgres://user:password@localhost/mydb?sslmode=disable",
})

// Connect to database
err = db.Connect(ctx)

// Get repository
repo := db.GetRepository()

// Create entity
user := &User{Name: "John", Email: "john@example.com"}
err = repo.Create(ctx, user)

// Find with conditions
var users []User
err = repo.Find(ctx, &User{Name: "John"}, &users)
```

## Transaction Management

### Basic Transaction Usage

```go
// Using WithTransaction (recommended)
err := db.WithTransaction(ctx, func(tx Transaction) error {
    repo := tx.GetRepository()

    // Create user
    user := &User{Name: "John", Email: "john@example.com"}
    err := repo.Create(ctx, user)
    if err != nil {
        return err // Transaction will be rolled back
    }

    // Create profile
    profile := &Profile{UserID: user.ID, Bio: "Software Developer"}
    err = repo.Create(ctx, profile)
    if err != nil {
        return err // Transaction will be rolled back
    }

    return nil // Transaction will be committed
})
```

### Manual Transaction Management

```go
// Begin transaction
tx, err := db.BeginTransaction(ctx)
if err != nil {
    return err
}

// Ensure cleanup
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()

// Perform operations
repo := tx.GetRepository()
user := &User{Name: "John", Email: "john@example.com"}
err = repo.Create(ctx, user)
if err != nil {
    return err
}

// Commit transaction
err = tx.Commit()
if err != nil {
    return err
}
```

## Query Building

### PostgreSQL Query Builder

```go
// Get query builder
qb := db.GetQueryBuilder()

// Build and execute query
var users []User
err := qb.
    Select("id", "name", "email").
    From("users").
    Where("age > ?", 18).
    OrderBy("name", "ASC").
    Limit(10).
    Execute(ctx, &users)
```

### MongoDB Query Builder

```go
// Get query builder
qb := db.GetQueryBuilder()

// Build and execute query
var users []User
err := qb.
    Select("name", "email").
    From("users").
    Where(bson.M{"age": bson.M{"$gt": 18}}).
    OrderBy("name", 1).
    Limit(10).
    Execute(ctx, &users)
```

## Health Checks

### Database Health Monitoring

```go
// Get health checker
healthChecker := db.GetHealthChecker()

// Check database health
health, err := healthChecker.CheckHealth(ctx)
if err != nil {
    log.Error("Database health check failed", err)
    return
}

if health.Status == "healthy" {
    log.Info("Database is healthy")
} else {
    log.Warn("Database health issues", "status", health.Status)
}
```

### Custom Health Checks

```go
// Implement custom health checker
type CustomHealthChecker struct {
    db Database
}

func (c *CustomHealthChecker) CheckHealth(ctx context.Context) (*HealthStatus, error) {
    // Custom health check logic
    err := c.db.Ping(ctx)
    if err != nil {
        return &HealthStatus{
            Status: "unhealthy",
            Details: map[string]interface{}{
                "error": err.Error(),
            },
        }, nil
    }

    return &HealthStatus{
        Status: "healthy",
        Details: map[string]interface{}{
            "timestamp": time.Now(),
        },
    }, nil
}
```

## Migration Management

### Running Migrations

```go
// Get migration manager
migrationManager := db.GetMigrationManager()

// Run all pending migrations
err := migrationManager.Migrate(ctx)

// Run specific migration
err := migrationManager.MigrateTo(ctx, "20231201_001")

// Rollback last migration
err := migrationManager.Rollback(ctx)

// Get migration status
status, err := migrationManager.GetStatus(ctx)
```

### Creating Migrations

```go
// PostgreSQL migration
func CreateUsersTable(tx Transaction) error {
    _, err := tx.Exec(`
        CREATE TABLE users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
    return err
}

// MongoDB migration
func CreateUsersCollection(tx Transaction) error {
    return tx.CreateCollection("users", bson.M{
        "validator": bson.M{
            "$jsonSchema": bson.M{
                "bsonType": "object",
                "required": []string{"name", "email"},
                "properties": bson.M{
                    "name": bson.M{"bsonType": "string"},
                    "email": bson.M{"bsonType": "string"},
                },
            },
        },
    })
}
```

## Monitoring and Observability

### Database Metrics

```go
// Get monitor
monitor := db.GetMonitor()

// Get database metrics
metrics, err := monitor.GetMetrics(ctx)
if err != nil {
    return err
}

log.Info("Database metrics",
    "connections", metrics.Connections,
    "queries_per_second", metrics.QueriesPerSecond,
    "average_query_time", metrics.AverageQueryTime,
)
```

### Custom Monitoring

```go
// Implement custom monitor
type CustomMonitor struct {
    db Database
}

func (m *CustomMonitor) GetMetrics(ctx context.Context) (*DatabaseMetrics, error) {
    // Custom monitoring logic
    return &DatabaseMetrics{
        Connections:        getConnectionCount(),
        QueriesPerSecond:   getQPS(),
        AverageQueryTime:   getAverageQueryTime(),
        MemoryUsage:        getMemoryUsage(),
    }, nil
}
```

## Configuration

### PostgreSQL Configuration

```go
config := &config.PostgreSQLConfig{
    Host:            "localhost",
    Port:            5432,
    User:            "postgres",
    Password:        "password",
    Database:        "mydb",
    SSLMode:         "disable",
    MaxConnections:  25,
    MinConnections:  5,
    ConnectionTimeout: 30 * time.Second,
    QueryTimeout:    10 * time.Second,
}
```

### MongoDB Configuration

```go
config := &config.MongoConfig{
    URI:              "mongodb://localhost:27017",
    Database:         "mydb",
    MaxPoolSize:      100,
    MinPoolSize:      10,
    ConnectTimeout:   30 * time.Second,
    SocketTimeout:    10 * time.Second,
    ServerSelectionTimeout: 5 * time.Second,
}
```

### GORM Configuration

```go
config := &config.GormConfig{
    DSN:             "postgres://user:password@localhost/mydb?sslmode=disable",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: time.Minute * 30,
}
```

## Best Practices

### 1. Connection Management

- Use connection pooling
- Set appropriate timeouts
- Monitor connection usage
- Handle connection failures gracefully

### 2. Transaction Management

- Keep transactions short
- Use `WithTransaction` for automatic cleanup
- Handle rollback scenarios
- Avoid long-running transactions

### 3. Query Optimization

- Use appropriate indexes
- Limit result sets
- Use query builders for complex queries
- Monitor query performance

### 4. Error Handling

- Handle database errors gracefully
- Log errors with context
- Implement retry logic for transient errors
- Provide meaningful error messages

### 5. Testing

- Use database transactions for test isolation
- Mock database for unit tests
- Use test databases for integration tests
- Clean up test data

## Examples

See the `examples/` directory for comprehensive usage examples:

- `examples/adapters/` - Database adapter examples
- `examples/main/` - Main application examples

## Migration Guide

When upgrading between versions:

1. **Check breaking changes** in the changelog
2. **Update configuration** if needed
3. **Run migrations** for schema changes
4. **Update tests** for API changes
5. **Monitor performance** after upgrade

## Future Enhancements

- **Multi-database support** - Support for multiple database connections
- **Read replicas** - Support for read/write splitting
- **Connection pooling** - Advanced connection pool management
- **Query caching** - Built-in query result caching
- **Database sharding** - Support for database sharding

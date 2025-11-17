# PostgreSQL Database Adapter

This package provides a shared PostgreSQL database adapter that can be used across all microservices.

## Features

- **Generic Database Interface**: Implements `interfaces.Database` for consistent database operations
- **Health Monitoring**: Built-in health checks and connection monitoring
- **Transaction Management**: Support for database transactions
- **Query Building**: Simple query builder for common operations
- **Migration Support**: Basic migration management
- **Connection Pooling**: Leverages Go's built-in connection pooling

## Usage

### In Microservices

Each microservice can create its own thin wrapper that delegates to this shared implementation:

```go
// In user-service/internal/infrastructure/persistence/postgresql/database_adapter.go
package postgresql

import (
    "database/sql"
    "backend-core/database/interfaces"
    "backend-core/database/postgresql"
)

// NewDatabaseAdapter creates a new database adapter using the shared backend-core implementation
func NewDatabaseAdapter(db *sql.DB) interfaces.Database {
    return postgresql.NewDatabaseAdapter(db)
}
```

### Direct Usage

```go
import (
    "database/sql"
    "backend-core/database/postgresql"
)

// Create database connection
db, err := sql.Open("postgres", dsn)
if err != nil {
    log.Fatal(err)
}

// Create adapter
adapter := postgresql.NewDatabaseAdapter(db)

// Use the adapter
err = adapter.Connect(ctx)
if err != nil {
    log.Fatal(err)
}

// Check health
status := adapter.GetHealthStatus()
stats := adapter.GetStats()
```

## Architecture

```
backend-core/database/postgresql/
├── adapter.go              # ✅ Shared PostgreSQL adapter
└── README.md               # ✅ Documentation

auth-service/internal/infrastructure/persistence/postgresql/
├── database_adapter.go     # ✅ Thin wrapper delegating to backend-core
└── user_repository.go      # ✅ Service-specific repository

user-service/internal/infrastructure/persistence/postgresql/
├── database_adapter.go     # ✅ Thin wrapper delegating to backend-core
└── user_repository.go      # ✅ Service-specific repository
```

## Benefits

1. **DRY Principle**: No code duplication across microservices
2. **Consistency**: Same database interface across all services
3. **Maintainability**: Single place to update database logic
4. **Testing**: Shared test utilities and mocks
5. **Performance**: Optimized connection pooling and monitoring

## Interfaces Implemented

- `interfaces.Database` - Core database operations
- `interfaces.DatabaseMonitor` - Database monitoring
- `interfaces.DatabaseHealthChecker` - Health checking
- `interfaces.QueryBuilder` - Query building
- `interfaces.MigrationManager` - Migration management
- `interfaces.TransactionManager` - Transaction management

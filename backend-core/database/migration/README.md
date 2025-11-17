# 🗄️ Shared Database Migration System

## 📋 **Overview**

The `backend-core/database/migration` package provides a **reusable migration system** that can be used by **all services** in the microservices architecture. This eliminates code duplication and ensures consistent migration management across services.

---

## 🏗️ **Architecture Benefits**

### **✅ Reusable Across Services**
- **Single Implementation**: One migration system for all services
- **Consistent Interface**: Same API across all services
- **Shared Logic**: Common migration logic in one place
- **Easy Maintenance**: Update once, benefit everywhere

### **✅ Service-Specific Migrations**
- **Isolated Migrations**: Each service has its own migration files
- **Service-Specific Schema**: Each service manages its own database schema
- **Independent Versioning**: Services can have different migration versions
- **Flexible Deployment**: Services can be deployed independently

---

## 📁 **Package Structure**

```
backend-core/database/migration/
├── migration_manager.go      # ✅ Core migration management
├── migration_loader.go        # ✅ Migration file loading
├── migration_service.go      # ✅ High-level migration operations
└── README.md                 # ✅ Documentation

backend-core/cmd/migrate/
└── main.go                   # ✅ Shared migration CLI tool
```

---

## 🚀 **Usage in Services**

### **1. Service-Specific Migration Directory**
Each service should have its own migrations directory:

```
auth-service/
├── migrations/               # ✅ Service-specific migrations
│   ├── 001_create_users.sql
│   ├── 002_create_roles.sql
│   └── 003_create_permissions.sql
└── ...

notification-service/
├── migrations/               # ✅ Service-specific migrations
│   ├── 001_create_notifications.sql
│   ├── 002_create_templates.sql
│   └── 003_create_subscriptions.sql
└── ...
```

### **2. Service Integration**
Each service can integrate the migration system:

```go
package main

import (
    "context"
    "database/sql"
    
    "backend-core/database/migration"
    "backend-core/logging"
)

func main() {
    // Initialize logger
    logger, _ := logging.NewLogger(&cfg.Logging)
    
    // Initialize database connection
    sqlDB, _ := sql.Open("postgres", dsn)
    
    // Initialize migration components
    migrationManager := migration.NewMigrationManager(sqlDB, logger)
    migrationLoader := migration.NewMigrationLoader("./migrations", logger)
    migrationService := migration.NewMigrationService(migrationManager, migrationLoader, logger)
    
    // Run migrations on startup
    if err := migrationService.RunMigrations(context.Background()); err != nil {
        logger.Fatal("Failed to run migrations", logging.Error(err))
    }
}
```

---

## 🛠️ **Migration CLI Tool**

### **Shared CLI Tool**
The `backend-core/cmd/migrate` provides a shared CLI tool that can be used by any service:

```bash
# Build the shared migration tool
go build -o bin/migrate ./backend-core/cmd/migrate

# Use in any service
./bin/migrate -command=up -config=./config.yaml -migrations=./migrations
./bin/migrate -command=down -target=2 -config=./config.yaml -migrations=./migrations
./bin/migrate -command=status -config=./config.yaml -migrations=./migrations
./bin/migrate -command=create -name=add_indexes -config=./config.yaml -migrations=./migrations
```

### **Service-Specific Makefiles**
Each service can have its own Makefile that uses the shared tool:

```makefile
# auth-service/Makefile
MIGRATE_TOOL = ../backend-core/bin/migrate
CONFIG_FILE = ./config/config.yaml
MIGRATIONS_DIR = ./migrations

.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	@$(MIGRATE_TOOL) -command=up -config=$(CONFIG_FILE) -migrations=$(MIGRATIONS_DIR)

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	@$(MIGRATE_TOOL) -command=down -target=$(TARGET) -config=$(CONFIG_FILE) -migrations=$(MIGRATIONS_DIR)

.PHONY: migrate-status
migrate-status:
	@echo "Checking migration status..."
	@$(MIGRATE_TOOL) -command=status -config=$(CONFIG_FILE) -migrations=$(MIGRATIONS_DIR)

.PHONY: migrate-create
migrate-create:
	@echo "Creating new migration..."
	@$(MIGRATE_TOOL) -command=create -name=$(NAME) -config=$(CONFIG_FILE) -migrations=$(MIGRATIONS_DIR)
```

---

## 📊 **Migration File Format**

### **Standard Template**
All services use the same migration file format:

```sql
-- Migration: migration_name
-- Description: Brief description of what this migration does

-- +++++ UP
-- Your up migration SQL here
CREATE TABLE example_table (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +++++ DOWN
-- Your down migration SQL here
DROP TABLE IF EXISTS example_table;
```

### **Service-Specific Examples**

#### **Auth Service Migrations**
```sql
-- Migration: create_users_table
-- Description: Create users table for authentication

-- +++++ UP
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(254) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- +++++ DOWN
DROP TABLE IF EXISTS users;
```

#### **Notification Service Migrations**
```sql
-- Migration: create_notifications_table
-- Description: Create notifications table for messaging

-- +++++ UP
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    message TEXT NOT NULL,
    sent_at TIMESTAMP DEFAULT NOW()
);

-- +++++ DOWN
DROP TABLE IF EXISTS notifications;
```

---

## 🔧 **Configuration**

### **Database Configuration**
Each service needs its own database configuration:

```yaml
# auth-service/config.yaml
database:
  host: localhost
  port: 5432
  username: auth_user
  password: auth_password
  database: auth_service
  sslmode: disable

# notification-service/config.yaml
database:
  host: localhost
  port: 5432
  username: notification_user
  password: notification_password
  database: notification_service
  sslmode: disable
```

### **Migration Configuration**
Migration-specific configuration:

```yaml
# Each service can have migration-specific settings
migration:
  table_name: schema_migrations
  checksum_validation: true
  rollback_safety: true
```

---

## 🎯 **Benefits for Multi-Service Architecture**

### **1. Consistency**
- **Same Interface**: All services use the same migration API
- **Same Format**: All services use the same migration file format
- **Same Commands**: All services use the same CLI commands
- **Same Safety**: All services have the same safety features

### **2. Maintainability**
- **Single Source**: Migration logic in one place
- **Easy Updates**: Update once, benefit all services
- **Bug Fixes**: Fix bugs once, benefit all services
- **Feature Additions**: Add features once, benefit all services

### **3. Development Efficiency**
- **Shared Knowledge**: Developers learn once, use everywhere
- **Shared Tools**: Same CLI tools for all services
- **Shared Documentation**: One set of documentation
- **Shared Testing**: Test once, use everywhere

### **4. Production Safety**
- **Consistent Safety**: Same safety features across all services
- **Consistent Rollback**: Same rollback capabilities
- **Consistent Monitoring**: Same monitoring and logging
- **Consistent Auditing**: Same audit trail

---

## 🚀 **Implementation Steps**

### **1. Move Migration System to backend-core**
```bash
# Move from auth-service to backend-core
mv auth-service/internal/infrastructure/migration/* backend-core/database/migration/
mv auth-service/cmd/migrate/* backend-core/cmd/migrate/
```

### **2. Update Service Dependencies**
```go
// In each service, import from backend-core
import "backend-core/database/migration"
```

### **3. Update Service Makefiles**
```makefile
# Use shared migration tool
MIGRATE_TOOL = ../backend-core/bin/migrate
```

### **4. Update Service Documentation**
```markdown
# Update each service's README to reference shared migration system
```

---

## 📈 **Future Enhancements**

### **Planned Features**
1. **Service Discovery**: Automatic service discovery for migrations
2. **Cross-Service Dependencies**: Handle migrations that depend on other services
3. **Migration Orchestration**: Coordinate migrations across multiple services
4. **Service-Specific Templates**: Pre-built templates for common service patterns
5. **Migration Analytics**: Track migration performance across services

---

## 🎉 **Summary**

### **✅ Benefits Achieved**
- **Reusable Migration System**: One system for all services
- **Consistent Interface**: Same API across all services
- **Easy Maintenance**: Update once, benefit everywhere
- **Service Isolation**: Each service manages its own migrations
- **Shared Tooling**: Common CLI tools and documentation

### **✅ Architecture Improvements**
- **Eliminated Duplication**: No more duplicate migration systems
- **Improved Consistency**: Same migration approach everywhere
- **Enhanced Maintainability**: Single source of truth
- **Better Developer Experience**: Learn once, use everywhere

---

**🎉 The shared migration system provides a robust, reusable, and consistent database schema management solution for all services in the microservices architecture!**

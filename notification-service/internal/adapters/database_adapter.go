package adapters

import (
	"backend-core/adapters"
	"backend-core/database/gorm"
)

// DatabaseAdapter wraps the shared database adapter for notification-service
type DatabaseAdapter = adapters.DatabaseAdapter

// NewDatabaseAdapter creates a new database adapter using the shared implementation
func NewDatabaseAdapter(db gorm.Database) *DatabaseAdapter {
	return adapters.NewDatabaseAdapter(db)
}

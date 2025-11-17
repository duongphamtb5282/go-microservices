package postgresql

import (
	"backend-core/database/gorm"
	"backend-core/logging"

	gormLib "gorm.io/gorm"
)

// PostgreSQLRepository implements the Repository interface for PostgreSQL using GORM
type PostgreSQLRepository[T any] struct {
	*gorm.GormRepository[T]
}

// NewPostgreSQLRepository creates a new PostgreSQL repository using GORM
func NewPostgreSQLRepository[T any](database gorm.Database, entityType string, logger *logging.Logger) *PostgreSQLRepository[T] {
	baseRepo := gorm.NewGormRepository[T](database, entityType, logger)

	return &PostgreSQLRepository[T]{
		GormRepository: baseRepo,
	}
}

// GetGormDB returns the underlying GORM database instance
func (r *PostgreSQLRepository[T]) GetGormDB() *gormLib.DB {
	// Access through the embedded GormRepository
	return r.GormRepository.GetGormDB()
}

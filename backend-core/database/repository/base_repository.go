package repository

import (
	"context"
	"fmt"

	"backend-core/database/interfaces"
	"backend-shared/models"
)

// BaseRepository provides common repository functionality
type BaseRepository[T any] struct {
	db     interfaces.Database
	logger interface{}
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](db interfaces.Database, logger interface{}) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:     db,
		logger: logger,
	}
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	// This would be implemented by the specific database implementation
	// For now, return a placeholder
	return fmt.Errorf("Create method must be implemented by specific repository")
}

// GetByID retrieves an entity by ID
func (r *BaseRepository[T]) GetByID(ctx context.Context, id models.EntityID) (*T, error) {
	// This would be implemented by the specific database implementation
	// For now, return a placeholder
	return nil, fmt.Errorf("GetByID method must be implemented by specific repository")
}

// Update updates an entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	// This would be implemented by the specific database implementation
	// For now, return a placeholder
	return fmt.Errorf("Update method must be implemented by specific repository")
}

// Delete deletes an entity by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id models.EntityID) error {
	// This would be implemented by the specific database implementation
	// For now, return a placeholder
	return fmt.Errorf("Delete method must be implemented by specific repository")
}

// Exists checks if an entity exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id models.EntityID) (bool, error) {
	// This would be implemented by the specific database implementation
	// For now, return a placeholder
	return false, fmt.Errorf("Exists method must be implemented by specific repository")
}

// GetDatabase returns the underlying database connection
func (r *BaseRepository[T]) GetDatabase() interfaces.Database {
	return r.db
}

// GetLogger returns the logger
func (r *BaseRepository[T]) GetLogger() interface{} {
	return r.logger
}

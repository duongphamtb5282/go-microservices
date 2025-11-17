package gorm

import (
	"context"
	"fmt"

	"backend-core/logging"

	"gorm.io/gorm"
)

// Repository defines the repository interface for data access
type Repository[T any] interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id interface{}) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id interface{}) error
	GetAll(ctx context.Context, filter map[string]interface{}, pagination Pagination) ([]*T, error)

	// Query operations
	Find(ctx context.Context, query Query) ([]*T, error)
	Count(ctx context.Context, filter map[string]interface{}) (int64, error)
	Exists(ctx context.Context, filter map[string]interface{}) (bool, error)

	// Batch operations
	CreateBatch(ctx context.Context, entities []*T) error
	UpdateBatch(ctx context.Context, entities []*T) error
	DeleteBatch(ctx context.Context, ids []interface{}) error

	// Transaction support
	WithTransaction(ctx context.Context, fn func(Repository[T]) error) error
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Query represents a database query
type Query struct {
	SQL        string                 `json:"sql"`
	Args       []interface{}          `json:"args"`
	Filter     map[string]interface{} `json:"filter"`
	Pagination Pagination             `json:"pagination"`
	OrderBy    string                 `json:"order_by"`
	Order      string                 `json:"order"`
}

// GormRepository provides common repository functionality using GORM
type GormRepository[T any] struct {
	database   Database
	entityType string
	logger     *logging.Logger
	tableName  string
}

// NewGormRepository creates a new GORM repository
func NewGormRepository[T any](database Database, entityType string, logger *logging.Logger) *GormRepository[T] {
	return &GormRepository[T]{
		database:   database,
		entityType: entityType,
		logger:     logger,
		tableName:  getTableName[T](),
	}
}

// GetGormDB returns the underlying GORM database instance
func (r *GormRepository[T]) GetGormDB() *gorm.DB {
	return r.database.GetGormDB()
}

// Create inserts a new entity
func (r *GormRepository[T]) Create(ctx context.Context, entity *T) error {
	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Create(entity).Error
	if err != nil {
		r.logger.Error("failed to create entity", "error", err)
		return err
	}

	// Entity created successfully
	return nil
}

// GetByID retrieves an entity by ID
func (r *GormRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
	var entity T
	db := r.database.GetGormDB().WithContext(ctx)

	// Use Where clause to properly handle string IDs (like UUIDs)
	err := db.Where("id = ?", id).First(&entity).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error("failed to get entity by ID", "error", err)
		return nil, err
	}

	return &entity, nil
}

// Update updates an entity
func (r *GormRepository[T]) Update(ctx context.Context, entity *T) error {
	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Save(entity).Error
	if err != nil {
		r.logger.Error("failed to update entity", "error", err)
		return err
	}

	// Entity updated successfully
	return nil
}

// Delete removes an entity by ID
func (r *GormRepository[T]) Delete(ctx context.Context, id interface{}) error {
	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Delete(new(T), id).Error
	if err != nil {
		r.logger.Error("failed to delete entity", "error", err)
		return err
	}

	// Entity deleted successfully
	return nil
}

// GetAll retrieves multiple entities with optional filtering and pagination
func (r *GormRepository[T]) GetAll(ctx context.Context, filter map[string]interface{}, pagination Pagination) ([]*T, error) {
	var entities []*T
	db := r.database.GetGormDB().WithContext(ctx)

	// Apply filters
	for field, value := range filter {
		db = db.Where(field+" = ?", value)
	}

	// Apply pagination
	if pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		db = db.Offset(offset).Limit(pagination.PageSize)
	}

	err := db.Find(&entities).Error
	if err != nil {
		r.logger.Error("failed to get all entities", "error", err)
		return nil, err
	}

	return entities, nil
}

// Find executes a complex query
func (r *GormRepository[T]) Find(ctx context.Context, query Query) ([]*T, error) {
	var entities []*T
	db := r.database.GetGormDB().WithContext(ctx)

	// Apply filters
	for field, value := range query.Filter {
		db = db.Where(field+" = ?", value)
	}

	// Apply ordering
	if query.OrderBy != "" {
		order := query.OrderBy
		if query.Order == "desc" {
			order += " DESC"
		} else {
			order += " ASC"
		}
		db = db.Order(order)
	}

	// Apply pagination
	if query.Pagination.PageSize > 0 {
		offset := (query.Pagination.Page - 1) * query.Pagination.PageSize
		db = db.Offset(offset).Limit(query.Pagination.PageSize)
	}

	err := db.Find(&entities).Error
	if err != nil {
		r.logger.Error("failed to find entities", "error", err)
		return nil, err
	}

	return entities, nil
}

// Count returns the number of entities matching the filter
func (r *GormRepository[T]) Count(ctx context.Context, filter map[string]interface{}) (int64, error) {
	var count int64
	db := r.database.GetGormDB().WithContext(ctx)

	// Apply filters
	for field, value := range filter {
		db = db.Where(field+" = ?", value)
	}

	err := db.Model(new(T)).Count(&count).Error
	if err != nil {
		r.logger.Error("failed to count entities", "error", err)
		return 0, err
	}

	return count, nil
}

// Exists checks if any entity matches the filter
func (r *GormRepository[T]) Exists(ctx context.Context, filter map[string]interface{}) (bool, error) {
	count, err := r.Count(ctx, filter)
	return count > 0, err
}

// CreateBatch inserts multiple entities
func (r *GormRepository[T]) CreateBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	db := r.database.GetGormDB().WithContext(ctx)
	err := db.CreateInBatches(entities, 100).Error
	if err != nil {
		r.logger.Error("failed to create batch entities", "error", err)
		return err
	}

	// Batch entities created successfully
	return nil
}

// UpdateBatch updates multiple entities
func (r *GormRepository[T]) UpdateBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}

	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Save(entities).Error
	if err != nil {
		r.logger.Error("failed to update batch entities", "error", err)
		return err
	}

	// Batch entities updated successfully
	return nil
}

// DeleteBatch removes multiple entities by IDs
func (r *GormRepository[T]) DeleteBatch(ctx context.Context, ids []interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Delete(new(T), ids).Error
	if err != nil {
		r.logger.Error("failed to delete batch entities", "error", err)
		return err
	}

	// Batch entities deleted successfully
	return nil
}

// WithTransaction executes a function within a transaction
func (r *GormRepository[T]) WithTransaction(ctx context.Context, fn func(Repository[T]) error) error {
	return r.database.WithTransaction(ctx, func(transaction Transaction) error {
		// Create a new repository instance with the transaction
		txRepo := &GormRepository[T]{
			database:   r.database,
			entityType: r.entityType,
			logger:     r.logger,
			tableName:  r.tableName,
		}
		return fn(txRepo)
	})
}

// FindWithRawSQL executes raw SQL query
func (r *GormRepository[T]) FindWithRawSQL(ctx context.Context, query string, args ...interface{}) ([]*T, error) {
	var entities []*T
	db := r.database.GetGormDB().WithContext(ctx)

	err := db.Raw(query, args...).Scan(&entities).Error
	if err != nil {
		r.logger.Error("failed to execute raw SQL", "error", err)
		return nil, err
	}

	return entities, nil
}

// ExecuteRawSQL executes raw SQL without returning results
func (r *GormRepository[T]) ExecuteRawSQL(ctx context.Context, query string, args ...interface{}) error {
	db := r.database.GetGormDB().WithContext(ctx)
	err := db.Exec(query, args...).Error
	if err != nil {
		r.logger.Error("failed to execute raw SQL", "error", err)
		return err
	}

	return nil
}

// BuildComplexQuery returns a GORM query builder for complex queries
func (r *GormRepository[T]) BuildComplexQuery(ctx context.Context) *gorm.DB {
	return r.database.GetGormDB().WithContext(ctx).Model(new(T))
}

// getTableName returns the table name for the entity type
func getTableName[T any]() string {
	var entity T
	// This is a simplified approach - in a real implementation,
	// you might want to use reflection or a more sophisticated method
	return fmt.Sprintf("%T", entity)
}

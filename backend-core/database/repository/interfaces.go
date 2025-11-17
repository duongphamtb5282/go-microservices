package repository

import (
	"context"
	"time"
)

// Repository defines the interface for data access operations
type Repository[T any] interface {
	// Create operations
	Create(ctx context.Context, entity *T) error
	CreateBatch(ctx context.Context, entities []*T) error

	// Read operations
	GetByID(ctx context.Context, id interface{}) (*T, error)
	GetByField(ctx context.Context, field string, value interface{}) (*T, error)
	GetAll(ctx context.Context, filter Filter, pagination Pagination) ([]*T, error)

	// Update operations
	Update(ctx context.Context, entity *T) error
	UpdateField(ctx context.Context, id interface{}, field string, value interface{}) error
	Upsert(ctx context.Context, filter Filter, entity *T) error

	// Delete operations
	Delete(ctx context.Context, id interface{}) error
	DeleteBatch(ctx context.Context, ids []interface{}) error

	// Query operations
	Count(ctx context.Context, filter Filter) (int64, error)
	Exists(ctx context.Context, filter Filter) (bool, error)
	Find(ctx context.Context, q Query) ([]*T, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(Repository[T]) error) error
}

// Filter represents database filter conditions
type Filter map[string]interface{}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Query represents a complex database query
type Query struct {
	Filter     Filter     `json:"filter"`
	Pagination Pagination `json:"pagination"`
	OrderBy    string     `json:"order_by"`
	Order      string     `json:"order"` // asc, desc
}

// CacheableRepository extends Repository with caching capabilities
type CacheableRepository[T any] interface {
	Repository[T]

	// Cache operations
	GetCached(ctx context.Context, key string) (*T, error)
	SetCache(ctx context.Context, key string, entity *T, ttl time.Duration) error
	InvalidateCache(ctx context.Context, pattern string) error
}

// SearchableRepository extends Repository with search capabilities
type SearchableRepository[T any] interface {
	Repository[T]

	// Search operations
	Search(ctx context.Context, query string, fields []string, limit, offset int) ([]*T, error)
	SearchByText(ctx context.Context, text string, limit, offset int) ([]*T, error)
	CreateSearchIndex(ctx context.Context, fields []string) error
}

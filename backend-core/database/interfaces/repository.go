package interfaces

import "context"

// Repository defines the repository interface for data access
type Repository[T any] interface {
	// Basic CRUD operations
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id interface{}) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id interface{}) error
	GetAll(ctx context.Context, filter Filter, pagination Pagination) ([]*T, error)

	// Query operations
	Find(ctx context.Context, query Query) ([]*T, error)
	Count(ctx context.Context, filter Filter) (int64, error)
	Exists(ctx context.Context, filter Filter) (bool, error)

	// Batch operations
	CreateBatch(ctx context.Context, entities []*T) error
	UpdateBatch(ctx context.Context, entities []*T) error
	DeleteBatch(ctx context.Context, ids []interface{}) error
}

// Filter represents a database filter
type Filter map[string]interface{}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Query represents a database query
type Query struct {
	SQL        string        `json:"sql"`
	Args       []interface{} `json:"args"`
	Filter     Filter        `json:"filter"`
	Pagination Pagination    `json:"pagination"`
	OrderBy    string        `json:"order_by"`
	Order      string        `json:"order"`
}

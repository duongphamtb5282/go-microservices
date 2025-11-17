package query

import "context"

// Index defines database index configuration
type Index struct {
	Keys    map[string]interface{} `json:"keys"`
	Options map[string]interface{} `json:"options"`
}

// IndexManager manages database indexes
type IndexManager interface {
	CreateIndex(ctx context.Context, collection string, index Index) error
	CreateIndexes(ctx context.Context, collection string, indexes []Index) error
	DropIndex(ctx context.Context, collection string, indexName string) error
	ListIndexes(ctx context.Context, collection string) ([]Index, error)
}

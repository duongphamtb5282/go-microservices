package cache

import (
	"context"
)

// DataSource defines the interface for data source operations
type DataSource interface {
	// LoadData loads data from the source
	LoadData(ctx context.Context, key string, dest interface{}) error

	// StoreData stores data in the source
	StoreData(ctx context.Context, key string, value interface{}) error

	// DeleteData deletes data from the source
	DeleteData(ctx context.Context, key string) error
}

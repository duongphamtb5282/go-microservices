package transformers

import "context"

// DataTransformer defines an interface for transforming data during reload
type DataTransformer interface {
	// GetName returns the name of the transformer
	GetName() string

	// GetDescription returns the description of the transformer
	GetDescription() string

	// Transform transforms data before storing in cache
	Transform(ctx context.Context, key string, data interface{}) (interface{}, error)

	// ShouldTransform determines if data should be transformed
	ShouldTransform(ctx context.Context, key string, data interface{}) bool

	// GetPriority returns the priority of this transformer (lower = higher priority)
	GetPriority() int
}

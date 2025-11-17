package transformers

import (
	"context"
	"time"
)

// MetadataTransformer adds metadata to data
type MetadataTransformer struct {
	name     string
	priority int
}

// NewMetadataTransformer creates a new metadata transformer
func NewMetadataTransformer(name string, priority int) *MetadataTransformer {
	return &MetadataTransformer{
		name:     name,
		priority: priority,
	}
}

func (t *MetadataTransformer) GetName() string {
	return t.name
}

func (t *MetadataTransformer) GetDescription() string {
	return "Adds metadata to data"
}

func (t *MetadataTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	// Add metadata
	if dataMap, ok := data.(map[string]interface{}); ok {
		dataMap["_metadata"] = map[string]interface{}{
			"transformed_at": time.Now().Unix(),
			"transformer":    t.name,
			"key":            key,
		}
		return dataMap, nil
	}
	return data, nil
}

func (t *MetadataTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	// Transform all data
	return true
}

func (t *MetadataTransformer) GetPriority() int {
	return t.priority
}

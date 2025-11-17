package transformers

import "context"

// CompressionTransformer compresses data
type CompressionTransformer struct {
	name     string
	priority int
	compress func(interface{}) (interface{}, error)
}

// NewCompressionTransformer creates a new compression transformer
func NewCompressionTransformer(name string, priority int, compress func(interface{}) (interface{}, error)) *CompressionTransformer {
	return &CompressionTransformer{
		name:     name,
		priority: priority,
		compress: compress,
	}
}

func (t *CompressionTransformer) GetName() string {
	return t.name
}

func (t *CompressionTransformer) GetDescription() string {
	return "Compresses data to save space"
}

func (t *CompressionTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	if t.compress != nil {
		return t.compress(data)
	}
	return data, nil
}

func (t *CompressionTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	// Only compress large data
	if dataMap, ok := data.(map[string]interface{}); ok {
		return len(dataMap) > 10
	}
	return false
}

func (t *CompressionTransformer) GetPriority() int {
	return t.priority
}

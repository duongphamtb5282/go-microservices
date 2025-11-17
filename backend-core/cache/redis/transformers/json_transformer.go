package transformers

import "context"

// JSONTransformer transforms data to JSON format
type JSONTransformer struct {
	name     string
	priority int
}

// NewJSONTransformer creates a new JSON transformer
func NewJSONTransformer(name string, priority int) *JSONTransformer {
	return &JSONTransformer{
		name:     name,
		priority: priority,
	}
}

func (t *JSONTransformer) GetName() string {
	return t.name
}

func (t *JSONTransformer) GetDescription() string {
	return "Transforms data to JSON format"
}

func (t *JSONTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	// In a real implementation, you would use json.Marshal
	// For now, we'll just return the data as-is
	return data, nil
}

func (t *JSONTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	// Transform all data
	return true
}

func (t *JSONTransformer) GetPriority() int {
	return t.priority
}

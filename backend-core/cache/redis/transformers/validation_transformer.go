package transformers

import "context"

// ValidationTransformer validates and sanitizes data
type ValidationTransformer struct {
	name     string
	priority int
	validate func(interface{}) (interface{}, error)
}

// NewValidationTransformer creates a new validation transformer
func NewValidationTransformer(name string, priority int, validate func(interface{}) (interface{}, error)) *ValidationTransformer {
	return &ValidationTransformer{
		name:     name,
		priority: priority,
		validate: validate,
	}
}

func (t *ValidationTransformer) GetName() string {
	return t.name
}

func (t *ValidationTransformer) GetDescription() string {
	return "Validates and sanitizes data"
}

func (t *ValidationTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	if t.validate != nil {
		return t.validate(data)
	}
	return data, nil
}

func (t *ValidationTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	// Transform all data
	return true
}

func (t *ValidationTransformer) GetPriority() int {
	return t.priority
}

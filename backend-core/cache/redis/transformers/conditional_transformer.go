package transformers

import "context"

// ConditionalTransformer applies transformation based on conditions
type ConditionalTransformer struct {
	name      string
	priority  int
	condition func(context.Context, string, interface{}) bool
	transform func(context.Context, string, interface{}) (interface{}, error)
}

// NewConditionalTransformer creates a new conditional transformer
func NewConditionalTransformer(name string, priority int, condition func(context.Context, string, interface{}) bool, transform func(context.Context, string, interface{}) (interface{}, error)) *ConditionalTransformer {
	return &ConditionalTransformer{
		name:      name,
		priority:  priority,
		condition: condition,
		transform: transform,
	}
}

func (t *ConditionalTransformer) GetName() string {
	return t.name
}

func (t *ConditionalTransformer) GetDescription() string {
	return "Applies transformation based on conditions"
}

func (t *ConditionalTransformer) Transform(ctx context.Context, key string, data interface{}) (interface{}, error) {
	if t.transform != nil {
		return t.transform(ctx, key, data)
	}
	return data, nil
}

func (t *ConditionalTransformer) ShouldTransform(ctx context.Context, key string, data interface{}) bool {
	if t.condition != nil {
		return t.condition(ctx, key, data)
	}
	return false
}

func (t *ConditionalTransformer) GetPriority() int {
	return t.priority
}

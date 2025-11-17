package transformers

import (
	"context"
	"fmt"
)

// DataTransformerManager manages data transformers
type DataTransformerManager struct {
	transformers []DataTransformer
}

// NewDataTransformerManager creates a new data transformer manager
func NewDataTransformerManager() *DataTransformerManager {
	return &DataTransformerManager{
		transformers: make([]DataTransformer, 0),
	}
}

// AddTransformer adds a data transformer
func (m *DataTransformerManager) AddTransformer(transformer DataTransformer) {
	m.transformers = append(m.transformers, transformer)
	// Sort by priority
	m.sortTransformers()
}

// RemoveTransformer removes a transformer by name
func (m *DataTransformerManager) RemoveTransformer(name string) {
	for i, transformer := range m.transformers {
		if transformer.GetName() == name {
			m.transformers = append(m.transformers[:i], m.transformers[i+1:]...)
			break
		}
	}
}

// TransformData transforms data using all applicable transformers
func (m *DataTransformerManager) TransformData(ctx context.Context, key string, data interface{}) (interface{}, error) {
	transformedData := data

	for _, transformer := range m.transformers {
		if transformer.ShouldTransform(ctx, key, transformedData) {
			transformed, err := transformer.Transform(ctx, key, transformedData)
			if err != nil {
				return nil, fmt.Errorf("transformer %s failed: %w", transformer.GetName(), err)
			}
			transformedData = transformed
		}
	}

	return transformedData, nil
}

// GetTransformers returns all registered transformers
func (m *DataTransformerManager) GetTransformers() []DataTransformer {
	return m.transformers
}

// ClearTransformers removes all transformers
func (m *DataTransformerManager) ClearTransformers() {
	m.transformers = make([]DataTransformer, 0)
}

// sortTransformers sorts transformers by priority
func (m *DataTransformerManager) sortTransformers() {
	for i := 0; i < len(m.transformers)-1; i++ {
		for j := i + 1; j < len(m.transformers); j++ {
			if m.transformers[i].GetPriority() > m.transformers[j].GetPriority() {
				m.transformers[i], m.transformers[j] = m.transformers[j], m.transformers[i]
			}
		}
	}
}

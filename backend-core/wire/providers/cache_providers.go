package providers

import (
	"backend-core/cache/decorators"
)

// CacheDecoratorFactoryProvider creates a cache decorator factory
func CacheDecoratorFactoryProvider() *decorators.CacheDecoratorFactory {
	return &decorators.CacheDecoratorFactory{}
}

// CacheDecoratorProvider creates a cache decorator
func CacheDecoratorProvider(factory *decorators.CacheDecoratorFactory) *decorators.CacheDecorator {
	// Simple implementation - would need proper factory method
	return &decorators.CacheDecorator{}
}

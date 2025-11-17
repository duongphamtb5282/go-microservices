package factory

import (
	"backend-core/middleware/core"
)

// MiddlewareBuilder builds middleware chains
type MiddlewareBuilder struct {
	chain   *core.Chain
	factory *MiddlewareFactory
	config  *Config
}

// NewMiddlewareBuilder creates a new middleware builder
func NewMiddlewareBuilder(factory *MiddlewareFactory, config *Config) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		chain:   core.NewChain(),
		factory: factory,
		config:  config,
	}
}

// TODO: Implement builder methods when factory methods are available

// AddCache adds cache middleware
// func (b *MiddlewareBuilder) AddCache() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateCacheMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddAuth adds auth middleware
// func (b *MiddlewareBuilder) AddAuth() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateAuthMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddMetrics adds metrics middleware
// func (b *MiddlewareBuilder) AddMetrics() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateMetricsMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddCors adds CORS middleware
// func (b *MiddlewareBuilder) AddCors() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateCorsMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddLogging adds logging middleware
// func (b *MiddlewareBuilder) AddLogging() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateLoggingMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddValidation adds validation middleware
// func (b *MiddlewareBuilder) AddValidation() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateValidationMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddRateLimit adds rate limit middleware
// func (b *MiddlewareBuilder) AddRateLimit() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateRateLimitMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddHealth adds health middleware
// func (b *MiddlewareBuilder) AddHealth() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateHealthMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddTracing adds tracing middleware
// func (b *MiddlewareBuilder) AddTracing() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateTracingMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// AddCompression adds compression middleware
// func (b *MiddlewareBuilder) AddCompression() *MiddlewareBuilder {
// 	if middleware, err := b.factory.CreateCompressionMiddleware(); err == nil {
// 		b.chain.Add(middleware)
// 	}
// 	return b
// }

// Build returns the middleware chain
func (b *MiddlewareBuilder) Build() *core.Chain {
	return b.chain
}

// GetChain returns the middleware chain
func (b *MiddlewareBuilder) GetChain() *core.Chain {
	return b.chain
}

// GetFactory returns the middleware factory
func (b *MiddlewareBuilder) GetFactory() *MiddlewareFactory {
	return b.factory
}

// GetConfig returns the builder configuration
func (b *MiddlewareBuilder) GetConfig() *Config {
	return b.config
}

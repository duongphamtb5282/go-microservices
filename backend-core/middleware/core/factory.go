package core

// MiddlewareFactory represents a middleware factory
type MiddlewareFactory interface {
	// CreateCacheMiddleware creates cache middleware
	CreateCacheMiddleware() (Middleware, error)

	// CreateAuthMiddleware creates auth middleware
	CreateAuthMiddleware() (Middleware, error)

	// CreateMetricsMiddleware creates metrics middleware
	CreateMetricsMiddleware() (Middleware, error)

	// CreateCorsMiddleware creates CORS middleware
	CreateCorsMiddleware() (Middleware, error)

	// CreateLoggingMiddleware creates logging middleware
	CreateLoggingMiddleware() (Middleware, error)

	// CreateValidationMiddleware creates validation middleware
	CreateValidationMiddleware() (Middleware, error)
}

// MiddlewareBuilder represents a middleware builder
type MiddlewareBuilder interface {
	// AddCache adds cache middleware
	AddCache() MiddlewareBuilder

	// AddAuth adds auth middleware
	AddAuth() MiddlewareBuilder

	// AddMetrics adds metrics middleware
	AddMetrics() MiddlewareBuilder

	// AddCors adds CORS middleware
	AddCors() MiddlewareBuilder

	// AddLogging adds logging middleware
	AddLogging() MiddlewareBuilder

	// AddValidation adds validation middleware
	AddValidation() MiddlewareBuilder

	// AddCustom adds custom middleware
	AddCustom(middleware Middleware) MiddlewareBuilder

	// Build builds the middleware chain
	Build() MiddlewareChain
}

package core

// FrameworkAdapter represents a framework adapter
type FrameworkAdapter interface {
	// Adapt adapts a middleware to the framework
	Adapt(middleware Middleware) interface{}

	// AdaptChain adapts a middleware chain to the framework
	AdaptChain(chain MiddlewareChain) interface{}
}

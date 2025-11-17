package echo

import (
	"context"
	"net/http"

	"backend-core/middleware/core"

	"github.com/labstack/echo/v4"
)

// EchoAdapter adapts middleware to Echo framework
type EchoAdapter struct {
	chain *core.Chain
}

// NewEchoAdapter creates a new Echo adapter
func NewEchoAdapter(chain *core.Chain) *EchoAdapter {
	return &EchoAdapter{
		chain: chain,
	}
}

// Adapt adapts a middleware to Echo
func (a *EchoAdapter) Adapt(middleware core.Middleware) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := &core.Request{
				Request:  c.Request(),
				Context:  c.Request().Context(),
				Metadata: make(map[string]interface{}),
			}

			// Add Echo context to metadata
			req.Metadata["echo_context"] = c

			resp, err := middleware.Execute(req.Context, req, func(ctx context.Context, req *core.Request) (*core.Response, error) {
				err := next(c)
				// Create a mock HTTP response for Echo
				httpResp := &http.Response{
					StatusCode: c.Response().Status,
					Header:     c.Response().Header(),
				}
				return &core.Response{
					Response: httpResp,
					Context:  ctx,
					Metadata: make(map[string]interface{}),
					Error:    err,
				}, err
			})

			if err != nil {
				return err
			}

			if resp != nil {
				// Apply response metadata
				for key, value := range resp.Metadata {
					c.Set(key, value)
				}
			}

			return nil
		}
	}
}

// AdaptChain adapts a middleware chain to Echo
func (a *EchoAdapter) AdaptChain() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := &core.Request{
				Request:  c.Request(),
				Context:  c.Request().Context(),
				Metadata: make(map[string]interface{}),
			}

			// Add Echo context to metadata
			req.Metadata["echo_context"] = c

			resp, err := a.chain.Execute(req.Context, req)

			if err != nil {
				return err
			}

			if resp != nil {
				// Apply response metadata
				for key, value := range resp.Metadata {
					c.Set(key, value)
				}
			}

			return next(c)
		}
	}
}

// Chain returns the middleware chain
func (a *EchoAdapter) Chain() *core.Chain {
	return a.chain
}

// EchoMiddleware represents an Echo-specific middleware
type EchoMiddleware struct {
	name     string
	priority int
	handler  echo.MiddlewareFunc
}

// NewEchoMiddleware creates a new Echo middleware
func NewEchoMiddleware(name string, priority int, handler echo.MiddlewareFunc) *EchoMiddleware {
	return &EchoMiddleware{
		name:     name,
		priority: priority,
		handler:  handler,
	}
}

// Name returns the middleware name
func (m *EchoMiddleware) Name() string {
	return m.name
}

// Priority returns the middleware priority
func (m *EchoMiddleware) Priority() int {
	return m.priority
}

// Execute executes the Echo middleware
func (m *EchoMiddleware) Execute(ctx context.Context, req *core.Request, next core.NextHandler) (*core.Response, error) {
	// Convert to Echo context
	echoCtx := req.Metadata["echo_context"].(echo.Context)

	// Execute Echo handler
	err := m.handler(func(c echo.Context) error {
		return nil
	})(echoCtx)

	if err != nil {
		return nil, err
	}

	// Continue with next handler
	return next(ctx, req)
}

// EchoChain represents an Echo-specific middleware chain
type EchoChain struct {
	*core.Chain
	app *echo.Echo
}

// NewEchoChain creates a new Echo chain
func NewEchoChain(app *echo.Echo) *EchoChain {
	return &EchoChain{
		Chain: core.NewChain(),
		app:   app,
	}
}

// Use adds a middleware to the Echo chain
func (c *EchoChain) Use(middleware echo.MiddlewareFunc) *EchoChain {
	c.app.Use(middleware)
	return c
}

// UseMiddleware adds a core middleware to the Echo chain
func (c *EchoChain) UseMiddleware(middleware core.Middleware) *EchoChain {
	c.Chain.Add(middleware)
	return c
}

// Build builds the Echo middleware chain
func (c *EchoChain) Build() *echo.Echo {
	return c.app
}

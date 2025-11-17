package gin

import (
	"context"
	"net/http"

	"backend-core/middleware/core"

	"github.com/gin-gonic/gin"
)

// GinAdapter adapts middleware to Gin framework
type GinAdapter struct {
	chain *core.Chain
}

// NewGinAdapter creates a new Gin adapter
func NewGinAdapter(chain *core.Chain) *GinAdapter {
	return &GinAdapter{
		chain: chain,
	}
}

// Adapt adapts a middleware to Gin
func (a *GinAdapter) Adapt(middleware core.Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &core.Request{
			Request:  c.Request,
			Context:  c.Request.Context(),
			Metadata: make(map[string]interface{}),
		}

		// Add Gin context to metadata
		req.Metadata["gin_context"] = c

		resp, err := middleware.Execute(req.Context, req, func(ctx context.Context, req *core.Request) (*core.Response, error) {
			c.Next()
			return &core.Response{
				Response: &http.Response{
					StatusCode: c.Writer.Status(),
					Header:     c.Writer.Header(),
				},
				Context:  ctx,
				Metadata: make(map[string]interface{}),
			}, nil
		})

		if err != nil {
			c.AbortWithError(500, err)
		}

		if resp != nil {
			// Apply response metadata
			for key, value := range resp.Metadata {
				c.Set(key, value)
			}
		}
	}
}

// AdaptChain adapts a middleware chain to Gin
func (a *GinAdapter) AdaptChain() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := &core.Request{
			Request:  c.Request,
			Context:  c.Request.Context(),
			Metadata: make(map[string]interface{}),
		}

		// Add Gin context to metadata
		req.Metadata["gin_context"] = c

		resp, err := a.chain.Execute(req.Context, req)

		if err != nil {
			c.AbortWithError(500, err)
		}

		if resp != nil {
			// Apply response metadata
			for key, value := range resp.Metadata {
				c.Set(key, value)
			}
		}
	}
}

// Chain returns the middleware chain
func (a *GinAdapter) Chain() *core.Chain {
	return a.chain
}

// GinMiddleware represents a Gin-specific middleware
type GinMiddleware struct {
	name     string
	priority int
	handler  gin.HandlerFunc
}

// NewGinMiddleware creates a new Gin middleware
func NewGinMiddleware(name string, priority int, handler gin.HandlerFunc) *GinMiddleware {
	return &GinMiddleware{
		name:     name,
		priority: priority,
		handler:  handler,
	}
}

// Name returns the middleware name
func (m *GinMiddleware) Name() string {
	return m.name
}

// Priority returns the middleware priority
func (m *GinMiddleware) Priority() int {
	return m.priority
}

// Execute executes the Gin middleware
func (m *GinMiddleware) Execute(ctx context.Context, req *core.Request, next core.NextHandler) (*core.Response, error) {
	// Convert to Gin context
	ginCtx := req.Metadata["gin_context"].(*gin.Context)

	// Execute Gin handler
	m.handler(ginCtx)

	// Continue with next handler
	return next(ctx, req)
}

// GinChain represents a Gin-specific middleware chain
type GinChain struct {
	*core.Chain
	router *gin.Engine
}

// NewGinChain creates a new Gin chain
func NewGinChain(router *gin.Engine) *GinChain {
	return &GinChain{
		Chain:  core.NewChain(),
		router: router,
	}
}

// Use adds a middleware to the Gin chain
func (c *GinChain) Use(middleware gin.HandlerFunc) *GinChain {
	c.router.Use(middleware)
	return c
}

// UseMiddleware adds a core middleware to the Gin chain
func (c *GinChain) UseMiddleware(middleware core.Middleware) *GinChain {
	c.Chain.Add(middleware)
	return c
}

// Build builds the Gin middleware chain
func (c *GinChain) Build() *gin.Engine {
	return c.router
}

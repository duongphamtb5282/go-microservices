package groups

import (
	"auth-service/src/interfaces/rest/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

// SystemRoutes defines system-related routes (health, cache, etc.)
type SystemRoutes struct {
	cacheMiddleware *middleware.CacheMiddleware
}

// NewSystemRoutes creates a new system routes group
func NewSystemRoutes(cacheMiddleware *middleware.CacheMiddleware) *SystemRoutes {
	return &SystemRoutes{
		cacheMiddleware: cacheMiddleware,
	}
}

// RegisterRoutes registers all system routes
func (r *SystemRoutes) RegisterRoutes(router *gin.Engine) {
	// Debug route first
	router.GET("/debug", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Debug route from systemRoutes"})
	})

	// Swagger documentation
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		content, err := os.ReadFile("docs/swagger.json")
		if err != nil {
			println("Error reading swagger.json:", err.Error())
			c.JSON(500, gin.H{"error": "Failed to read swagger spec"})
			return
		}
		println("Successfully served swagger.json, size:", len(content))
		c.Header("Content-Type", "application/json")
		c.Data(200, "application/json", content)
	})

	// Swagger UI - serve basic HTML page for swagger UI
	router.GET("/swagger", func(c *gin.Context) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Auth Service API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
		c.Header("Content-Type", "text/html")
		c.String(200, html)
	})

	// Swagger UI - catch-all for other swagger files (must come after specific routes)
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Test route
	router.GET("/test", func(c *gin.Context) {
		println("Test route hit")
		c.JSON(200, gin.H{"message": "test route works"})
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		println("Health route hit - modified version")
		c.JSON(200, gin.H{
			"status":    "healthy-modified",
			"cache":     "enabled",
			"decorator": "active",
			"strategy":  "write_through",
			"unique_id": "test-12345",
		})
	})

	// Cache management endpoints
	if r.cacheMiddleware != nil {
		router.GET("/cache/stats", r.cacheMiddleware.CacheStatsHandler())
		router.GET("/cache/health", r.cacheMiddleware.CacheHealthHandler())
		router.POST("/cache/clear", r.cacheMiddleware.CacheClearHandler())
	} else {
		// Return cache unavailable responses
		router.GET("/cache/stats", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"success": false,
				"error":   "Cache service unavailable",
			})
		})
		router.GET("/cache/health", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"success": false,
				"error":   "Cache service unavailable",
			})
		})
		router.POST("/cache/clear", func(c *gin.Context) {
			c.JSON(503, gin.H{
				"success": false,
				"error":   "Cache service unavailable",
			})
		})
	}
}

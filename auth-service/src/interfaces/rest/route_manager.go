package handlers

import (
	"auth-service/src/interfaces/rest/groups"
	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// RouteManager manages all application routes
type RouteManager struct {
	authHandler     *handlers.AuthHandler
	userHandler     *handlers.UserHandler
	cacheMiddleware *middleware.CacheMiddleware
	logger          *logging.Logger
}

// NewRouteManager creates a new route manager
func NewRouteManager(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	cacheMiddleware *middleware.CacheMiddleware,
	logger *logging.Logger,
) *RouteManager {
	return &RouteManager{
		authHandler:     authHandler,
		userHandler:     userHandler,
		cacheMiddleware: cacheMiddleware,
		logger:          logger,
	}
}

// SetupRoutes configures all application routes
func (rm *RouteManager) SetupRoutes() *gin.Engine {
	// Debug: log that setup was called
	println("RouteManager SetupRoutes called")
	// Create Gin engine
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Initialize route groups
	authRoutes := groups.NewAuthRoutes(rm.authHandler)
	userRoutes := groups.NewUserRoutes(rm.userHandler, rm.cacheMiddleware)
	systemRoutes := groups.NewSystemRoutes(rm.cacheMiddleware)

	// Register system routes (health, cache, etc.)
	systemRoutes.RegisterRoutes(router)

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Status endpoint
		api.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "auth-service",
				"version": "1.0.0",
				"status":  "running",
			})
		})

		// Register auth routes
		authRoutes.RegisterRoutes(api)

		// Register user routes
		userRoutes.RegisterRoutes(api)
	}

	rm.logger.Info("All routes registered successfully")
	return router
}

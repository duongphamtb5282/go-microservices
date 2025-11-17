package router

import (
	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// Router configures HTTP routes
type Router struct {
	userHandler  *handlers.UserHandler
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware
	logger       *logging.Logger
}

// NewRouter creates a new router
func NewRouter(
	userHandler *handlers.UserHandler,
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware,
	logger *logging.Logger,
) *Router {
	return &Router{
		userHandler:  userHandler,
		keycloakAuth: keycloakAuth,
		logger:       logger,
	}
}

// RouterInstance creates a new router (without New prefix)
func RouterInstance(
	userHandler *handlers.UserHandler,
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware,
	logger *logging.Logger,
) *Router {
	return &Router{
		userHandler:  userHandler,
		keycloakAuth: keycloakAuth,
		logger:       logger,
	}
}

// SetupRoutes configures all HTTP routes using domain-based structure
func (r *Router) SetupRoutes() *gin.Engine {
	// Create Gin engine
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Create API group (all routes public for testing)
	apiGroup := router.Group("/api/v1")

	// Health check endpoint
	apiGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "auth-service",
			"status":  "ok",
		})
	})

	// User management endpoints (public for testing)
	apiGroup.POST("/users", r.userHandler.CreateUser)
	apiGroup.GET("/users/:username", r.userHandler.GetUser)
	apiGroup.GET("/users", r.userHandler.ListUsers)
	// TODO: Implement missing methods in UserHandler
	// apiGroup.PUT("/users/:username", r.userHandler.UpdateUser)
	// apiGroup.DELETE("/users/:username", r.userHandler.DeleteUser)

	// Auth endpoints (public for testing)
	// TODO: Implement missing methods in UserHandler
	// apiGroup.POST("/auth/verify-email", r.userHandler.VerifyEmail)
	// apiGroup.POST("/auth/resend-verification", r.userHandler.ResendVerification)
	apiGroup.POST("/auth/activate", r.userHandler.ActivateUser)
	// apiGroup.POST("/auth/deactivate", r.userHandler.DeactivateUser)

	// System endpoints
	apiGroup.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "auth-service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Admin endpoints (protected by Keycloak roles)
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(r.keycloakAuth.RequireKeycloakRole([]string{"admin"}))

	// Admin user management
	adminGroup.GET("/users", r.userHandler.ListUsers)
	adminGroup.GET("/users/:username", r.userHandler.GetUser)
	adminGroup.POST("/users", r.userHandler.CreateUser)
	// TODO: Add more admin endpoints as needed
	// adminGroup.PUT("/users/:username", r.userHandler.UpdateUser)
	// adminGroup.DELETE("/users/:username", r.userHandler.DeleteUser)

	return router
}

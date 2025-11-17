package groups

import (
	"fmt"

	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

// AdminRoutes handles admin-specific routes with Keycloak role-based authorization
type AdminRoutes struct {
	userHandler  *handlers.UserHandler
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware
}

// NewAdminRoutes creates a new admin routes group
func NewAdminRoutes(userHandler *handlers.UserHandler, keycloakAuth *middleware.KeycloakAuthorizationMiddleware) *AdminRoutes {
	return &AdminRoutes{
		userHandler:  userHandler,
		keycloakAuth: keycloakAuth,
	}
}

// RegisterRoutes registers admin routes with role-based protection
func (ar *AdminRoutes) RegisterRoutes(router *gin.RouterGroup) {
	// Debug logging
	fmt.Printf("DEBUG: Registering admin routes at path: %s\n", router.BasePath())
	fmt.Printf("DEBUG: Keycloak auth middleware is nil: %v\n", ar.keycloakAuth == nil)

	// Create admin subgroup with admin role requirement
	admin := router.Group("/admin")

	// Only add middleware if Keycloak auth is available
	if ar.keycloakAuth != nil {
		admin.Use(ar.keycloakAuth.RequireKeycloakRole([]string{"admin"}))
		fmt.Printf("DEBUG: Keycloak middleware added to admin routes\n")
	} else {
		fmt.Printf("DEBUG: Keycloak middleware not available, registering admin routes without protection\n")
	}

	// Debug logging
	fmt.Printf("DEBUG: Admin routes group created: %s/admin\n", router.BasePath())

	// Admin user management endpoints
	{
		fmt.Printf("DEBUG: Adding GET /users route\n")
		admin.GET("/users", func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /users endpoint called\n")
			c.JSON(200, gin.H{"message": "Admin users endpoint", "status": "success"})
		})
		fmt.Printf("DEBUG: Adding GET /test route\n")
		admin.GET("/test", func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /test endpoint called\n")
			c.JSON(200, gin.H{"message": "Admin test endpoint", "status": "success"})
		})
		// TODO: Add proper admin endpoints
		// admin.GET("/users", ar.userHandler.ListUsers)
		// admin.GET("/users/:username", ar.userHandler.GetUser)
		// admin.POST("/users", ar.userHandler.CreateUser)
	}

	fmt.Printf("DEBUG: Admin routes registered successfully\n")
}

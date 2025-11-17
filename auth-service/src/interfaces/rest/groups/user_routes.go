package groups

import (
	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes defines user-related routes
type UserRoutes struct {
	userHandler     *handlers.UserHandler
	cacheMiddleware *middleware.CacheMiddleware
}

// NewUserRoutes creates a new user routes group
func NewUserRoutes(userHandler *handlers.UserHandler, cacheMiddleware *middleware.CacheMiddleware) *UserRoutes {
	return &UserRoutes{
		userHandler:     userHandler,
		cacheMiddleware: cacheMiddleware,
	}
}

// RegisterRoutes registers all user routes
func (r *UserRoutes) RegisterRoutes(router *gin.RouterGroup) {
	userCacheConfig := middleware.UserCacheConfig()

	// Check if cache middleware is available
	if r.cacheMiddleware != nil {
		// User routes with cache strategies
		router.GET("/users/:id", r.cacheMiddleware.CacheGet(userCacheConfig), r.userHandler.GetUser)
		router.GET("/users", r.cacheMiddleware.CacheList(userCacheConfig), r.userHandler.ListUsers)

		// Apply cache invalidation to mutation endpoints
		router.POST("/users", r.cacheMiddleware.InvalidateCache(userCacheConfig), r.userHandler.CreateUser)
		// TODO: Implement UpdateUser and DeleteUser methods in UserHandler
		// router.PUT("/users/:id", r.cacheMiddleware.InvalidateCache(userCacheConfig), r.userHandler.UpdateUser)
		// router.DELETE("/users/:id", r.cacheMiddleware.InvalidateCache(userCacheConfig), r.userHandler.DeleteUser)
	} else {
		// Fallback to routes without cache middleware
		router.GET("/users/:id", r.userHandler.GetUser)
		router.GET("/users", r.userHandler.ListUsers)
		router.POST("/users", r.userHandler.CreateUser)
		// TODO: Implement UpdateUser and DeleteUser methods in UserHandler
		// router.PUT("/users/:id", r.userHandler.UpdateUser)
		// router.DELETE("/users/:id", r.userHandler.DeleteUser)
	}
}

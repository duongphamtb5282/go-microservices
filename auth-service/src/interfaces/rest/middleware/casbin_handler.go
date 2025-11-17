package middleware

import (
	"backend-core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CasbinHandler middleware for RBAC authorization
func CasbinHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Casbin RBAC authorization
		// For now, just allow all authenticated users

		// Get user ID from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get requested resource and action
		resource := c.Request.URL.Path
		action := c.Request.Method

		// TODO: Check permissions using Casbin
		// For now, just allow all authenticated users
		// TODO: Pass logger from context or create a global logger
		_ = userID
		_ = resource
		_ = action

		c.Next()
	}
}

// CasbinHandlerWithEnforcer creates Casbin handler with enforcer
func CasbinHandlerWithEnforcer(enforcer interface{}, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Get requested resource and action
		resource := c.Request.URL.Path
		action := c.Request.Method

		// TODO: Implement actual Casbin authorization check
		// allowed, err := enforcer.Enforce(userID, resource, action)
		// if err != nil || !allowed {
		//     logger.Error("Authorization failed", %v)
		//     c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		//     c.Abort()
		//     return
		// }

		logger.Info("Authorization check",
			logging.String("user_id", userID.(string)),
			logging.String("resource", resource),
			logging.String("action", action))

		c.Next()
	}
}

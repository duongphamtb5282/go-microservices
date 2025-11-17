package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// OperationRecord middleware for logging operations
func OperationRecord() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get user information from context
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")

		// Log operation start
		// TODO: Pass logger from context or create a global logger
		// For now, we'll skip logging until logger is properly injected

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log operation completion
		// TODO: Pass logger from context or create a global logger
		// For now, we'll skip logging until logger is properly injected
		_ = duration
		_ = userID
		_ = userEmail
	}
}

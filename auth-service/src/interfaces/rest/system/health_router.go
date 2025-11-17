package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthRouter struct{}

// InitHealthRouter initializes health check routes (public)
func (s *HealthRouter) InitHealthRouter(PublicGroup *gin.RouterGroup) {
	healthRouter := PublicGroup.Group("/health")
	{
		healthRouter.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "auth-service"})
		})
		healthRouter.GET("/detailed", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "auth-service",
				"version": "1.0.0",
				"components": gin.H{
					"database": "connected",
					"cache":    "connected",
					"kafka":    "connected",
				},
			})
		})
		healthRouter.GET("/ready", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
		})
		healthRouter.GET("/live", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "alive"})
		})
	}
}

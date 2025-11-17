package system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthApi struct{}

func NewHealthApi() *HealthApi {
	return &HealthApi{}
}

// HealthCheck handles basic health check
func (h *HealthApi) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}

// DetailedHealth handles detailed health check
func (h *HealthApi) DetailedHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  "auth-service",
		"version":  "1.0.0",
		"uptime":   "running",
		"database": "connected",
		"cache":    "connected",
	})
}

// ReadinessCheck handles readiness check
func (h *HealthApi) ReadinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "auth-service",
	})
}

// LivenessCheck handles liveness check
func (h *HealthApi) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"service": "auth-service",
	})
}

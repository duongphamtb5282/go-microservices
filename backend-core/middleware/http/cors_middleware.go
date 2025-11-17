package http

import (
	"backend-core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware provides CORS functionality
type CORSMiddleware struct {
	config *CORSConfig
	logger *logging.Logger
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config *CORSConfig, logger *logging.Logger) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig()
	}
	return &CORSMiddleware{
		config: config,
		logger: logger,
	}
}

// Handler returns the CORS middleware handler
func (c *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("Origin")

		// Check if origin is allowed
		if c.isOriginAllowed(origin) {
			ctx.Header("Access-Control-Allow-Origin", origin)
		} else if len(c.config.AllowOrigins) == 1 && c.config.AllowOrigins[0] == "*" {
			ctx.Header("Access-Control-Allow-Origin", "*")
		}

		// Set CORS headers
		ctx.Header("Access-Control-Allow-Methods", c.joinStrings(c.config.AllowMethods))
		ctx.Header("Access-Control-Allow-Headers", c.joinStrings(c.config.AllowHeaders))
		ctx.Header("Access-Control-Expose-Headers", c.joinStrings(c.config.ExposeHeaders))
		ctx.Header("Access-Control-Max-Age", string(rune(c.config.MaxAge)))

		if c.config.AllowCredentials {
			ctx.Header("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if ctx.Request.Method == "OPTIONS" {
			c.logger.Debug("Handling CORS preflight request",
				logging.String("origin", origin),
				logging.String("method", ctx.Request.Method))
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range c.config.AllowOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}
	}
	return false
}

// joinStrings joins a slice of strings with comma separator
func (c *CORSMiddleware) joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

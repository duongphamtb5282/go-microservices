package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend-core/cache/decorators"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// CacheConfig defines cache configuration for a resource
type CacheConfig struct {
	ResourceName string        // e.g., "user", "role", "permission"
	IDParam      string        // URL parameter name for resource ID, e.g., "id", "username"
	ListKey      string        // Cache key prefix for lists, e.g., "users:list"
	TTL          time.Duration // Time to live for cached items
	ListTTL      time.Duration // Time to live for cached lists
}

// CacheMiddleware provides generic cache functionality for any resource
type CacheMiddleware struct {
	cacheDecorator *decorators.CacheDecorator
	logger         *logging.Logger
}

// NewCacheMiddleware creates a new cache middleware
func NewCacheMiddleware(cacheDecorator *decorators.CacheDecorator, logger *logging.Logger) *CacheMiddleware {
	return &CacheMiddleware{
		cacheDecorator: cacheDecorator,
		logger:         logger,
	}
}

// CacheGet provides generic cache for GET operations by ID
func (m *CacheMiddleware) CacheGet(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID := c.Param(config.IDParam)
		if resourceID == "" {
			c.Next()
			return
		}

		cacheKey := fmt.Sprintf("%s:%s", config.ResourceName, resourceID)
		m.logger.Debug("Checking cache for resource",
			logging.String("resource", config.ResourceName),
			logging.String("key", cacheKey))

		// Try to get from cache
		var cachedData interface{}
		err := m.cacheDecorator.Get(context.Background(), cacheKey, &cachedData)
		if err == nil {
			m.logger.Debug("Cache hit for resource",
				logging.String("resource", config.ResourceName),
				logging.String("key", cacheKey))
			c.JSON(200, gin.H{
				"success": true,
				"data":    cachedData,
				"cached":  true,
			})
			c.Abort()
			return
		}

		m.logger.Debug("Cache miss for resource",
			logging.String("resource", config.ResourceName),
			logging.String("key", cacheKey))

		// Store original response writer
		originalWriter := c.Writer

		// Create a custom response writer to capture the response
		responseWriter := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		// Continue with the next handler
		c.Next()

		// If the response was successful, cache it
		if responseWriter.statusCode == 200 {
			m.logger.Debug("Caching resource response",
				logging.String("resource", config.ResourceName),
				logging.String("key", cacheKey))
			if err := m.cacheDecorator.Set(context.Background(), cacheKey, responseWriter.body, config.TTL); err != nil {
				m.logger.Error("Failed to cache resource response",
					logging.String("resource", config.ResourceName),
					logging.Error(err))
			}
		}

		// Write the response to the original writer
		originalWriter.WriteHeader(responseWriter.statusCode)
		originalWriter.Write(responseWriter.body)
	}
}

// CacheList provides generic cache for GET list operations
func (m *CacheMiddleware) CacheList(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Build cache key from query parameters
		queryParams := c.Request.URL.Query()
		var keyParts []string

		// Add resource list prefix
		keyParts = append(keyParts, config.ListKey)

		// Add relevant query parameters
		for _, param := range []string{"page", "limit", "search", "sort", "order"} {
			if value := queryParams.Get(param); value != "" {
				keyParts = append(keyParts, fmt.Sprintf("%s:%s", param, value))
			}
		}

		cacheKey := strings.Join(keyParts, ":")

		m.logger.Debug("Checking cache for resource list",
			logging.String("resource", config.ResourceName),
			logging.String("key", cacheKey))

		// Try to get from cache
		var cachedList interface{}
		err := m.cacheDecorator.Get(context.Background(), cacheKey, &cachedList)
		if err == nil {
			m.logger.Debug("Cache hit for resource list",
				logging.String("resource", config.ResourceName),
				logging.String("key", cacheKey))
			c.JSON(200, gin.H{
				"success": true,
				"data":    cachedList,
				"cached":  true,
			})
			c.Abort()
			return
		}

		m.logger.Debug("Cache miss for resource list",
			logging.String("resource", config.ResourceName),
			logging.String("key", cacheKey))

		// Store original response writer
		originalWriter := c.Writer

		// Create a custom response writer to capture the response
		responseWriter := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		// Continue with the next handler
		c.Next()

		// If the response was successful, cache it
		if responseWriter.statusCode == 200 {
			m.logger.Debug("Caching resource list response",
				logging.String("resource", config.ResourceName),
				logging.String("key", cacheKey))
			if err := m.cacheDecorator.Set(context.Background(), cacheKey, responseWriter.body, config.ListTTL); err != nil {
				m.logger.Error("Failed to cache resource list response",
					logging.String("resource", config.ResourceName),
					logging.Error(err))
			}
		}

		// Write the response to the original writer
		originalWriter.WriteHeader(responseWriter.statusCode)
		originalWriter.Write(responseWriter.body)
	}
}

// InvalidateCache invalidates cache for a specific resource
func (m *CacheMiddleware) InvalidateCache(config CacheConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue with the next handler first
		c.Next()

		// If the response was successful, invalidate cache
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			m.logger.Debug("Invalidating resource cache",
				logging.String("resource", config.ResourceName))

			// Invalidate specific resource cache if we have resource ID
			if resourceID := c.Param(config.IDParam); resourceID != "" {
				cacheKey := fmt.Sprintf("%s:%s", config.ResourceName, resourceID)
				if err := m.cacheDecorator.Delete(context.Background(), cacheKey); err != nil {
					m.logger.Error("Failed to invalidate resource cache",
						logging.String("resource", config.ResourceName),
						logging.Error(err))
				}
			}

			// Invalidate resource list cache
			pattern := fmt.Sprintf("%s*", config.ListKey)
			if err := m.cacheDecorator.DeletePattern(context.Background(), pattern); err != nil {
				m.logger.Error("Failed to invalidate resource list cache",
					logging.String("resource", config.ResourceName),
					logging.Error(err))
			}
		}
	}
}

// Helper functions to create cache configurations for common resources

// UserCacheConfig returns cache configuration for user resources
func UserCacheConfig() CacheConfig {
	return CacheConfig{
		ResourceName: "user",
		IDParam:      "id",
		ListKey:      "users:list",
		TTL:          10 * time.Minute, // Individual users cached for 10 minutes
		ListTTL:      2 * time.Minute,  // User lists cached for 2 minutes
	}
}

// RoleCacheConfig returns cache configuration for role resources
func RoleCacheConfig() CacheConfig {
	return CacheConfig{
		ResourceName: "role",
		IDParam:      "id",
		ListKey:      "roles:list",
		TTL:          15 * time.Minute, // Roles cached for 15 minutes
		ListTTL:      5 * time.Minute,  // Role lists cached for 5 minutes
	}
}

// PermissionCacheConfig returns cache configuration for permission resources
func PermissionCacheConfig() CacheConfig {
	return CacheConfig{
		ResourceName: "permission",
		IDParam:      "id",
		ListKey:      "permissions:list",
		TTL:          30 * time.Minute, // Permissions cached for 30 minutes (less likely to change)
		ListTTL:      10 * time.Minute, // Permission lists cached for 10 minutes
	}
}

// Convenience methods for backward compatibility (deprecated - use CacheGet/CacheList instead)

// CacheUserGet provides cache for GET /users/:id (deprecated)
func (m *CacheMiddleware) CacheUserGet() gin.HandlerFunc {
	return m.CacheGet(UserCacheConfig())
}

// CacheUserList provides cache for GET /users (deprecated)
func (m *CacheMiddleware) CacheUserList() gin.HandlerFunc {
	return m.CacheList(UserCacheConfig())
}

// InvalidateUserCache invalidates user-related cache (deprecated)
func (m *CacheMiddleware) InvalidateUserCache() gin.HandlerFunc {
	return m.InvalidateCache(UserCacheConfig())
}

// CacheStatsHandler provides cache statistics
func (m *CacheMiddleware) CacheStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := m.cacheDecorator.GetCacheStats(context.Background())
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Failed to get cache statistics",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"data":    stats,
		})
	}
}

// CacheHealthHandler provides cache health check
func (m *CacheMiddleware) CacheHealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := m.cacheDecorator.HealthCheck(context.Background())
		if err != nil {
			c.JSON(503, gin.H{
				"success": false,
				"error":   "Cache is unhealthy",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"status":  "healthy",
		})
	}
}

// CacheClearHandler provides cache clear functionality
func (m *CacheMiddleware) CacheClearHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" {
			c.JSON(405, gin.H{
				"success": false,
				"error":   "Method not allowed",
			})
			return
		}

		// Clear all cache
		if err := m.cacheDecorator.DeletePattern(context.Background(), "*"); err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"error":   "Failed to clear cache",
			})
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"message": "Cache cleared successfully",
		})
	}
}

// CacheResponseWriter is a custom response writer that captures the response
type CacheResponseWriter struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

// Write captures the response body
func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

// WriteHeader captures the status code
func (w *CacheResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Status returns the status code
func (w *CacheResponseWriter) Status() int {
	return w.statusCode
}

// Body returns the response body
func (w *CacheResponseWriter) Body() []byte {
	return w.body
}

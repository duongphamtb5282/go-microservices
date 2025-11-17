package cache

import (
	"context"
	"net/http"
	"strings"
	"time"

	"backend-core/cache/decorators"
	"backend-core/logging"
)

// HTTPCacheMiddleware provides HTTP middleware for caching
type HTTPCacheMiddleware struct {
	cacheDecorator *decorators.CacheDecorator
	logger         *logging.Logger
}

// NewHTTPCacheMiddleware creates a new HTTP cache middleware
func NewHTTPCacheMiddleware(cacheDecorator *decorators.CacheDecorator, logger *logging.Logger) *HTTPCacheMiddleware {
	return &HTTPCacheMiddleware{
		cacheDecorator: cacheDecorator,
		logger:         logger,
	}
}

// CacheGet middleware caches GET requests
func (m *HTTPCacheMiddleware) CacheGet(ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Build cache key from request
			cacheKey := m.buildCacheKey(r)

			// Try to get from cache
			var cachedResponse interface{}
			err := m.cacheDecorator.Get(r.Context(), cacheKey, &cachedResponse)
			if err == nil {
				m.logger.Info("Response served from cache",
					logging.String("method", r.Method),
					logging.String("path", r.URL.Path),
					logging.String("cache_key", cacheKey))

				// Write cached response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// In a real implementation, you would serialize cachedResponse to JSON
				// w.Write(cachedResponse.([]byte))
				return
			}

			// Cache miss - log and continue
			if err != nil {
				m.logger.Debug("Cache miss for GET request",
					logging.String("cache_key", cacheKey),
					logging.Error(err))
			}

			// Create a response writer that captures the response
			responseWriter := &cacheResponseWriter{
				ResponseWriter: w,
				cacheKey:       cacheKey,
				ttl:            ttl,
				middleware:     m,
			}

			// Continue to the next handler
			next.ServeHTTP(responseWriter, r)
		})
	}
}

// CacheEntity middleware caches responses based on entity type and ID
func (m *HTTPCacheMiddleware) CacheEntity(entityType string, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Extract entity ID from URL path
			entityID := m.extractEntityID(r.URL.Path)
			if entityID == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Build cache key
			cacheKey := m.buildEntityCacheKey(entityType, entityID)

			// Try to get from cache
			var cachedResponse interface{}
			err := m.cacheDecorator.Get(r.Context(), cacheKey, &cachedResponse)
			if err == nil {
				m.logger.Info("Entity response served from cache",
					logging.String("entity_type", entityType),
					logging.String("entity_id", entityID),
					logging.String("cache_key", cacheKey))

				// Write cached response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				// In a real implementation, you would serialize cachedResponse to JSON
				// w.Write(cachedResponse.([]byte))
				return
			}

			// Cache miss - log and continue
			if err != nil {
				m.logger.Debug("Cache miss for entity request",
					logging.String("cache_key", cacheKey),
					logging.Error(err))
			}

			// Create a response writer that captures the response
			responseWriter := &entityCacheResponseWriter{
				ResponseWriter: w,
				entityType:     entityType,
				entityID:       entityID,
				cacheKey:       cacheKey,
				ttl:            ttl,
				middleware:     m,
			}

			// Continue to the next handler
			next.ServeHTTP(responseWriter, r)
		})
	}
}

// InvalidateCache middleware invalidates cache for non-GET requests
func (m *HTTPCacheMiddleware) InvalidateCache() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Continue to the next handler first
			next.ServeHTTP(w, r)

			// Only invalidate for non-GET requests
			if r.Method == http.MethodGet {
				return
			}

			// Invalidate cache based on the request
			m.invalidateCacheForRequest(r)
		})
	}
}

// buildCacheKey creates a cache key from the HTTP request
func (m *HTTPCacheMiddleware) buildCacheKey(r *http.Request) string {
	// Simple cache key based on method, path, and query parameters
	key := r.Method + ":" + r.URL.Path
	if r.URL.RawQuery != "" {
		key += "?" + r.URL.RawQuery
	}
	return key
}

// buildEntityCacheKey creates a cache key for entity-based caching
func (m *HTTPCacheMiddleware) buildEntityCacheKey(entityType, entityID string) string {
	return "entity:" + entityType + ":" + entityID
}

// extractEntityID extracts entity ID from URL path
func (m *HTTPCacheMiddleware) extractEntityID(path string) string {
	// Simple extraction - assumes entity ID is the last segment
	// In a real implementation, you might use a more sophisticated approach
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// invalidateCacheForRequest invalidates cache entries related to the request
func (m *HTTPCacheMiddleware) invalidateCacheForRequest(r *http.Request) {
	// Invalidate general cache for this path
	cacheKey := m.buildCacheKey(r)
	err := m.cacheDecorator.Delete(r.Context(), cacheKey)
	if err != nil {
		m.logger.Debug("Failed to invalidate cache",
			logging.String("cache_key", cacheKey),
			logging.Error(err))
	}

	// Invalidate entity cache if applicable
	entityID := m.extractEntityID(r.URL.Path)
	if entityID != "" {
		// Try to determine entity type from path
		entityType := m.extractEntityType(r.URL.Path)
		if entityType != "" {
			entityCacheKey := m.buildEntityCacheKey(entityType, entityID)
			err := m.cacheDecorator.Delete(r.Context(), entityCacheKey)
			if err != nil {
				m.logger.Debug("Failed to invalidate entity cache",
					logging.String("cache_key", entityCacheKey),
					logging.Error(err))
			}
		}
	}
}

// extractEntityType extracts entity type from URL path
func (m *HTTPCacheMiddleware) extractEntityType(path string) string {
	// Simple extraction - assumes entity type is the first segment after API version
	// In a real implementation, you might use a more sophisticated approach
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 {
		return parts[1] // Assuming /api/v1/entityType/...
	}
	return ""
}

// cacheResponseWriter captures the response for caching
type cacheResponseWriter struct {
	http.ResponseWriter
	cacheKey   string
	ttl        time.Duration
	middleware *HTTPCacheMiddleware
	body       []byte
	statusCode int
}

// Write captures the response body
func (w *cacheResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

// WriteHeader captures the status code
func (w *cacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Header returns the response headers
func (w *cacheResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// entityCacheResponseWriter captures the response for entity caching
type entityCacheResponseWriter struct {
	http.ResponseWriter
	entityType string
	entityID   string
	cacheKey   string
	ttl        time.Duration
	middleware *HTTPCacheMiddleware
	body       []byte
	statusCode int
}

// Write captures the response body
func (w *entityCacheResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

// WriteHeader captures the status code
func (w *entityCacheResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Header returns the response headers
func (w *entityCacheResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Close caches the response when the response is complete
func (w *entityCacheResponseWriter) Close() error {
	// If the response was successful, cache it
	if w.statusCode == 200 {
		w.middleware.logger.Debug("Caching entity response",
			logging.String("entity_type", w.entityType),
			logging.String("entity_id", w.entityID),
			logging.String("cache_key", w.cacheKey))

		err := w.middleware.cacheDecorator.Set(context.Background(), w.cacheKey, w.body, w.ttl)
		if err != nil {
			w.middleware.logger.Error("Failed to cache entity response",
				logging.String("cache_key", w.cacheKey),
				logging.Error(err))
		} else {
			w.middleware.logger.Debug("Entity response cached successfully",
				logging.String("entity_type", w.entityType),
				logging.String("cache_key", w.cacheKey))
		}
	}
	return nil
}

package middleware

import (
	"backend-core/logging"
	"backend-core/security"

	"github.com/gin-gonic/gin"
)

// JWTMiddlewareWrapper wraps the JWT authentication function for easy use
type JWTMiddlewareWrapper struct {
	jwtManager *security.JWTManager
	logger     *logging.Logger
}

// NewJWTMiddlewareWrapper creates a new JWT middleware wrapper
func NewJWTMiddlewareWrapper(jwtManager *security.JWTManager, logger *logging.Logger) *JWTMiddlewareWrapper {
	return &JWTMiddlewareWrapper{
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// RequireAuth returns the JWT authentication middleware
func (w *JWTMiddlewareWrapper) RequireAuth() gin.HandlerFunc {
	return JWTAuthMiddleware(w.jwtManager, w.logger)
}

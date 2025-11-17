package middleware

import (
	"auth-service/src/infrastructure/config"

	"backend-core/logging"
	errorHandling "backend-core/middleware/error_handling"
	httpMiddleware "backend-core/middleware/http"

	"github.com/gin-gonic/gin"
)

// MiddlewareSetup handles all middleware configuration
type MiddlewareSetup struct {
	config *config.Config
	logger *logging.Logger
}

// NewMiddlewareSetup creates a new middleware setup
func NewMiddlewareSetup(config *config.Config, logger *logging.Logger) *MiddlewareSetup {
	return &MiddlewareSetup{
		config: config,
		logger: logger,
	}
}

// SetupMiddleware configures all middleware for the application
func (ms *MiddlewareSetup) SetupMiddleware(router *gin.Engine) {
	// 1. Error handling middleware (comprehensive error handling)
	errorHandler := errorHandling.CreateErrorMiddleware(ms.logger)
	router.Use(errorHandler.Handler())      // Panic recovery
	router.Use(errorHandler.ErrorHandler()) // Service error handling

	// 2. Logging middleware for request/response logging
	router.Use(gin.Logger())

	// 3. Recovery middleware for panic recovery (backup)
	router.Use(gin.Recovery())

	// 4. Initialize HTTP middleware factory from backend-core
	httpMiddlewareFactory := httpMiddleware.NewHTTPMiddlewareFactory(ms.logger)

	// 5. CORS middleware from backend-core
	corsMiddleware := httpMiddlewareFactory.CreateDefaultCORSMiddleware()
	router.Use(corsMiddleware.Handler())

	// 6. Security Headers middleware from backend-core
	securityMiddleware := httpMiddlewareFactory.CreateDefaultSecurityHeadersMiddleware()
	router.Use(securityMiddleware.Handler())

	// 7. Request correlation middleware from backend-core
	requestCorrelationMiddleware := httpMiddlewareFactory.CreateDefaultRequestCorrelationMiddleware()
	router.Use(requestCorrelationMiddleware.Handler())

	ms.logger.Info("All middleware configured successfully")
}

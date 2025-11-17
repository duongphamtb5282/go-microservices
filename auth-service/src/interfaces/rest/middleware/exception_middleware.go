package middleware

import (
	"os"

	"auth-service/src/infrastructure/config"

	backendConfig "backend-core/config"
	"backend-core/middleware/exception"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// SetupExceptionMiddleware sets up exception handling middleware for the auth service
func SetupExceptionMiddleware(cfg *config.Config, logger *logging.Logger) []gin.HandlerFunc {
	// Determine environment
	env := getEnvironment()

	// Convert auth-service logging config to backend-core logging config
	backendLoggingConfig := &backendConfig.LoggingConfig{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	}

	// Use the new configuration-driven approach
	return exception.CreateMiddlewareFromLoggingConfig("auth-service", "1.0.0", env, backendLoggingConfig)
}

// getEnvironment determines the current environment
func getEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		return "development"
	}
	return env
}

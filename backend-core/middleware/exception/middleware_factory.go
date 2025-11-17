package exception

import (
	"backend-core/config"
	"backend-core/logging"
	"backend-core/middleware/error_handling"

	"github.com/gin-gonic/gin"
)

// MiddlewareConfig holds configuration for the exception middleware
type MiddlewareConfig struct {
	ServiceName  string
	Version      string
	IncludeStack bool
	HideInternal bool
	Logger       *logging.Logger
}

// EnvironmentConfig holds environment-specific configuration
type EnvironmentConfig struct {
	IncludeStack bool
	HideInternal bool
	LogLevel     string
}

// GetEnvironmentConfig returns configuration based on environment
func GetEnvironmentConfig(environment string) EnvironmentConfig {
	switch environment {
	case "production":
		return EnvironmentConfig{
			IncludeStack: false, // Don't include stack traces in production
			HideInternal: true,  // Hide internal error details
			LogLevel:     "info",
		}
	case "test":
		return EnvironmentConfig{
			IncludeStack: true,  // Include stack traces for debugging
			HideInternal: false, // Show all error details
			LogLevel:     "debug",
		}
	default: // development
		return EnvironmentConfig{
			IncludeStack: true,  // Include stack traces for debugging
			HideInternal: false, // Show internal error details
			LogLevel:     "debug",
		}
	}
}

// DefaultMiddlewareConfig returns default configuration
func DefaultMiddlewareConfig(serviceName, version string) *MiddlewareConfig {
	// Create a default logger using backend-core logging
	logger, _ := logging.NewLogger(&config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	})
	return &MiddlewareConfig{
		ServiceName:  serviceName,
		Version:      version,
		IncludeStack: true,
		HideInternal: true,
		Logger:       logger,
	}
}

// MiddlewareFactory creates exception handling middleware
type MiddlewareFactory struct {
	config *MiddlewareConfig
}

// NewMiddlewareFactory creates a new middleware factory
func NewMiddlewareFactory(config *MiddlewareConfig) *MiddlewareFactory {
	return &MiddlewareFactory{
		config: config,
	}
}

// CreateExceptionMiddleware creates the main exception handling middleware
func (f *MiddlewareFactory) CreateExceptionMiddleware() gin.HandlerFunc {
	errorMiddleware := error_handling.CreateErrorMiddleware(f.config.Logger)
	return errorMiddleware.Handler()
}

// CreateErrorHandler creates an error handler middleware
func (f *MiddlewareFactory) CreateErrorHandler() gin.HandlerFunc {
	errorMiddleware := error_handling.CreateErrorMiddleware(f.config.Logger)
	return errorMiddleware.ErrorHandler()
}

// CreateRecoveryMiddleware creates a recovery middleware
func (f *MiddlewareFactory) CreateRecoveryMiddleware() gin.HandlerFunc {
	handler := NewDefaultExceptionHandler(f.config.Logger, f.config.ServiceName, f.config.Version)
	handler.WithStackTrace(f.config.IncludeStack)
	handler.WithHideInternal(f.config.HideInternal)

	return gin.Recovery()
}

// CreateFullMiddlewareStack creates a complete middleware stack
func (f *MiddlewareFactory) CreateFullMiddlewareStack() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		f.CreateRecoveryMiddleware(),
		f.CreateExceptionMiddleware(),
		f.CreateErrorHandler(),
	}
}

// CreateMiddlewareFromConfig creates middleware based on configuration
func CreateMiddlewareFromConfig(serviceName, version, environment string, logger *logging.Logger) []gin.HandlerFunc {
	// Get environment-specific configuration
	envConfig := GetEnvironmentConfig(environment)

	// Create middleware configuration
	middlewareConfig := &MiddlewareConfig{
		ServiceName:  serviceName,
		Version:      version,
		IncludeStack: envConfig.IncludeStack,
		HideInternal: envConfig.HideInternal,
		Logger:       logger,
	}

	// Create factory and return middleware stack
	factory := NewMiddlewareFactory(middlewareConfig)
	return factory.CreateFullMiddlewareStack()
}

// CreateMiddlewareFromLoggingConfig creates middleware with custom logging configuration
func CreateMiddlewareFromLoggingConfig(serviceName, version, environment string, loggingConfig *config.LoggingConfig) []gin.HandlerFunc {
	// Get environment-specific configuration
	envConfig := GetEnvironmentConfig(environment)

	// Create logger from configuration
	logger, err := logging.NewLogger(loggingConfig)
	if err != nil {
		// Fallback to default logger
		logger, _ = logging.NewLogger(&config.LoggingConfig{
			Level:  envConfig.LogLevel,
			Format: "json",
			Output: "stdout",
		})
	}

	// Create middleware configuration
	middlewareConfig := &MiddlewareConfig{
		ServiceName:  serviceName,
		Version:      version,
		IncludeStack: envConfig.IncludeStack,
		HideInternal: envConfig.HideInternal,
		Logger:       logger,
	}

	// Create factory and return middleware stack
	factory := NewMiddlewareFactory(middlewareConfig)
	return factory.CreateFullMiddlewareStack()
}

// Convenience functions for backward compatibility (deprecated - use CreateMiddlewareFromConfig instead)

// CreateProductionMiddleware creates middleware optimized for production
// Deprecated: Use CreateMiddlewareFromConfig instead
func CreateProductionMiddleware(serviceName, version string, logger *logging.Logger) []gin.HandlerFunc {
	return CreateMiddlewareFromConfig(serviceName, version, "production", logger)
}

// CreateDevelopmentMiddleware creates middleware optimized for development
// Deprecated: Use CreateMiddlewareFromConfig instead
func CreateDevelopmentMiddleware(serviceName, version string, logger *logging.Logger) []gin.HandlerFunc {
	return CreateMiddlewareFromConfig(serviceName, version, "development", logger)
}

// CreateTestMiddleware creates middleware optimized for testing
// Deprecated: Use CreateMiddlewareFromConfig instead
func CreateTestMiddleware(serviceName, version string) []gin.HandlerFunc {
	// Create a default logger for testing
	logger, _ := logging.NewLogger(&config.LoggingConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})
	return CreateMiddlewareFromConfig(serviceName, version, "test", logger)
}

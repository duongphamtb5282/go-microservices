package applications

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"auth-service/src/applications/providers"
	"auth-service/src/applications/services"
	"auth-service/src/infrastructure/adapters"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/identity/keycloak"
	"auth-service/src/interfaces/rest/middleware"

	// localTelemetry "auth-service/src/infrastructure/telemetry" // Temporarily disabled
	"auth-service/src/infrastructure/worker"
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/database"
	"backend-core/logging"

	// backendTelemetry "backend-core/telemetry" // Temporarily disabled

	"github.com/gin-gonic/gin"
)

// ServiceFactory creates the complete auth service with proper dependency injection
type ServiceFactory struct {
	cfg        *config.Config
	db         database.Database
	dbAdapter  *adapters.DatabaseAdapter
	logger     *logging.Logger
	workerPool *worker.WorkerPool
}

// NewServiceFactory creates a new service factory
func NewServiceFactory(cfg *config.Config, db database.Database, dbAdapter *adapters.DatabaseAdapter, logger *logging.Logger) *ServiceFactory {
	// Create worker pool for async task processing
	workerPool := providers.WorkerPoolProvider(logger)

	// Temporarily disable telemetry for testing login issue
	logger.Info("Temporarily disabling telemetry for testing")

	return &ServiceFactory{
		cfg:        cfg,
		db:         db,
		dbAdapter:  dbAdapter,
		logger:     logger,
		workerPool: workerPool,
	}
}

// CreateRouter creates the complete router with all dependencies
func (f *ServiceFactory) CreateRouter(brokers []string) *gin.Engine {
	fmt.Printf("SERVICE_FACTORY: CreateRouter called\n")
	// Temporarily disable telemetry for testing
	f.logger.Info("Temporarily disabling telemetry for testing")
	// var backendCoreTelemetry *backendTelemetry.Telemetry
	// var businessMetrics *backendTelemetry.BusinessMetrics

	// Create security managers from backend-core/security
	jwtManager := providers.JWTManagerProvider(f.cfg, f.logger)
	// AuthManager available for password hashing and verification in handlers
	_ = providers.AuthManagerProvider(jwtManager, f.cfg, f.logger)

	// Create repositories
	userRepo := providers.UserRepositoryProvider(f.db, f.logger)
	userCache := providers.UserCacheProvider(f.db, f.logger)

	// Create domain services
	userDomainService := providers.UserDomainServiceProvider(userRepo, f.logger)

	// Create event bus
	eventBus := providers.EventBusProvider(brokers, f.logger)

	// Create application services with telemetry
	// TODO: Initialize AdminClient for gRPC communication with admin-service
	var adminClient services.AdminClient = nil // Optional for now

	// Temporarily disable telemetry providers
	f.logger.Info("Temporarily disabling telemetry providers for testing")
	// var providerTelemetry *localTelemetry.SimpleTelemetry
	// var providerBusinessMetrics *localTelemetry.BusinessMetrics
	// telemetryConfig := localTelemetry.TelemetryConfigFromEnv()
	// providerTelemetry = localTelemetry.NewSimpleTelemetry(telemetryConfig, f.logger)
	// providerBusinessMetrics = &localTelemetry.BusinessMetrics{}

	// Create and register auth task handler
	authTaskHandler := providers.AuthTaskHandlerProvider(userCache, eventBus, adminClient, f.logger)
	f.workerPool.RegisterHandler(authTaskHandler)

	userApplicationService := providers.UserApplicationServiceProvider(
		userRepo,
		userCache,
		userDomainService,
		eventBus,
		adminClient,
		f.workerPool, // Add worker pool
		f.logger,
	)

	// Create authorization repositories (commented out for now as they're not used in simplified router)
	// roleRepo := providers.RoleRepositoryProvider(f.db, f.logger)
	// permissionRepo := providers.PermissionRepositoryProvider(f.db, f.logger)

	// Create cache for Keycloak (using a simple in-memory cache for now)
	// TODO: Use proper cache from backend-core
	// keycloakAdapter, err := providers.KeycloakAdapterProvider(f.cfg, nil, f.logger)
	// if err != nil {
	//	f.logger.Warn("Failed to create Keycloak adapter, authorization may not work", "error", err)
	//	keycloakAdapter = nil
	// }
	// keycloakApplicationService := providers.KeycloakApplicationServiceProvider(keycloakAdapter, f.logger)

	// Create handlers
	userHandler := providers.UserHandlerProvider(userApplicationService, f.logger)

	// Create identity provider
	identityProvider := f.createIdentityProvider(userApplicationService)

	// Create auth handler
	authHandler := providers.AuthHandlerProvider(userApplicationService, jwtManager, f.logger, identityProvider, &f.cfg.Authorization)

	// Create cache middleware
	cacheMiddleware := providers.CacheMiddlewareProvider(f.cfg, f.logger)

	// Create Keycloak authorization middleware
	var keycloakAdapter *keycloak.KeycloakAdapter
	var keycloakAuth *middleware.KeycloakAuthorizationMiddleware

	// Create Keycloak adapter if configured
	f.logger.Info("Checking authorization config", "mode", f.cfg.Authorization.Mode, "identity_provider", f.cfg.Authorization.IdentityProvider)
	if f.cfg.Authorization.IdentityProvider == config.IdentityProviderKeycloak {
		f.logger.Info("Creating Keycloak adapter for authorization")
		var err error
		keycloakAdapter, err = providers.KeycloakAdapterProvider(f.cfg, nil, f.logger)
		if err != nil {
			f.logger.Warn("Failed to create Keycloak adapter, authorization may not work", "error", err)
			keycloakAdapter = nil
		} else {
			f.logger.Info("Keycloak adapter created successfully")
			// Create Keycloak authorization middleware
			keycloakAuth = providers.KeycloakAuthorizationMiddlewareProvider(keycloakAdapter, f.logger)
			f.logger.Info("Keycloak authorization middleware created successfully")
		}
	} else {
		f.logger.Info("Keycloak not configured, identity provider mode:", "mode", f.cfg.Authorization.IdentityProvider)
	}

	// Temporarily disable telemetry middleware for testing
	f.logger.Info("Temporarily disabling telemetry middleware for testing")
	var telemetryMiddleware gin.HandlerFunc = nil
	// telemetryMiddleware = f.createPrometheusMetricsMiddleware()

	// Create route manager (includes Swagger support)
	f.logger.Info("Creating route manager")
	routeManager := providers.RouteManagerProvider(authHandler, userHandler, cacheMiddleware, keycloakAuth, f.logger, telemetryMiddleware)

	// Setup routes and middleware
	f.logger.Info("Setting up routes")
	ginRouter := routeManager.SetupRoutes()
	f.logger.Info("Routes setup completed")

	// Setup middleware
	// middlewareSetup.SetupMiddleware(ginRouter)

	// Temporarily disable metrics endpoint
	// if f.metricsService != nil {
	// 	f.logger.Info("Adding /metrics endpoint for Prometheus")
	// 	ginRouter.GET("/metrics", f.metricsService.MetricsHandler())
	// }

	// Note: Health check endpoint is already registered in router.go
	// No need to add it again here to avoid duplicate route panic

	return ginRouter
}

// Shutdown gracefully shuts down all components
func (f *ServiceFactory) Shutdown(ctx context.Context) error {
	f.logger.Info("Shutting down service factory components")

	// Temporarily disable metrics service shutdown
	// if f.metricsService != nil {
	// 	if err := f.metricsService.Shutdown(ctx); err != nil {
	// 		f.logger.Error("Failed to shutdown metrics service", "error", err)
	// 		// Don't return error, continue with other shutdowns
	// 	} else {
	// 		f.logger.Info("Metrics service shutdown completed")
	// 	}
	// }

	// Shutdown worker pool
	if f.workerPool != nil {
		if err := f.workerPool.Shutdown(ctx); err != nil {
			f.logger.Error("Failed to shutdown worker pool", "error", err)
			return fmt.Errorf("failed to shutdown worker pool: %w", err)
		}
		f.logger.Info("Worker pool shutdown completed")
	}

	f.logger.Info("All service factory components shutdown completed")
	return nil
}

// addHealthCheckEndpoint adds a health check endpoint that uses the database adapter
func (f *ServiceFactory) addHealthCheckEndpoint(router *gin.Engine) {
	router.GET("/api/v1/health", func(c *gin.Context) {
		// Basic health check
		response := gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "auth-service",
		}

		// Add worker pool health check
		if f.workerPool != nil {
			stats := f.workerPool.GetStats()
			response["worker_pool"] = gin.H{
				"status":          "healthy",
				"active_workers":  stats["active_workers"],
				"queue_size":      stats["queue_size"],
				"tasks_processed": stats["tasks_processed"],
			}
		}

		// Add database health check if adapter is available
		if f.dbAdapter != nil {
			healthCheck := f.dbAdapter.HealthCheck()
			response["database"] = gin.H{
				"status":    healthCheck.GetStatus(),
				"timestamp": healthCheck.GetTimestamp().Format(time.RFC3339),
			}
			if healthCheck.GetError() != nil {
				response["database"].(gin.H)["error"] = healthCheck.GetError().Error()
			}
		} else {
			response["database"] = gin.H{
				"status": "not_configured",
			}
		}

		c.JSON(http.StatusOK, response)
	})
}

// createPrometheusMetricsMiddleware creates HTTP middleware that records metrics directly to Prometheus
func (f *ServiceFactory) createPrometheusMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Prometheus middleware triggered for %s %s\n", c.Request.Method, c.Request.URL.Path)
		// start := time.Now() // Temporarily disabled

		// Process request
		c.Next()

		// Temporarily disable metrics recording
		// duration := time.Since(start).Seconds()
		// status := fmt.Sprintf("%d", c.Writer.Status())
		// method := c.Request.Method
		// path := c.FullPath()
		// if path == "" {
		// 	path = c.Request.URL.Path
		// }

		// Temporarily disable HTTP metrics recording
		// if f.metricsService != nil {
		// 	fmt.Printf("DEBUG: Recording metrics for %s %s -> %s (%.3fs)\n", method, path, status, duration)
		// 	f.metricsService.RecordHTTPMetrics(method, path, status, duration, c.Writer.Status())
		// } else {
		// 	fmt.Printf("DEBUG: Metrics service is nil\n")
		// }
	}
}

// createIdentityProvider creates the appropriate identity provider based on configuration
func (f *ServiceFactory) createIdentityProvider(userService *services.UserApplicationService) handlers.IdentityProvider {
	switch f.cfg.Authorization.IdentityProvider {
	case config.IdentityProviderDatabase:
		return handlers.NewDatabaseIdentityProvider(userService, f.logger)
	case config.IdentityProviderKeycloak:
		// TODO: Create Keycloak identity provider
		f.logger.Warn("Keycloak identity provider not yet implemented, falling back to database")
		return handlers.NewDatabaseIdentityProvider(userService, f.logger)
	case config.IdentityProviderPingAM:
		// TODO: Create PingAM identity provider
		f.logger.Warn("PingAM identity provider not yet implemented, falling back to database")
		return handlers.NewDatabaseIdentityProvider(userService, f.logger)
	default:
		f.logger.Warn("Unknown identity provider, using database as default",
			"provider", string(f.cfg.Authorization.IdentityProvider))
		return handlers.NewDatabaseIdentityProvider(userService, f.logger)
	}
}

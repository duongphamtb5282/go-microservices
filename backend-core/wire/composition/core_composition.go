package composition

import (
	"backend-core/cache/decorators"
	"backend-core/config"
	"backend-core/database"
	"backend-core/logging"
	"backend-core/monitoring"
	"backend-core/security"
	"backend-core/telemetry"
	"backend-core/wire/providers"
	"net/http"
)

// CoreComposition composes the core infrastructure
type CoreComposition struct {
	// Core Infrastructure
	Config            *config.Config
	Logger            *logging.Logger
	Database          database.Database
	CacheDecorator    *decorators.CacheDecorator
	SecurityManager   *security.SecurityManager
	MonitoringManager *monitoring.MonitoringManager
	Telemetry         telemetry.TelemetryInterface
	BusinessMetrics   *telemetry.BusinessMetrics

	// Factories
	RepositoryFactory *providers.RepositoryFactory
	ServiceFactory    *providers.ServiceFactory
	HandlerFactory    *providers.HandlerFactory
	MiddlewareFactory *providers.MiddlewareFactory

	// Application
	BaseApplication *providers.BaseApplication
}

// ComposeCoreInfrastructure creates the core infrastructure composition
func ComposeCoreInfrastructure() *CoreComposition {
	// 1. Create configuration
	config := providers.ConfigProvider()

	// 2. Create logger
	logger := providers.LoggerProvider(config)

	// 3. Create database
	database := providers.DatabaseProvider(config)

	// 4. Create cache
	cacheFactory := providers.CacheDecoratorFactoryProvider()
	cacheDecorator := providers.CacheDecoratorProvider(cacheFactory)

	// 5. Create security manager
	securityManager := providers.SecurityManagerProvider()

	// 6. Create monitoring manager
	monitoringManager := providers.MonitoringManagerProvider()

	// 7. Create telemetry
	telemetryConfig := telemetry.TelemetryConfigFromEnv()
	telemetryInstance := telemetry.NewSimpleTelemetry(telemetryConfig, logger)

	// 8. Create business metrics
	businessMetrics := &telemetry.BusinessMetrics{}

	// 9. Create factories
	repositoryFactory := providers.RepositoryFactoryProvider(database, cacheDecorator)
	serviceFactory := providers.ServiceFactoryProvider(repositoryFactory, logger)
	handlerFactory := providers.HandlerFactoryProvider(serviceFactory, logger)
	middlewareFactory := providers.MiddlewareFactoryProvider(cacheDecorator, securityManager, monitoringManager, logger)

	// 10. Create base application (with placeholder HTTP server and router)
	baseApplication := providers.BaseApplicationProvider(
		logger,
		database,
		cacheDecorator,
		securityManager,
		monitoringManager,
		repositoryFactory,
		serviceFactory,
		handlerFactory,
		middlewareFactory,
		&http.Server{}, // Placeholder
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), // Placeholder
	)

	return &CoreComposition{
		Config:            config,
		Logger:            logger,
		Database:          database,
		CacheDecorator:    cacheDecorator,
		SecurityManager:   securityManager,
		MonitoringManager: monitoringManager,
		Telemetry:         telemetryInstance,
		BusinessMetrics:   businessMetrics,
		RepositoryFactory: repositoryFactory,
		ServiceFactory:    serviceFactory,
		HandlerFactory:    handlerFactory,
		MiddlewareFactory: middlewareFactory,
		BaseApplication:   baseApplication,
	}
}

// GetLogger returns the logger from the composition
func (c *CoreComposition) GetLogger() *logging.Logger {
	return c.Logger
}

// GetDatabase returns the database from the composition
func (c *CoreComposition) GetDatabase() database.Database {
	return c.Database
}

// GetCacheDecorator returns the cache decorator from the composition
func (c *CoreComposition) GetCacheDecorator() *decorators.CacheDecorator {
	return c.CacheDecorator
}

// GetSecurityManager returns the security manager from the composition
func (c *CoreComposition) GetSecurityManager() *security.SecurityManager {
	return c.SecurityManager
}

// GetMonitoringManager returns the monitoring manager from the composition
func (c *CoreComposition) GetMonitoringManager() *monitoring.MonitoringManager {
	return c.MonitoringManager
}

// GetBaseApplication returns the base application from the composition
func (c *CoreComposition) GetBaseApplication() *providers.BaseApplication {
	return c.BaseApplication
}

// GetTelemetry returns the telemetry from the composition
func (c *CoreComposition) GetTelemetry() telemetry.TelemetryInterface {
	return c.Telemetry
}

// GetBusinessMetrics returns the business metrics from the composition
func (c *CoreComposition) GetBusinessMetrics() *telemetry.BusinessMetrics {
	return c.BusinessMetrics
}

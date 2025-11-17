package providers

import (
	"backend-core/cache/decorators"
	"backend-core/database"
	"backend-core/logging"
	"backend-core/monitoring"
	"backend-core/security"
	"context"
	"net/http"
)

// BaseApplicationProvider creates a base application
func BaseApplicationProvider(
	logger *logging.Logger,
	database database.Database,
	cacheDecorator *decorators.CacheDecorator,
	securityManager *security.SecurityManager,
	monitoringManager *monitoring.MonitoringManager,
	repositoryFactory *RepositoryFactory,
	serviceFactory *ServiceFactory,
	handlerFactory *HandlerFactory,
	middlewareFactory *MiddlewareFactory,
	httpServer *http.Server,
	router http.Handler,
) *BaseApplication {
	return &BaseApplication{
		logger:            logger,
		database:          database,
		cacheDecorator:    cacheDecorator,
		securityManager:   securityManager,
		monitoringManager: monitoringManager,
		repositoryFactory: repositoryFactory,
		serviceFactory:    serviceFactory,
		handlerFactory:    handlerFactory,
		middlewareFactory: middlewareFactory,
		httpServer:        httpServer,
		router:            router,
	}
}

// BaseApplication provides common application functionality
type BaseApplication struct {
	logger            *logging.Logger
	database          database.Database
	cacheDecorator    *decorators.CacheDecorator
	securityManager   *security.SecurityManager
	monitoringManager *monitoring.MonitoringManager
	repositoryFactory *RepositoryFactory
	serviceFactory    *ServiceFactory
	handlerFactory    *HandlerFactory
	middlewareFactory *MiddlewareFactory
	httpServer        *http.Server
	router            http.Handler
}

// GetLogger returns the logger
func (a *BaseApplication) GetLogger() *logging.Logger {
	return a.logger
}

// GetDatabase returns the database
func (a *BaseApplication) GetDatabase() database.Database {
	return a.database
}

// GetCacheDecorator returns the cache decorator
func (a *BaseApplication) GetCacheDecorator() *decorators.CacheDecorator {
	return a.cacheDecorator
}

// GetSecurityManager returns the security manager
func (a *BaseApplication) GetSecurityManager() *security.SecurityManager {
	return a.securityManager
}

// GetMonitoringManager returns the monitoring manager
func (a *BaseApplication) GetMonitoringManager() *monitoring.MonitoringManager {
	return a.monitoringManager
}

// GetHTTPServer returns the HTTP server
func (a *BaseApplication) GetHTTPServer() *http.Server {
	return a.httpServer
}

// GetRouter returns the router
func (a *BaseApplication) GetRouter() http.Handler {
	return a.router
}

// Start starts the application
func (a *BaseApplication) Start(ctx context.Context) error {
	a.logger.Info("Starting application")
	return a.httpServer.ListenAndServe()
}

// Stop stops the application
func (a *BaseApplication) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application")
	return a.httpServer.Shutdown(ctx)
}

// HealthCheck performs a health check
func (a *BaseApplication) HealthCheck() error {
	ctx := context.Background()

	// Check database
	if err := a.database.Ping(ctx); err != nil {
		return err
	}

	// Check cache
	if err := a.cacheDecorator.HealthCheck(ctx); err != nil {
		return err
	}

	return nil
}

// GetMetrics returns application metrics
func (a *BaseApplication) GetMetrics() map[string]interface{} {
	ctx := context.Background()
	cacheStats, _ := a.cacheDecorator.GetCacheStats(ctx)

	return map[string]interface{}{
		"database":   a.database.GetStats(),
		"cache":      cacheStats,
		"security":   a.securityManager.GetStats(),
		"monitoring": a.monitoringManager.GetStats(),
	}
}

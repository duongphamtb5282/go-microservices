package wire

import (
	"backend-core/cache/decorators"
	"backend-core/config"
	"backend-core/database"
	"backend-core/logging"
	"backend-core/monitoring"
	"backend-core/security"
	"backend-core/wire/providers"
)

// GenericLoggerProvider creates a logger for any service
func GenericLoggerProvider(cfg *config.Config, serviceName string) (*logging.Logger, error) {
	return logging.NewLogger(&cfg.Logging)
}

// GenericDatabaseProvider creates a database connection for any service
func GenericDatabaseProvider(cfg *config.Config, logger *logging.Logger, serviceName string) (database.Database, error) {
	// Simple implementation - would need proper database factory
	return nil, nil
}

// GenericCacheProvider creates a cache decorator for any service
func GenericCacheProvider(cfg *config.Config, logger *logging.Logger, serviceName string, strategy string) (*decorators.CacheDecorator, error) {
	// Simple implementation - would need proper cache factory
	return &decorators.CacheDecorator{}, nil
}

// GenericSecurityProvider creates a security manager for any service
func GenericSecurityProvider(cfg *config.Config, logger *logging.Logger, serviceName string) (*security.SecurityManager, error) {
	return security.NewSecurityManager(), nil
}

// GenericMonitoringProvider creates monitoring components for any service
func GenericMonitoringProvider(cfg *config.Config, logger *logging.Logger, serviceName string) (*monitoring.MonitoringManager, error) {
	return monitoring.NewMonitoringManager(), nil
}

// GenericApplicationProvider creates a generic application for any service
func GenericApplicationProvider(
	logger *logging.Logger,
	database database.Database,
	cacheDecorator *decorators.CacheDecorator,
	securityManager *security.SecurityManager,
	monitoringManager *monitoring.MonitoringManager,
	repositoryFactory *providers.RepositoryFactory,
	serviceFactory *providers.ServiceFactory,
	handlerFactory *providers.HandlerFactory,
	middlewareFactory *providers.MiddlewareFactory,
) *providers.BaseApplication {
	return providers.BaseApplicationProvider(
		logger,
		database,
		cacheDecorator,
		securityManager,
		monitoringManager,
		repositoryFactory,
		serviceFactory,
		handlerFactory,
		middlewareFactory,
		nil, // httpServer placeholder
		nil, // router placeholder
	)
}

package providers

import (
	"backend-core/cache/decorators"
	"backend-core/database"
	"backend-core/logging"
	"backend-core/monitoring"
	"backend-core/security"
)

// RepositoryFactoryProvider creates a repository factory
func RepositoryFactoryProvider(
	database database.Database,
	cacheDecorator *decorators.CacheDecorator,
) *RepositoryFactory {
	return &RepositoryFactory{
		database:       database,
		cacheDecorator: cacheDecorator,
	}
}

// ServiceFactoryProvider creates a service factory
func ServiceFactoryProvider(
	repositoryFactory *RepositoryFactory,
	logger *logging.Logger,
) *ServiceFactory {
	return &ServiceFactory{
		repositoryFactory: repositoryFactory,
		logger:            logger,
	}
}

// HandlerFactoryProvider creates a handler factory
func HandlerFactoryProvider(
	serviceFactory *ServiceFactory,
	logger *logging.Logger,
) *HandlerFactory {
	return &HandlerFactory{
		serviceFactory: serviceFactory,
		logger:         logger,
	}
}

// MiddlewareFactoryProvider creates a middleware factory
func MiddlewareFactoryProvider(
	cacheDecorator *decorators.CacheDecorator,
	securityManager *security.SecurityManager,
	monitoringManager *monitoring.MonitoringManager,
	logger *logging.Logger,
) *MiddlewareFactory {
	return &MiddlewareFactory{
		cacheDecorator:    cacheDecorator,
		securityManager:   securityManager,
		monitoringManager: monitoringManager,
		logger:            logger,
	}
}

// Factory structs
type RepositoryFactory struct {
	database       database.Database
	cacheDecorator *decorators.CacheDecorator
}

type ServiceFactory struct {
	repositoryFactory *RepositoryFactory
	logger            *logging.Logger
}

type HandlerFactory struct {
	serviceFactory *ServiceFactory
	logger         *logging.Logger
}

type MiddlewareFactory struct {
	cacheDecorator    *decorators.CacheDecorator
	securityManager   *security.SecurityManager
	monitoringManager *monitoring.MonitoringManager
	logger            *logging.Logger
}

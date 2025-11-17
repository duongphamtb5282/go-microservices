package wire

import (
	"backend-core/wire/providers"

	"github.com/google/wire"
)

// ServiceProviderFactory creates providers for specific services
type ServiceProviderFactory struct {
	serviceName string
	strategy    string
}

// NewServiceProviderFactory creates a new service provider factory
func NewServiceProviderFactory(serviceName, strategy string) *ServiceProviderFactory {
	return &ServiceProviderFactory{
		serviceName: serviceName,
		strategy:    strategy,
	}
}

// CreateUserServiceProviders creates providers for user service
func (f *ServiceProviderFactory) CreateUserServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

// CreatePaymentServiceProviders creates providers for payment service
func (f *ServiceProviderFactory) CreatePaymentServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

// CreateNotificationServiceProviders creates providers for notification service
func (f *ServiceProviderFactory) CreateNotificationServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

// CreateOrderServiceProviders creates providers for order service
func (f *ServiceProviderFactory) CreateOrderServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

// CreateInventoryServiceProviders creates providers for inventory service
func (f *ServiceProviderFactory) CreateInventoryServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

// CreateGenericServiceProviders creates providers for any service
func (f *ServiceProviderFactory) CreateGenericServiceProviders() wire.ProviderSet {
	return wire.NewSet(
		// Core infrastructure providers
		providers.ConfigProvider,
		providers.LoggerProvider,
		providers.DatabaseProvider,
		providers.CacheDecoratorFactoryProvider,
		providers.CacheDecoratorProvider,
		providers.SecurityManagerProvider,
		providers.MonitoringManagerProvider,

		// Factory providers
		providers.RepositoryFactoryProvider,
		providers.ServiceFactoryProvider,
		providers.HandlerFactoryProvider,
		providers.MiddlewareFactoryProvider,

		// Application providers
		providers.BaseApplicationProvider,
	)
}

//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
)

// InitializeBaseApplication creates a base application with core dependencies
func InitializeBaseApplication(serviceName, strategy string) (Application, error) {
	wire.Build(
		BaseProviderSet,
		NewBaseApplication,
	)
	return &BaseApplication{}, nil
}

// InitializeServiceApplication creates a service application with all dependencies
func InitializeServiceApplication(serviceName, strategy string) (Application, error) {
	wire.Build(
		ApplicationProviderSet,
		NewServiceApplication,
	)
	return &ServiceApplication{}, nil
}

// InitializeAuthService creates an auth service application
func InitializeAuthService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("auth-service", "balanced").CreateAuthServiceProviders(),
	)
	return &AuthApplication{}, nil
}

// InitializeUserService creates a user service application
func InitializeUserService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("user-service", "balanced").CreateUserServiceProviders(),
	)
	return &UserApplication{}, nil
}

// InitializePaymentService creates a payment service application
func InitializePaymentService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("payment-service", "high-performance").CreatePaymentServiceProviders(),
	)
	return &PaymentApplication{}, nil
}

// InitializeNotificationService creates a notification service application
func InitializeNotificationService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("notification-service", "balanced").CreateNotificationServiceProviders(),
	)
	return &NotificationApplication{}, nil
}

// InitializeOrderService creates an order service application
func InitializeOrderService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("order-service", "high-performance").CreateOrderServiceProviders(),
	)
	return &OrderApplication{}, nil
}

// InitializeInventoryService creates an inventory service application
func InitializeInventoryService() (Application, error) {
	wire.Build(
		NewServiceProviderFactory("inventory-service", "balanced").CreateInventoryServiceProviders(),
	)
	return &InventoryApplication{}, nil
}

// InitializeGenericService creates a generic service application
func InitializeGenericService(serviceName, strategy string) (Application, error) {
	wire.Build(
		NewServiceProviderFactory(serviceName, strategy).CreateGenericServiceProviders(),
	)
	return &GenericApplication{}, nil
}

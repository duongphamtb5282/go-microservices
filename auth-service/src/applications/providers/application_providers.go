package providers

import (
	"time"

	"auth-service/src/applications/services"
	"auth-service/src/domain/repositories"
	domainServices "auth-service/src/domain/services"
	"auth-service/src/infrastructure/messaging/kafka"

	// "auth-service/src/infrastructure/telemetry" // Temporarily disabled
	"auth-service/src/infrastructure/worker"
	"backend-core/logging"
)

// EventBusProvider creates an event bus
func EventBusProvider(brokers []string, logger *logging.Logger) services.EventBus {
	return kafka.NewKafkaEventBus(brokers, logger)
}

// WorkerPoolProvider creates a worker pool for async task processing
func WorkerPoolProvider(logger *logging.Logger) *worker.WorkerPool {
	config := &worker.WorkerPoolConfig{
		Name:           "auth-service-worker-pool",
		NumWorkers:     10, // Start with 10 workers, can be configured
		QueueSize:      1000,
		MaxRetries:     3,
		DefaultTimeout: 30 * time.Second,
	}
	return worker.NewWorkerPool(config, logger)
}

// AuthTaskHandlerProvider creates an auth task handler
func AuthTaskHandlerProvider(
	userCache repositories.UserCache,
	eventBus services.EventBus,
	adminClient services.AdminClient,
	logger *logging.Logger,
) *worker.AuthTaskHandler {
	return worker.NewAuthTaskHandler(userCache, eventBus, adminClient, logger)
}

// UserApplicationServiceProvider creates a user application service
func UserApplicationServiceProvider(
	userRepo repositories.UserRepository,
	userCache repositories.UserCache,
	userDomainService *domainServices.UserDomainService,
	eventBus services.EventBus,
	adminClient services.AdminClient,
	workerPool *worker.WorkerPool,
	logger *logging.Logger,
	// telemetry *telemetry.SimpleTelemetry, // Temporarily disabled
	// businessMetrics *telemetry.BusinessMetrics, // Temporarily disabled
) *services.UserApplicationService {
	return services.NewUserApplicationService(
		userRepo,
		userCache,
		userDomainService,
		eventBus,
		adminClient,
		workerPool,
		logger,
		// telemetry, // Temporarily disabled
		// businessMetrics, // Temporarily disabled
	)
}

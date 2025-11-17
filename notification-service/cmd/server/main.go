package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notification-service/internal/application/notification"
	"notification-service/internal/config"
	"notification-service/internal/infrastructure/events"

	"backend-core/logging"
	"backend-core/wire"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize core infrastructure
	coreComposition := wire.InitializeCoreInfrastructure()
	logger := coreComposition.GetLogger()
	defer logger.Sync()

	logger.Info("starting notification-service", "version", "1.0.0", "port", cfg.Server.Port)

	// Create notification handler
	notificationHandler := notification.NewNotificationHandler(logger)

	// Create Kafka consumer using backend-core
	consumer, err := events.NewKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, []string{cfg.Kafka.Topics.UserEvents}, logger)
	if err != nil {
		logger.Fatal("Failed to create Kafka consumer", logging.Error(err))
	}
	defer consumer.Close()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming messages
	go func() {
		if err := consumer.ConsumeMessages(ctx, []string{cfg.Kafka.Topics.UserEvents}, notificationHandler); err != nil {
			logger.Error("failed to consume messages", "error", err)
		}
	}()

	logger.Info("notification service started successfully",
		"topic", cfg.Kafka.Topics.UserEvents,
		"group_id", cfg.Kafka.GroupID,
		"brokers", cfg.Kafka.Brokers)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down notification service...")
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logger.Info("Notification service stopped")
}

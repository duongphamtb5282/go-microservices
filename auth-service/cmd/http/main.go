package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/src/applications"
	"auth-service/src/infrastructure/adapters"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/utils"
	backendCoreConfig "backend-core/config"
	"backend-core/database"
	"backend-core/database/gorm"
	"backend-core/logging"

	// "backend-core/telemetry" // Temporarily disabled

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	fileEnabled := os.Getenv("LOG_FILE_ENABLED") == "true"
	filePath := os.Getenv("LOG_FILE_PATH")
	if filePath == "" {
		filePath = "/var/log/auth-service.log"
	}

	var logConfig backendCoreConfig.LoggingConfig
	if fileEnabled {
		logConfig = backendCoreConfig.LoggingConfig{
			Level:      cfg.Logging.Level,
			Format:     cfg.Logging.Format,
			Output:     "file",
			FilePath:   filePath,
			MaxSize:    100,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}
	} else {
		logConfig = backendCoreConfig.LoggingConfig{
			Level:  cfg.Logging.Level,
			Format: cfg.Logging.Format,
			Output: cfg.Logging.Output,
		}
	}

	logger, err := logging.NewLogger(&logConfig)
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	// Temporarily disable telemetry for testing login issue
	logger.Info("Temporarily disabling telemetry for testing")
	// telemetryConfig := telemetry.TelemetryConfig{
	// 	ServiceName:    "auth-service",
	// 	ServiceVersion: "1.0.0",
	// 	Environment:    getEnv("APP_ENV", "development"),
	// 	OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
	// 	Enabled:        getEnvBool("OTEL_ENABLED", true),
	// }

	// tel, err := telemetry.NewTelemetry(telemetryConfig)
	// if err != nil {
	// 	logger.Warn("Failed to initialize telemetry, continuing without tracing", "error", err)
	// } else {
	// 	logger.Info("Telemetry initialized successfully",
	// 		"service", telemetryConfig.ServiceName,
	// 		"endpoint", telemetryConfig.OTLPEndpoint)
	// 	defer func() {
	// 		if err := tel.Shutdown(context.Background()); err != nil {
	// 			logger.Error("Failed to shutdown telemetry", "error", err)
	// 		}
	// 	}()
	// }

	// Initialize database using backend-core with auth-service adapter
	var db database.Database
	var dbAdapter *adapters.DatabaseAdapter
	db, dbAdapter, err = initializeDatabaseWithAdapter(cfg.Database, logger)
	if err != nil {
		logger.Warn("Failed to initialize database, using in-memory storage",
			"error", err,
			"host", cfg.Database.Host,
			"port", cfg.Database.Port,
			"username", cfg.Database.Username,
			"database", cfg.Database.Name)
		db = nil // Use nil to indicate no database
		dbAdapter = nil
	}

	// Initialize the complete auth service using service factory
	serviceFactory := applications.NewServiceFactory(cfg, db, dbAdapter, logger)
	ginRouter := serviceFactory.CreateRouter(cfg.Kafka.Brokers)

	// Ensure graceful shutdown of service factory
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := serviceFactory.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown service factory", "error", err)
		}
	}()

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: ginRouter, // Modified
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", "error", err)
	}

	logger.Info("Server exited")
}

// initializeDatabaseWithAdapter initializes the database connection using backend-core with auth-service adapter
func initializeDatabaseWithAdapter(cfg config.DatabaseConfig, logger *logging.Logger) (database.Database, *adapters.DatabaseAdapter, error) {
	// Convert auth-service config to backend-core config
	dbConfig := &backendCoreConfig.DatabaseConfig{
		Type:                       cfg.Type,
		Host:                       cfg.Host,
		Port:                       cfg.Port,
		Username:                   cfg.Username,
		Password:                   cfg.Password,
		Database:                   cfg.Name,
		SSLMode:                    cfg.SSLMode,
		MaxOpenConns:               cfg.MaxOpenConns,
		MaxIdleConns:               cfg.MaxIdleConns,
		ConnMaxLifetime:            utils.ParseDuration(cfg.ConnMaxLifetime),
		ConnMaxIdleTime:            utils.ParseDuration(cfg.ConnMaxIdleTime),
		QueryTimeout:               utils.ParseDuration(cfg.QueryTimeout),
		AcquireTimeout:             utils.ParseDuration(cfg.AcquireTimeout),
		AcquireRetryAttempts:       cfg.AcquireRetryAttempts,
		PreparedStatementCacheSize: cfg.PreparedStatementCacheSize,
		LogLevel:                   cfg.LogLevel,
		SlowQueryThreshold:         utils.ParseDuration(cfg.SlowThreshold),
	}

	// Create database using backend-core factory
	factory := database.NewDatabaseFactory()
	db, err := factory.CreatePostgreSQLDatabase(dbConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Connect to the database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	logger.Info("✅ Database connection established")

	// Create auth-service specific database adapter
	// Note: We need to cast the database to the correct type for the adapter
	var dbAdapter *adapters.DatabaseAdapter
	if gormDB, ok := db.(gorm.Database); ok {
		dbAdapter = adapters.NewDatabaseAdapter(gormDB, logger)

		// Initialize database with auth-service specific configuration
		ctx := context.Background()
		if err := dbAdapter.InitializeDatabase(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to initialize database with adapter: %w", err)
		}
	} else {
		// Fallback: create adapter without specific GORM database
		dbAdapter = nil
	}

	logger.Info("Database connected successfully using backend-core with auth-service adapter")
	return db, dbAdapter, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool retrieves a boolean environment variable or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes":
			return true
		case "false", "0", "no":
			return false
		}
	}
	return defaultValue
}

package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"admin-service/src/applications/services"
	adminConfig "admin-service/src/infrastructure/config"
	"admin-service/src/infrastructure/persistence"
	grpcServer "admin-service/src/interfaces/grpc"
	"admin-service/src/interfaces/websocket"
	backendCoreConfig "backend-core/config"
	"backend-core/database"
	"backend-core/database/postgresql"
	grpcauth "backend-core/grpc/interceptors/auth"
	grpclogging "backend-core/grpc/interceptors/logging"
	grpcserver "backend-core/grpc/server"
	"backend-core/logging"
	"backend-core/telemetry"
	pb "backend-shared/proto/admin"
	"backend-shared/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load configuration
	cfg, err := adminConfig.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger - convert admin config to backend-core config
	logConfig := &backendCoreConfig.LoggingConfig{
		Level:      cfg.Logging.Level,
		Format:     cfg.Logging.Format,
		Output:     cfg.Logging.Output,
		FilePath:   "/var/log/admin-service.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	logger, err := logging.NewLogger(logConfig)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Starting admin-service", "version", "1.0.0", "http_port", cfg.Server.Port, "grpc_port", cfg.GRPC.Port)

	// Initialize OpenTelemetry tracing
	telemetryConfig := telemetry.TelemetryConfig{
		ServiceName:    "admin-service",
		ServiceVersion: "1.0.0",
		Environment:    getEnv("APP_ENV", "development"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		Enabled:        getEnvBool("OTEL_ENABLED", true),
	}

	tel, err := telemetry.NewTelemetry(telemetryConfig)
	if err != nil {
		logger.Warn("Failed to initialize telemetry, continuing without tracing", "error", err)
	} else {
		logger.Info("Telemetry initialized successfully", "service", telemetryConfig.ServiceName, "endpoint", telemetryConfig.OTLPEndpoint)
		defer func() {
			ctx := context.Background()
			if err := tel.Shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown telemetry", "error", err)
			}
		}()
	}

	// Initialize database
	dbConfig := &backendCoreConfig.DatabaseConfig{
		Type:                       cfg.Database.Type,
		Host:                       cfg.Database.Host,
		Port:                       cfg.Database.Port,
		Username:                   cfg.Database.Username,
		Password:                   cfg.Database.Password,
		Database:                   cfg.Database.Name,
		SSLMode:                    cfg.Database.SSLMode,
		MaxOpenConns:               cfg.Database.MaxOpenConns,
		MaxIdleConns:               cfg.Database.MaxIdleConns,
		ConnMaxLifetime:            utils.ParseDuration(cfg.Database.ConnMaxLifetime),
		ConnMaxIdleTime:            utils.ParseDuration(cfg.Database.ConnMaxIdleTime),
		QueryTimeout:               30 * time.Second,
		AcquireTimeout:             30 * time.Second,
		AcquireRetryAttempts:       3,
		PreparedStatementCacheSize: 100,
		LogLevel:                   cfg.Logging.Level,
		SlowQueryThreshold:         utils.ParseDuration("1s"), // Default slow query threshold
	}

	// Create database using factory
	factory := database.NewDatabaseFactory()
	db, err := factory.CreatePostgreSQLDatabase(dbConfig)
	if err != nil {
		logger.Fatal("Failed to create database", "error", err)
	}

	// Connect to database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer func() {
		if err := db.Disconnect(ctx); err != nil {
			logger.Error("Failed to disconnect from database", "error", err)
		}
	}()

	logger.Info("Database connected successfully")

	// Extract GORM DB for repositories
	if postgresWrapper, ok := db.(*postgresql.DatabaseWrapper); ok {
		postgresDB := postgresWrapper.PostgreSQLDatabase
		rawGormDB := postgresDB.GetGormDB()

		// Initialize repositories
		userEventRepo := persistence.NewUserEventRepository(rawGormDB, logger)

		// Initialize services
		userEventService := services.NewUserEventService(userEventRepo, logger)

		// Start gRPC and HTTP servers
		startServers(cfg, userEventService, logger)
	} else {
		logger.Fatal("Failed to extract GORM database")
	}
}

func startServers(cfg *adminConfig.Config, userEventService *services.UserEventService, logger *logging.Logger) {
	// Start gRPC server
	go startGRPCServer(cfg, userEventService, logger)

	// Start HTTP server with WebSocket support
	startHTTPServer(cfg, userEventService, logger)
}

func startGRPCServer(cfg *adminConfig.Config, userEventService *services.UserEventService, logger *logging.Logger) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", cfg.GRPC.Host, cfg.GRPC.Port))
	if err != nil {
		logger.Fatal("Failed to listen on gRPC port", "error", err, "port", cfg.GRPC.Port)
	}

	// Build gRPC server with middleware using backend-core
	grpcServerConfig := &grpcserver.ServerConfig{
		Host:                 cfg.GRPC.Host,
		Port:                 cfg.GRPC.Port,
		MaxMessageSize:       10 * 1024 * 1024, // 10MB
		ConnectionTimeout:    120 * time.Second,
		KeepaliveTime:        30 * time.Second,
		KeepaliveTimeout:     10 * time.Second,
		MaxConcurrentStreams: 100,
	}

	// Configure authentication
	authConfig := &grpcauth.AuthConfig{
		Enabled:     cfg.GRPC.Auth.Enabled,
		JWTSecret:   cfg.GRPC.Auth.JWTSecret,
		JWTIssuer:   cfg.GRPC.Auth.JWTIssuer,
		JWTAudience: "microservices-clients",
		ExemptMethods: []string{
			"/grpc.health.v1.Health/Check",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
		APIKeys: make(map[string]string),
	}

	// Add API keys if configured
	if cfg.GRPC.Auth.APIKey != "" {
		authConfig.APIKeys["auth-service"] = cfg.GRPC.Auth.APIKey
	}

	// Configure logging
	loggingConfig := &grpclogging.LoggingConfig{
		LogPayload:        false,
		LogPayloadOnError: true,
	}

	// Build server with all interceptors in proper order
	grpcSrv := grpcserver.NewServerBuilder(grpcServerConfig).
		WithRecovery(logger).               // 1. Catch panics first
		WithLogging(logger, loggingConfig). // 2. Log all requests
		WithTracing().                      // 3. Add tracing
		WithAuth(logger, authConfig).       // 4. Authenticate
		WithValidation().                   // 5. Validate input
		Build()

	// Register service
	adminServer := grpcServer.NewAdminServer(userEventService, logger)
	pb.RegisterAdminServiceServer(grpcSrv, adminServer)

	// Register reflection service for grpc_cli and similar tools
	reflection.Register(grpcSrv)

	logger.Info("gRPC server starting with middleware", "address", lis.Addr().String(), "auth_enabled", cfg.GRPC.Auth.Enabled, "middleware", "recovery,logging,tracing,auth,validation")

	if err := grpcSrv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve gRPC", "error", err)
	}
}

func startHTTPServer(cfg *adminConfig.Config, userEventService *services.UserEventService, logger *logging.Logger) {
	// Set Gin mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "admin-service",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// WebSocket chat endpoint
	chatHub := websocket.NewChatHub(logger)
	go chatHub.Run()

	chatHandler := websocket.NewChatHandler(chatHub, logger)
	router.GET("/chat", chatHandler.HandleChat)
	router.GET("/chat/stats", chatHandler.HandleChatStats)

	// API routes
	api := router.Group("/api/v1")
	{
		// User events endpoint (HTTP alternative to gRPC)
		api.GET("/events", func(c *gin.Context) {
			// This can be expanded to provide HTTP access to events
			c.JSON(http.StatusOK, gin.H{
				"message": "Use gRPC for user events or WebSocket for chat",
			})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("HTTP server starting", "address", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
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

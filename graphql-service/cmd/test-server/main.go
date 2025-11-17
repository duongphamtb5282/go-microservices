package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"graphql-service/internal/infrastructure/database"
	"graphql-service/internal/infrastructure/persistence/mongodb"
	"graphql-service/internal/interfaces/graphql"
	"graphql-service/internal/interfaces/graphql/resolvers"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Configuration
	mongoURI := getEnv("MONGO_URI", "mongodb://admin:admin_password@localhost:27017/graphql_service?authSource=admin")
	port := getEnv("PORT", "8086")

	// Connect to MongoDB
	client, err := connectToMongoDB(mongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Get database
	db := client.Database("graphql_service")

	// Initialize database with migrations
	if err := database.Initialize(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create repositories
	userRepo := mongodb.NewUserRepository(db.Collection("users"), nil)

	// Create resolvers
	userResolver := resolvers.NewUserResolver(userRepo, nil)

	// Create GraphQL server
	server := graphql.NewTestServer(userResolver, nil)

	// Start server in a goroutine
	go func() {
		log.Printf("GraphQL test server starting on port %s", port)
		if err := server.Start(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down GraphQL test server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = ctx
	log.Println("GraphQL test server stopped")
}

// connectToMongoDB establishes connection to MongoDB
func connectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Connected to MongoDB successfully")
	return client, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

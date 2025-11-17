package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"graphql-service/internal/infrastructure/database/migration"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		mongoURI = flag.String("uri", "mongodb://admin:admin_password@localhost:27017/graphql_service?authSource=admin", "MongoDB connection URI")
		action   = flag.String("action", "migrate", "Action to perform: migrate, status, rollback")
	)
	flag.Parse()

	// Connect to MongoDB
	client, err := connectToMongoDB(*mongoURI)
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

	// Create migration runner
	migrationRunner := migration.NewMigrationRunner(db, nil)

	// Execute action
	switch *action {
	case "migrate":
		if err := migrationRunner.RunMigrations(context.TODO()); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("✅ Migrations completed successfully")
	case "status":
		migrations, err := migrationRunner.GetMigrationStatus(context.TODO())
		if err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
		fmt.Println("📊 Migration Status:")
		fmt.Println("===================")
		if len(migrations) == 0 {
			fmt.Println("No migrations applied yet")
		} else {
			for _, m := range migrations {
				fmt.Printf("✅ %s: %s (applied at: %s)\n", m.Version, m.Description, m.AppliedAt.Format("2006-01-02 15:04:05"))
			}
		}
	case "rollback":
		fmt.Println("⚠️  Rollback functionality not implemented yet")
		os.Exit(1)
	default:
		fmt.Printf("Unknown action: %s\n", *action)
		fmt.Println("Available actions: migrate, status, rollback")
		os.Exit(1)
	}
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

	fmt.Println("✅ Connected to MongoDB successfully")
	return client, nil
}

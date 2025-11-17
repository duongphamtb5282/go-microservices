package database

import (
	"context"
	"log"

	"graphql-service/internal/infrastructure/database/migration"

	"go.mongodb.org/mongo-driver/mongo"
)

// Initialize sets up the database with collections and indexes using migrations
func Initialize(db *mongo.Database) error {
	ctx := context.Background()

	// Run migrations
	migrationRunner := migration.NewMigrationRunner(db, nil)
	if err := migrationRunner.RunMigrations(ctx); err != nil {
		return err
	}

	log.Println("Database initialization completed successfully")
	return nil
}

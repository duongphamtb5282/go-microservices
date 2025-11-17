package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MigrationInterface defines the interface for individual migrations
type MigrationInterface interface {
	Version() string
	Description() string
	Up(ctx context.Context, db *mongo.Database) error
	Down(ctx context.Context, db *mongo.Database) error
	Checksum() string
}

// GetAllMigrations returns all available migrations
func GetAllMigrations() []MigrationInterface {
	return []MigrationInterface{
		NewUsersMigration001(),
		NewOrdersMigration002(),
		// Add new migrations here as you create them
		// NewProductsMigration003(),
		// NewNotificationsMigration004(),
	}
}

// GetMigrationByVersion returns a specific migration by version
func GetMigrationByVersion(version string) MigrationInterface {
	migrations := GetAllMigrations()
	for _, migration := range migrations {
		if migration.Version() == version {
			return migration
		}
	}
	return nil
}

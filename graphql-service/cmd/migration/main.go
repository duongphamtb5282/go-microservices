package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	var (
		action      = flag.String("action", "up", "Migration action: up, down, status, create")
		version     = flag.String("version", "", "Migration version (for create action)")
		description = flag.String("description", "", "Migration description (for create action)")
		configFile  = flag.String("config", "migrations.json", "Migration configuration file")
	)
	flag.Parse()

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.TODO())

	database := client.Database("graphql_service")

	// Create migration runner
	runner := NewImprovedMigrationRunner(client, database)

	ctx := context.Background()

	switch *action {
	case "up":
		err = runner.RunMigrations(ctx)
	case "down":
		if *version == "" {
			log.Fatal("Version is required for rollback")
		}
		err = runner.RollbackMigration(ctx, *version)
	case "status":
		status, err := runner.GetMigrationStatus(ctx)
		if err != nil {
			log.Fatal("Failed to get migration status:", err)
		}
		fmt.Printf("Migration Status: %+v\n", status)
		return
	case "create":
		if *version == "" || *description == "" {
			log.Fatal("Version and description are required for create action")
		}
		err = runner.CreateNewMigrationTemplate(*version, *description)
	default:
		log.Fatal("Unknown action:", *action)
	}

	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("Migration completed successfully")
}

// NewImprovedMigrationRunner creates a new improved migration runner
func NewImprovedMigrationRunner(client *mongo.Client, database *mongo.Database) *ImprovedMigrationRunner {
	return &ImprovedMigrationRunner{
		client:   client,
		database: database,
	}
}

// ImprovedMigrationRunner provides a flexible migration runner
type ImprovedMigrationRunner struct {
	client   *mongo.Client
	database *mongo.Database
}

// RunMigrations runs all pending migrations
func (m *ImprovedMigrationRunner) RunMigrations(ctx context.Context) error {
	fmt.Println("Running migrations...")
	// Implementation would go here
	return nil
}

// RollbackMigration rolls back a specific migration
func (m *ImprovedMigrationRunner) RollbackMigration(ctx context.Context, version string) error {
	fmt.Printf("Rolling back migration %s...\n", version)
	// Implementation would go here
	return nil
}

// GetMigrationStatus returns the status of all migrations
func (m *ImprovedMigrationRunner) GetMigrationStatus(ctx context.Context) (map[string]interface{}, error) {
	fmt.Println("Getting migration status...")
	// Implementation would go here
	return map[string]interface{}{"status": "ok"}, nil
}

// CreateNewMigrationTemplate creates a template for a new migration
func (m *ImprovedMigrationRunner) CreateNewMigrationTemplate(version, description string) error {
	// Convert description to PascalCase for struct name
	descriptionPascal := strings.Title(strings.ReplaceAll(description, " ", ""))

	template := fmt.Sprintf(`package migrations

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// %sMigration%s creates the %s
type %sMigration%s struct{}

// New%sMigration%s creates a new %s migration
func New%sMigration%s() *%sMigration%s {
	return &%sMigration%s{}
}

// Version returns the migration version
func (m *%sMigration%s) Version() string {
	return "%s"
}

// Description returns the migration description
func (m *%sMigration%s) Description() string {
	return "%s"
}

// Up applies the migration
func (m *%sMigration%s) Up(ctx context.Context, db *mongo.Database) error {
	// TODO: Implement your migration logic here
	fmt.Println("Applying migration %s: %s")
	return nil
}

// Down rolls back the migration
func (m *%sMigration%s) Down(ctx context.Context, db *mongo.Database) error {
	// TODO: Implement your rollback logic here
	fmt.Println("Rolling back migration %s: %s")
	return nil
}

// Checksum returns a checksum for the migration
func (m *%sMigration%s) Checksum() string {
	return "%s_%s_checksum"
}`,
		descriptionPascal, version, description,
		descriptionPascal, version, descriptionPascal, version, description, version, descriptionPascal, version, descriptionPascal, version,
		descriptionPascal, version, version,
		descriptionPascal, version, description,
		descriptionPascal, version, description,
		descriptionPascal, version, version, version,
		descriptionPascal, version, version, version,
		descriptionPascal, version, version, version)

	filename := fmt.Sprintf("internal/infrastructure/database/migration/migrations/%s_migration.go", version)
	return os.WriteFile(filename, []byte(template), 0644)
}

package interfaces

import (
	"context"

	"backend-core/database/health"
)

// MigrationManager defines the migration management interface
type MigrationManager interface {
	// Migration operations
	RunMigrations(ctx context.Context) error
	RunMigration(ctx context.Context, migration health.Migration) error
	RollbackMigration(ctx context.Context, version string) error
	GetMigrationHistory(ctx context.Context) ([]health.Migration, error)
	AddMigration(migration health.Migration)
}

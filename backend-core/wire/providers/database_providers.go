package providers

import (
	"backend-core/config"
	"backend-core/database"
)

// DatabaseProvider creates a database provider
func DatabaseProvider(cfg *config.Config) database.Database {
	// Implementation would create database connection
	return nil
}

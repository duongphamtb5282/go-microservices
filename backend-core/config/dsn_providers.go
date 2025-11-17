package config

import "fmt"

// NewDSNProvider creates a DSN provider based on database type
func NewDSNProvider(dbType string, config *DatabaseConfig) DsnProvider {
	switch dbType {
	case "postgresql", "postgres":
		if config.PostgreSQL != nil {
			return config.PostgreSQL
		}
		// Fallback to basic config
		postgresConfig := NewPostgreSQLConfig()
		postgresConfig.SetConnection(config.Host, config.Port, config.Database)
		postgresConfig.SetCredentials(config.Username, config.Password)
		postgresConfig.SetSSLMode(config.SSLMode)
		return postgresConfig
	case "mongodb":
		if config.MongoDB != nil {
			return config.MongoDB
		}
		// Fallback to basic config
		mongoConfig := NewMongoDBConfig()
		mongoConfig.SetDatabase(config.Database)
		mongoConfig.SetCredentials(config.Username, config.Password)
		mongoConfig.AddHost(config.Host, fmt.Sprintf("%d", config.Port))
		return mongoConfig
	default:
		return nil
	}
}

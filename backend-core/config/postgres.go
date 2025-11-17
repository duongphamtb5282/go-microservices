package config

import (
	"fmt"
	"strings"
)

// PostgreSQLConfig holds PostgreSQL-specific configuration
type PostgreSQLConfig struct {
	Host     string            `mapstructure:"host"`
	Port     int               `mapstructure:"port"`
	Database string            `mapstructure:"database"`
	Username string            `mapstructure:"username"`
	Password string            `mapstructure:"password"`
	SSLMode  string            `mapstructure:"ssl_mode"`
	Params   map[string]string `mapstructure:"params"`
}

// NewPostgreSQLConfig creates a new PostgreSQL configuration with defaults
func NewPostgreSQLConfig() *PostgreSQLConfig {
	return &PostgreSQLConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "microservices",
		Username: "postgres",
		Password: "",
		SSLMode:  "disable",
		Params:   make(map[string]string),
	}
}

// DSN returns PostgreSQL connection string
func (p *PostgreSQLConfig) Dsn() string {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.Username,
		p.Password,
		p.Host,
		p.Port,
		p.Database,
		p.SSLMode,
	)

	// Add additional parameters
	if len(p.Params) > 0 {
		var paramParts []string
		for key, value := range p.Params {
			paramParts = append(paramParts, fmt.Sprintf("%s=%s", key, value))
		}
		dsn += "&" + strings.Join(paramParts, "&")
	}

	return dsn
}

// AddParam adds a parameter to the DSN
func (p *PostgreSQLConfig) AddParam(key, value string) {
	if p.Params == nil {
		p.Params = make(map[string]string)
	}
	p.Params[key] = value
}

// SetSSLMode sets the SSL mode
func (p *PostgreSQLConfig) SetSSLMode(mode string) {
	p.SSLMode = mode
}

// SetCredentials sets username and password
func (p *PostgreSQLConfig) SetCredentials(username, password string) {
	p.Username = username
	p.Password = password
}

// SetConnection sets host, port, and database
func (p *PostgreSQLConfig) SetConnection(host string, port int, database string) {
	p.Host = host
	p.Port = port
	p.Database = database
}

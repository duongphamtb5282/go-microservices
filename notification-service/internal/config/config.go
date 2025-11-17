package config

import (
	"backend-core/config"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Server  ServerConfig         `mapstructure:"server" json:"server" yaml:"server"`
	Kafka   KafkaConfig          `mapstructure:"kafka" json:"kafka" yaml:"kafka"`
	Logging config.LoggingConfig `mapstructure:"logging" json:"logging" yaml:"logging"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string `mapstructure:"port" json:"port" yaml:"port"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers      []string     `mapstructure:"brokers" json:"brokers" yaml:"brokers"`
	GroupID      string       `mapstructure:"group_id" json:"group_id" yaml:"group_id"`
	Topics       TopicsConfig `mapstructure:"topics" json:"topics" yaml:"topics"`
	RetryCount   int          `mapstructure:"retry_count" json:"retry_count" yaml:"retry_count"`
	RetryBackoff string       `mapstructure:"retry_backoff" json:"retry_backoff" yaml:"retry_backoff"`
}

// TopicsConfig holds Kafka topics configuration
type TopicsConfig struct {
	UserEvents string `mapstructure:"user_events" json:"user_events" yaml:"user_events"`
	AuthEvents string `mapstructure:"auth_events" json:"auth_events" yaml:"auth_events"`
	AuditLogs  string `mapstructure:"audit_logs" json:"audit_logs" yaml:"audit_logs"`
}

// Load loads configuration from environment variables and files
func Load() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.setDefaults()

	// Override with environment variables
	cfg.loadFromEnv()

	return cfg, nil
}

// setDefaults sets default configuration values
func (c *Config) setDefaults() {
	// Server defaults
	c.Server.Port = "8086"

	// Kafka defaults
	c.Kafka.Brokers = []string{"localhost:9092"}
	c.Kafka.GroupID = "notification-service"
	c.Kafka.Topics.UserEvents = "user.events"
	c.Kafka.Topics.AuthEvents = "auth.events"
	c.Kafka.Topics.AuditLogs = "audit.logs"
	c.Kafka.RetryCount = 3
	c.Kafka.RetryBackoff = "1s"

	// Logging defaults
	c.Logging.Level = "info"
	c.Logging.Format = "json"
	c.Logging.Output = "stdout"
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		c.Server.Port = port
	}

	// Kafka configuration
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		c.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		c.Kafka.GroupID = groupID
	}
	if userEvents := os.Getenv("KAFKA_TOPIC_USER_EVENTS"); userEvents != "" {
		c.Kafka.Topics.UserEvents = userEvents
	}
	if authEvents := os.Getenv("KAFKA_TOPIC_AUTH_EVENTS"); authEvents != "" {
		c.Kafka.Topics.AuthEvents = authEvents
	}
	if auditLogs := os.Getenv("KAFKA_TOPIC_AUDIT_LOGS"); auditLogs != "" {
		c.Kafka.Topics.AuditLogs = auditLogs
	}
	if retryCount := os.Getenv("KAFKA_RETRY_COUNT"); retryCount != "" {
		if count, err := strconv.Atoi(retryCount); err == nil {
			c.Kafka.RetryCount = count
		}
	}

	// Logging configuration
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Logging.Format = format
	}
	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		c.Logging.Output = output
	}
}

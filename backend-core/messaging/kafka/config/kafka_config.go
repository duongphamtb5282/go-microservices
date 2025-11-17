package config

import (
	"errors"
	"fmt"
	"time"
)

// KafkaConfig contains the main Kafka configuration
type KafkaConfig struct {
	BootstrapServers []string           `yaml:"bootstrap_servers" json:"bootstrap_servers"`
	ClientID         string             `yaml:"client_id" json:"client_id"`
	Security         *SecurityConfig    `yaml:"security" json:"security"`
	Producer         *ProducerConfig    `yaml:"producer" json:"producer"`
	Consumer         *ConsumerConfig    `yaml:"consumer" json:"consumer"`
	Topics           map[string]*TopicConfig `yaml:"topics" json:"topics"`
	Monitoring       *MonitoringConfig  `yaml:"monitoring" json:"monitoring"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	Protocol         string `yaml:"protocol" json:"protocol"`                   // PLAINTEXT, SSL, SASL_PLAINTEXT, SASL_SSL
	SASLMechanism    string `yaml:"sasl_mechanism" json:"sasl_mechanism"`        // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	SASLUsername     string `yaml:"sasl_username" json:"sasl_username"`
	SASLPassword     string `yaml:"sasl_password" json:"sasl_password"`
	SSLKeyLocation   string `yaml:"ssl_key_location" json:"ssl_key_location"`
	SSLCertLocation  string `yaml:"ssl_cert_location" json:"ssl_cert_location"`
	SSLKeyPassword   string `yaml:"ssl_key_password" json:"ssl_key_password"`
	TruststoreLocation string `yaml:"truststore_location" json:"truststore_location"`
	TruststorePassword string `yaml:"truststore_password" json:"truststore_password"`
}

// ProducerConfig contains producer-specific configuration
type ProducerConfig struct {
	Acks             string        `yaml:"acks" json:"acks"`                       // all, 1, 0
	Retries          int           `yaml:"retries" json:"retries"`
	BatchSize        int           `yaml:"batch_size" json:"batch_size"`
	LingerMs         int           `yaml:"linger_ms" json:"linger_ms"`
	CompressionType  string        `yaml:"compression_type" json:"compression_type"` // none, gzip, snappy, lz4, zstd
	MaxRequestSize   int           `yaml:"max_request_size" json:"max_request_size"`
	DeliveryTimeout  time.Duration `yaml:"delivery_timeout" json:"delivery_timeout"`
	RequestTimeout   time.Duration `yaml:"request_timeout" json:"request_timeout"`
	EnableIdempotent bool          `yaml:"enable_idempotent" json:"enable_idempotent"`
	MaxInFlight      int           `yaml:"max_in_flight" json:"max_in_flight"`
}

// ConsumerConfig contains consumer-specific configuration
type ConsumerConfig struct {
	GroupID          string        `yaml:"group_id" json:"group_id"`
	AutoOffsetReset  string        `yaml:"auto_offset_reset" json:"auto_offset_reset"` // earliest, latest, none
	EnableAutoCommit bool          `yaml:"enable_auto_commit" json:"enable_auto_commit"`
	SessionTimeout   time.Duration `yaml:"session_timeout" json:"session_timeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval" json:"heartbeat_interval"`
	MaxPollRecords   int           `yaml:"max_poll_records" json:"max_poll_records"`
	MaxPollInterval   time.Duration `yaml:"max_poll_interval" json:"max_poll_interval"`
	FetchMinBytes    int           `yaml:"fetch_min_bytes" json:"fetch_min_bytes"`
	FetchMaxWait     time.Duration `yaml:"fetch_max_wait" json:"fetch_max_wait"`
	IsolationLevel   string        `yaml:"isolation_level" json:"isolation_level"`   // read_uncommitted, read_committed
}

// TopicConfig contains topic-specific configuration
type TopicConfig struct {
	Name              string            `yaml:"name" json:"name"`
	Partitions        int               `yaml:"partitions" json:"partitions"`
	ReplicationFactor int               `yaml:"replication_factor" json:"replication_factor"`
	Config            map[string]string `yaml:"config" json:"config"`
	RetentionMs       int64             `yaml:"retention_ms" json:"retention_ms"`
	SegmentMs         int64             `yaml:"segment_ms" json:"segment_ms"`
	CleanupPolicy     string            `yaml:"cleanup_policy" json:"cleanup_policy"`
	CompressionType   string            `yaml:"compression_type" json:"compression_type"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled           bool          `yaml:"enabled" json:"enabled"`
	MetricsPort       int           `yaml:"metrics_port" json:"metrics_port"`
	HealthCheckPath   string        `yaml:"health_check_path" json:"health_check_path"`
	PrometheusEnabled bool          `yaml:"prometheus_enabled" json:"prometheus_enabled"`
	JMXEnabled        bool          `yaml:"jmx_enabled" json:"jmx_enabled"`
	LogLevel          string        `yaml:"log_level" json:"log_level"`
	StatsInterval     time.Duration `yaml:"stats_interval" json:"stats_interval"`
}

// DefaultKafkaConfig returns a default Kafka configuration
func DefaultKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		BootstrapServers: []string{"localhost:9092"},
		ClientID:         "kafka-client",
		Security: &SecurityConfig{
			Protocol: "PLAINTEXT",
		},
		Producer: &ProducerConfig{
			Acks:             "all",
			Retries:          3,
			BatchSize:        16384,
			LingerMs:         5,
			CompressionType:  "snappy",
			MaxRequestSize:   1048576,
			DeliveryTimeout:  120 * time.Second,
			RequestTimeout:   30 * time.Second,
			EnableIdempotent: true,
			MaxInFlight:      5,
		},
		Consumer: &ConsumerConfig{
			GroupID:           "default-group",
			AutoOffsetReset:   "latest",
			EnableAutoCommit:  true,
			SessionTimeout:    30 * time.Second,
			HeartbeatInterval: 3 * time.Second,
			MaxPollRecords:    500,
			MaxPollInterval:   5 * time.Minute,
			FetchMinBytes:     1,
			FetchMaxWait:      500 * time.Millisecond,
			IsolationLevel:    "read_uncommitted",
		},
		Topics: make(map[string]*TopicConfig),
		Monitoring: &MonitoringConfig{
			Enabled:           true,
			MetricsPort:       9090,
			HealthCheckPath:   "/health",
			PrometheusEnabled: true,
			JMXEnabled:        false,
			LogLevel:          "info",
			StatsInterval:     60 * time.Second,
		},
	}
}

// Validate validates the Kafka configuration
func (c *KafkaConfig) Validate() error {
	if len(c.BootstrapServers) == 0 {
		return errors.New("bootstrap servers cannot be empty")
	}
	
	if c.ClientID == "" {
		return errors.New("client ID cannot be empty")
	}
	
	if c.Producer != nil {
		if err := c.Producer.Validate(); err != nil {
			return fmt.Errorf("producer config validation failed: %w", err)
		}
	}
	
	if c.Consumer != nil {
		if err := c.Consumer.Validate(); err != nil {
			return fmt.Errorf("consumer config validation failed: %w", err)
		}
	}
	
	return nil
}

// Validate validates the producer configuration
func (c *ProducerConfig) Validate() error {
	if c.Acks != "all" && c.Acks != "1" && c.Acks != "0" {
		return errors.New("acks must be 'all', '1', or '0'")
	}
	
	if c.Retries < 0 {
		return errors.New("retries cannot be negative")
	}
	
	if c.BatchSize <= 0 {
		return errors.New("batch size must be positive")
	}
	
	if c.LingerMs < 0 {
		return errors.New("linger ms cannot be negative")
	}
	
	validCompressionTypes := []string{"none", "gzip", "snappy", "lz4", "zstd"}
	if !contains(validCompressionTypes, c.CompressionType) {
		return errors.New("invalid compression type")
	}
	
	return nil
}

// Validate validates the consumer configuration
func (c *ConsumerConfig) Validate() error {
	if c.GroupID == "" {
		return errors.New("group ID cannot be empty")
	}
	
	if c.AutoOffsetReset != "earliest" && c.AutoOffsetReset != "latest" && c.AutoOffsetReset != "none" {
		return errors.New("auto offset reset must be 'earliest', 'latest', or 'none'")
	}
	
	if c.SessionTimeout <= 0 {
		return errors.New("session timeout must be positive")
	}
	
	if c.HeartbeatInterval <= 0 {
		return errors.New("heartbeat interval must be positive")
	}
	
	if c.MaxPollRecords <= 0 {
		return errors.New("max poll records must be positive")
	}
	
	if c.MaxPollInterval <= 0 {
		return errors.New("max poll interval must be positive")
	}
	
	validIsolationLevels := []string{"read_uncommitted", "read_committed"}
	if !contains(validIsolationLevels, c.IsolationLevel) {
		return errors.New("invalid isolation level")
	}
	
	return nil
}

// Validate validates the topic configuration
func (c *TopicConfig) Validate() error {
	if c.Name == "" {
		return errors.New("topic name cannot be empty")
	}
	
	if c.Partitions <= 0 {
		return errors.New("partitions must be positive")
	}
	
	if c.ReplicationFactor <= 0 {
		return errors.New("replication factor must be positive")
	}
	
	validCleanupPolicies := []string{"delete", "compact", "delete,compact"}
	if c.CleanupPolicy != "" && !contains(validCleanupPolicies, c.CleanupPolicy) {
		return errors.New("invalid cleanup policy")
	}
	
	validCompressionTypes := []string{"none", "gzip", "snappy", "lz4", "zstd"}
	if c.CompressionType != "" && !contains(validCompressionTypes, c.CompressionType) {
		return errors.New("invalid compression type")
	}
	
	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

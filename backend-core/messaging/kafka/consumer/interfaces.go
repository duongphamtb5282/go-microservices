package consumer

import (
	"context"
	"time"
)

// Message represents a Kafka message
type Message struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int32
	Offset    int64
	Timestamp time.Time
}

// ConsumerMessage represents a consumed message
type ConsumerMessage struct {
	Topic     string
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Partition int32
	Offset    int64
	Timestamp time.Time
}

// Consumer defines the interface for Kafka consumers
type Consumer interface {
	// Subscribe subscribes to topics
	Subscribe(topics []string) error

	// Unsubscribe unsubscribes from topics
	Unsubscribe(topics []string) error

	// Poll polls for messages
	Poll(ctx context.Context, timeout time.Duration) ([]*ConsumerMessage, error)

	// Commit commits offsets
	Commit() error

	// Close closes the consumer
	Close() error

	// GetMetadata returns consumer metadata
	GetMetadata() (*ConsumerMetadata, error)
}

// ConsumerMetadata contains consumer metadata
type ConsumerMetadata struct {
	GroupID    string
	Topics     []string
	Partitions map[string][]int32
	Offsets    map[string]map[int32]int64
}

// ConsumerHandler defines the interface for message handlers
type ConsumerHandler interface {
	// Handle handles a consumed message
	Handle(ctx context.Context, message *ConsumerMessage) error

	// GetTopics returns the topics this handler is interested in
	GetTopics() []string

	// GetGroupID returns the consumer group ID
	GetGroupID() string
}

// ConsumerGroup defines the interface for consumer groups
type ConsumerGroup interface {
	// Start starts the consumer group
	Start(ctx context.Context) error

	// Stop stops the consumer group
	Stop() error

	// AddHandler adds a message handler
	AddHandler(handler ConsumerHandler) error

	// RemoveHandler removes a message handler
	RemoveHandler(handler ConsumerHandler) error

	// GetStatus returns the consumer group status
	GetStatus() (*ConsumerGroupStatus, error)
}

// ConsumerGroupStatus contains consumer group status
type ConsumerGroupStatus struct {
	GroupID    string
	State      string
	Members    []string
	Partitions map[string][]int32
	Lag        map[string]int64
}

// ConsumerGroupConfig contains consumer group configuration
type ConsumerGroupConfig struct {
	GroupID          string
	BootstrapServers []string
	ClientID         string
	AutoOffsetReset  string
	EnableAutoCommit bool
	SessionTimeout   time.Duration
	SecurityConfig   *SecurityConfig
	Topics           []string
	Config           *Config
}

// Config contains consumer configuration
type Config struct {
	BootstrapServers []string
	GroupID          string
	ClientID         string
	AutoOffsetReset  string
	EnableAutoCommit bool
	SessionTimeout   time.Duration
	SecurityConfig   *SecurityConfig
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	Protocol        string
	SASLMechanism   string
	SASLUsername    string
	SASLPassword    string
	SSLKeyLocation  string
	SSLCertLocation string
	SSLKeyPassword  string
}

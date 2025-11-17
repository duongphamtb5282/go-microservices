package producer

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

// ProducerMessage represents a message to be produced
type ProducerMessage struct {
	Topic   string
	Key     []byte
	Value   []byte
	Headers map[string]string
}

// Producer defines the interface for Kafka producers
type Producer interface {
	// Send sends a message to Kafka
	Send(ctx context.Context, message *ProducerMessage) error

	// SendBatch sends multiple messages to Kafka
	SendBatch(ctx context.Context, messages []*ProducerMessage) error

	// SendAsync sends a message asynchronously
	SendAsync(ctx context.Context, message *ProducerMessage) error

	// Close closes the producer
	Close() error

	// Flush flushes any pending messages
	Flush() error
}

// Config contains producer configuration
type Config struct {
	BootstrapServers []string
	ClientID         string
	Acks             string
	Retries          int
	BatchSize        int
	LingerMs         int
	CompressionType  string
	SecurityConfig   *SecurityConfig
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	SecurityProtocol string
	SASLMechanism    string
	SASLUsername     string
	SASLPassword     string
	SSLKeyLocation   string
	SSLCertLocation  string
	SSLKeyPassword   string
}

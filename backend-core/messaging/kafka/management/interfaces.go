package management

import (
	"backend-core/messaging/kafka/consumer"
	"backend-core/messaging/kafka/producer"
)

// KafkaManager defines the interface for Kafka management
type KafkaManager interface {
	// CreateProducer creates a new producer
	CreateProducer(config *producer.Config) (producer.Producer, error)

	// CreateConsumer creates a new consumer
	CreateConsumer(config *consumer.Config) (consumer.Consumer, error)

	// CreateConsumerGroup creates a consumer group
	CreateConsumerGroup(config *ConsumerGroupConfig) (consumer.ConsumerGroup, error)

	// GetClusterInfo returns cluster information
	GetClusterInfo() (*ClusterInfo, error)

	// CreateTopic creates a new topic
	CreateTopic(config *TopicConfig) error

	// DeleteTopic deletes a topic
	DeleteTopic(topic string) error

	// ListTopics lists all topics
	ListTopics() ([]string, error)

	// GetTopicMetadata returns topic metadata
	GetTopicMetadata(topic string) (*TopicMetadata, error)

	// Close closes the manager and all its resources
	Close() error
}

// ConsumerGroupConfig contains consumer group configuration
type ConsumerGroupConfig struct {
	BootstrapServers []string
	GroupID          string
	Topics           []string
	Handlers         []consumer.ConsumerHandler
	Config           *consumer.Config
}

// ClusterInfo contains cluster information
type ClusterInfo struct {
	Brokers    []string
	Controller string
	Topics     []string
	Partitions int
	Replicas   int
}

// TopicMetadata contains topic metadata
type TopicMetadata struct {
	Name       string
	Partitions []PartitionInfo
	Config     map[string]string
}

// PartitionInfo contains partition information
type PartitionInfo struct {
	ID       int32
	Leader   int32
	Replicas []int32
	ISR      []int32
	Offsets  map[string]int64
}

// TopicConfig contains topic configuration
type TopicConfig struct {
	Name              string
	Partitions        int
	ReplicationFactor int
	Config            map[string]string
}

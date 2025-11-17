package interfaces

import (
	"backend-core/messaging/kafka/consumer"
	"backend-core/messaging/kafka/management"
	"backend-core/messaging/kafka/producer"
)

// Re-export types from consumer package
type Message = consumer.Message
type ConsumerMessage = consumer.ConsumerMessage
type Consumer = consumer.Consumer
type ConsumerMetadata = consumer.ConsumerMetadata
type ConsumerHandler = consumer.ConsumerHandler
type ConsumerGroup = consumer.ConsumerGroup
type ConsumerGroupStatus = consumer.ConsumerGroupStatus
type ConsumerConfig = consumer.Config
type ConsumerSecurityConfig = consumer.SecurityConfig

// Re-export types from producer package
type ProducerMessage = producer.ProducerMessage
type Producer = producer.Producer
type ProducerConfig = producer.Config
type ProducerSecurityConfig = producer.SecurityConfig

// Re-export types from management package
type KafkaManager = management.KafkaManager
type ConsumerGroupConfig = management.ConsumerGroupConfig
type ClusterInfo = management.ClusterInfo
type TopicMetadata = management.TopicMetadata
type PartitionInfo = management.PartitionInfo
type TopicConfig = management.TopicConfig

// Legacy types for backward compatibility
type SecurityConfig = consumer.SecurityConfig

// Legacy interfaces for backward compatibility
// ConsumerGroup is already defined as a type alias above

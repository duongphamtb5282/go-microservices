package consumer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend-core/logging"
	"backend-core/messaging/kafka/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

// KafkaConsumer implements the Consumer interface using Confluent Kafka
type KafkaConsumer struct {
	consumer *kafka.Consumer
	config   *config.ConsumerConfig
	logger   *logging.Logger
	mu       sync.RWMutex
	closed   bool
	topics   []string
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(cfg *config.KafkaConfig, logger *logging.Logger) (Consumer, error) {
	if cfg.Consumer == nil {
		return nil, fmt.Errorf("consumer configuration is required")
	}

	// Create Confluent Kafka configuration map
	configMap := kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(cfg.BootstrapServers, ","),
		"group.id":                 cfg.Consumer.GroupID,
		"client.id":                cfg.ClientID,
		"enable.auto.commit":       cfg.Consumer.EnableAutoCommit,
		"session.timeout.ms":       int(cfg.Consumer.SessionTimeout.Milliseconds()),
		"heartbeat.interval.ms":   int(cfg.Consumer.HeartbeatInterval.Milliseconds()),
		"max.poll.interval.ms":     int(cfg.Consumer.MaxPollInterval.Milliseconds()),
		"fetch.min.bytes":          cfg.Consumer.FetchMinBytes,
		"fetch.max.wait.ms":        int(cfg.Consumer.FetchMaxWait.Milliseconds()),
		"max.partition.fetch.bytes": 1048576, // 1MB default
	}

	// Configure auto offset reset
	switch cfg.Consumer.AutoOffsetReset {
	case "earliest":
		configMap["auto.offset.reset"] = "earliest"
	case "latest":
		configMap["auto.offset.reset"] = "latest"
	case "none":
		configMap["auto.offset.reset"] = "error"
	default:
		configMap["auto.offset.reset"] = "latest"
	}

	// Configure isolation level
	switch cfg.Consumer.IsolationLevel {
	case "read_committed":
		configMap["isolation.level"] = "read_committed"
	case "read_uncommitted":
		configMap["isolation.level"] = "read_uncommitted"
	default:
		configMap["isolation.level"] = "read_uncommitted"
	}

	// Configure security
	if err := configureSecurity(&configMap, cfg.Security); err != nil {
		return nil, fmt.Errorf("failed to configure security: %w", err)
	}

	// Create consumer
	consumer, err := kafka.NewConsumer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	kafkaConsumer := &KafkaConsumer{
		consumer: consumer,
		config:   cfg.Consumer,
		logger:   logger,
		topics:   make([]string, 0),
	}

	return kafkaConsumer, nil
}

// Subscribe subscribes to topics
func (c *KafkaConsumer) Subscribe(topics []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	// Subscribe to topics
	topicList := make([]string, len(topics))
	copy(topicList, topics)

	err := c.consumer.SubscribeTopics(topicList, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	c.topics = append(c.topics, topics...)
	c.logger.Info("Subscribed to topics", zap.Strings("topics", topics))
	return nil
}

// Unsubscribe unsubscribes from topics
func (c *KafkaConsumer) Unsubscribe(topics []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	// Remove topics from the list
	for _, topic := range topics {
		for i, t := range c.topics {
			if t == topic {
				c.topics = append(c.topics[:i], c.topics[i+1:]...)
				break
			}
		}
	}

	// Unsubscribe from all topics and resubscribe to remaining ones
	if len(c.topics) > 0 {
		err := c.consumer.Unsubscribe()
		if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}

		err = c.consumer.SubscribeTopics(c.topics, nil)
		if err != nil {
			return fmt.Errorf("failed to resubscribe: %w", err)
		}
	} else {
		err := c.consumer.Unsubscribe()
		if err != nil {
			return fmt.Errorf("failed to unsubscribe: %w", err)
		}
	}

	c.logger.Info("Unsubscribed from topics", zap.Strings("topics", topics))
	return nil
}

// Poll polls for messages
func (c *KafkaConsumer) Poll(ctx context.Context, timeout time.Duration) ([]*ConsumerMessage, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, fmt.Errorf("consumer is closed")
	}
	c.mu.RUnlock()

	if len(c.topics) == 0 {
		return nil, fmt.Errorf("no topics subscribed")
	}

	var messages []*ConsumerMessage
	timeoutMs := int(timeout.Milliseconds())
	if timeoutMs <= 0 {
		timeoutMs = 100 // Default 100ms
	}

	// Poll for messages until timeout
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return messages, ctx.Err()
		default:
			remainingMs := int(time.Until(deadline).Milliseconds())
			if remainingMs <= 0 {
				return messages, nil
			}

			msg, err := c.consumer.ReadMessage(time.Duration(remainingMs) * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					// Timeout is expected, return what we have
					return messages, nil
				}
				c.logger.Error("Failed to read message",
					zap.Error(err),
				)
				continue
			}

			consumerMsg := &ConsumerMessage{
				Topic:     *msg.TopicPartition.Topic,
				Key:       msg.Key,
				Value:     msg.Value,
				Headers:   make(map[string]string),
				Partition: msg.TopicPartition.Partition,
				Offset:    int64(msg.TopicPartition.Offset),
				Timestamp: msg.Timestamp,
			}

			// Convert headers
			for _, header := range msg.Headers {
				consumerMsg.Headers[header.Key] = string(header.Value)
			}

			messages = append(messages, consumerMsg)
		}
	}

	return messages, nil
}

// Commit commits offsets
func (c *KafkaConsumer) Commit() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return fmt.Errorf("consumer is closed")
	}

	_, err := c.consumer.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit offsets: %w", err)
	}

	c.logger.Debug("Offsets committed")
	return nil
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	err := c.consumer.Close()
	if err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}

	c.closed = true
	c.logger.Info("Consumer closed successfully")
	return nil
}

// GetMetadata returns consumer metadata
func (c *KafkaConsumer) GetMetadata() (*ConsumerMetadata, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("consumer is closed")
	}

	metadata := &ConsumerMetadata{
		GroupID:    c.config.GroupID,
		Topics:     c.topics,
		Partitions: make(map[string][]int32),
		Offsets:    make(map[string]map[int32]int64),
	}

	// Get metadata for subscribed topics
	metadataResult, err := c.consumer.GetMetadata(nil, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	// Extract partition and offset information
	for _, topic := range c.topics {
		topicMetadata, exists := metadataResult.Topics[topic]
		if !exists {
			continue
		}

		partitions := make([]int32, 0, len(topicMetadata.Partitions))
		offsets := make(map[int32]int64)

		for partitionID, _ := range topicMetadata.Partitions {
			partitionIDInt32 := int32(partitionID)
			partitions = append(partitions, partitionIDInt32)

			// Get watermark offsets
			_, high, err := c.consumer.QueryWatermarkOffsets(topic, partitionIDInt32, 5000)
			if err == nil {
				// Use high watermark as offset
				offsets[partitionIDInt32] = int64(high)
			}
		}

		metadata.Partitions[topic] = partitions
		metadata.Offsets[topic] = offsets
	}

	return metadata, nil
}

// configureSecurity configures security settings for Confluent Kafka
func configureSecurity(configMap *kafka.ConfigMap, security *config.SecurityConfig) error {
	if security == nil {
		return nil
	}

	switch security.Protocol {
	case "PLAINTEXT":
		// No additional configuration needed
	case "SSL":
		(*configMap)["security.protocol"] = "ssl"
		if security.SSLCertLocation != "" {
			(*configMap)["ssl.certificate.location"] = security.SSLCertLocation
		}
		if security.SSLKeyLocation != "" {
			(*configMap)["ssl.key.location"] = security.SSLKeyLocation
		}
		if security.TruststoreLocation != "" {
			(*configMap)["ssl.ca.location"] = security.TruststoreLocation
		}
	case "SASL_PLAINTEXT":
		(*configMap)["security.protocol"] = "sasl_plaintext"
		(*configMap)["sasl.mechanism"] = security.SASLMechanism
		(*configMap)["sasl.username"] = security.SASLUsername
		(*configMap)["sasl.password"] = security.SASLPassword
	case "SASL_SSL":
		(*configMap)["security.protocol"] = "sasl_ssl"
		(*configMap)["sasl.mechanism"] = security.SASLMechanism
		(*configMap)["sasl.username"] = security.SASLUsername
		(*configMap)["sasl.password"] = security.SASLPassword
		if security.SSLCertLocation != "" {
			(*configMap)["ssl.certificate.location"] = security.SSLCertLocation
		}
		if security.SSLKeyLocation != "" {
			(*configMap)["ssl.key.location"] = security.SSLKeyLocation
		}
		if security.TruststoreLocation != "" {
			(*configMap)["ssl.ca.location"] = security.TruststoreLocation
		}
	default:
		return fmt.Errorf("unsupported security protocol: %s", security.Protocol)
	}

	return nil
}

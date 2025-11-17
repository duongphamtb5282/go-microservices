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

// KafkaConsumerGroup implements the ConsumerGroup interface using Confluent Kafka
type KafkaConsumerGroup struct {
	config   *ConsumerGroupConfig
	logger   *logging.Logger
	handlers map[string]ConsumerHandler
	mu       sync.RWMutex
	closed   bool
	consumer *kafka.Consumer
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewKafkaConsumerGroup creates a new Kafka consumer group
func NewKafkaConsumerGroup(config *ConsumerGroupConfig, logger *logging.Logger) (ConsumerGroup, error) {
	// Create Confluent Kafka configuration map
	configMap := kafka.ConfigMap{
		"bootstrap.servers":      strings.Join(config.BootstrapServers, ","),
		"group.id":               config.GroupID,
		"client.id":              config.ClientID,
		"enable.auto.commit":     config.EnableAutoCommit,
		"session.timeout.ms":     int(config.SessionTimeout.Milliseconds()),
		"heartbeat.interval.ms":  int(3 * time.Second.Milliseconds()),
		"max.poll.interval.ms":   int(5 * time.Minute.Milliseconds()),
	}

	// Configure auto offset reset
	switch config.AutoOffsetReset {
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
	configMap["isolation.level"] = "read_uncommitted"

	// Configure security
	if err := configureSecurity(&configMap, convertSecurityConfig(config.SecurityConfig)); err != nil {
		return nil, fmt.Errorf("failed to configure security: %w", err)
	}

	// Create consumer
	consumer, err := kafka.NewConsumer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	consumerGroup := &KafkaConsumerGroup{
		config:   config,
		logger:   logger,
		handlers: make(map[string]ConsumerHandler),
		consumer: consumer,
	}

	return consumerGroup, nil
}

// Start starts the consumer group
func (cg *KafkaConsumerGroup) Start(ctx context.Context) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.closed {
		return fmt.Errorf("consumer group is closed")
	}

	cg.ctx, cg.cancel = context.WithCancel(ctx)

	// Subscribe to topics
	if len(cg.config.Topics) > 0 {
		err := cg.consumer.SubscribeTopics(cg.config.Topics, nil)
		if err != nil {
			return fmt.Errorf("failed to subscribe to topics: %w", err)
		}
	}

	// Start consumer group in a goroutine
	cg.wg.Add(1)
	go func() {
		defer cg.wg.Done()
		cg.run()
	}()

	cg.logger.Info("Consumer group started", zap.String("group_id", cg.config.GroupID))
	return nil
}

// Stop stops the consumer group
func (cg *KafkaConsumerGroup) Stop() error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.closed {
		return nil
	}

	if cg.cancel != nil {
		cg.cancel()
	}

	cg.wg.Wait()

	if err := cg.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %w", err)
	}

	cg.closed = true
	cg.logger.Info("Consumer group stopped", zap.String("group_id", cg.config.GroupID))
	return nil
}

// AddHandler adds a message handler
func (cg *KafkaConsumerGroup) AddHandler(handler ConsumerHandler) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.closed {
		return fmt.Errorf("consumer group is closed")
	}

	handlerID := fmt.Sprintf("handler_%d", time.Now().UnixNano())
	cg.handlers[handlerID] = handler

	cg.logger.Info("Handler added to consumer group",
		zap.String("handler_id", handlerID),
		zap.Strings("topics", handler.GetTopics()),
	)

	return nil
}

// RemoveHandler removes a message handler
func (cg *KafkaConsumerGroup) RemoveHandler(handler ConsumerHandler) error {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	if cg.closed {
		return fmt.Errorf("consumer group is closed")
	}

	// Find and remove handler
	for id, h := range cg.handlers {
		if h == handler {
			delete(cg.handlers, id)
			cg.logger.Info("Handler removed from consumer group", zap.String("handler_id", id))
			return nil
		}
	}

	return fmt.Errorf("handler not found")
}

// GetStatus returns the consumer group status
func (cg *KafkaConsumerGroup) GetStatus() (*ConsumerGroupStatus, error) {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	if cg.closed {
		return nil, fmt.Errorf("consumer group is closed")
	}

	status := &ConsumerGroupStatus{
		GroupID:    cg.config.GroupID,
		State:      "running",
		Members:    make([]string, 0),
		Partitions: make(map[string][]int32),
		Lag:        make(map[string]int64),
	}

	// Add handler topics to partitions
	for _, handler := range cg.handlers {
		for _, topic := range handler.GetTopics() {
			status.Partitions[topic] = []int32{} // Would need actual partition info
		}
	}

	return status, nil
}

// run runs the consumer group
func (cg *KafkaConsumerGroup) run() {
	for {
		select {
		case <-cg.ctx.Done():
			cg.logger.Info("Consumer group context cancelled")
			return
		default:
			// Poll for messages
			msg, err := cg.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					// Timeout is expected, continue polling
					continue
				}
				if err.(kafka.Error).Code() == kafka.ErrPartitionEOF {
					// End of partition, continue
					continue
				}
				cg.logger.Error("Consumer group error", zap.Error(err))
				time.Sleep(1 * time.Second) // Wait before retrying
				continue
			}

			// Convert to our message format
			consumerMessage := &ConsumerMessage{
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
				consumerMessage.Headers[header.Key] = string(header.Value)
			}

			// Find appropriate handler
			cg.mu.RLock()
			handlersSnapshot := make(map[string]ConsumerHandler, len(cg.handlers))
			for id, handler := range cg.handlers {
				handlersSnapshot[id] = handler
			}
			cg.mu.RUnlock()

			handled := false
			for _, handler := range handlersSnapshot {
				topics := handler.GetTopics()
				for _, topic := range topics {
					if topic == consumerMessage.Topic {
						// Handle message
						if err := handler.Handle(cg.ctx, consumerMessage); err != nil {
							cg.logger.Error("Handler failed to process message",
								zap.String("topic", consumerMessage.Topic),
								zap.Int32("partition", consumerMessage.Partition),
								zap.Int64("offset", consumerMessage.Offset),
								zap.Error(err),
							)
						} else {
							// Commit offset if auto-commit is disabled
							if !cg.config.EnableAutoCommit {
								_, err := cg.consumer.CommitMessage(msg)
								if err != nil {
									cg.logger.Error("Failed to commit message",
										zap.String("topic", consumerMessage.Topic),
										zap.Error(err),
									)
								}
							}
						}
						handled = true
						break
					}
				}
				if handled {
					break
				}
			}

			if !handled {
				cg.logger.Warn("No handler found for message",
					zap.String("topic", consumerMessage.Topic),
				)
			}
		}
	}
}

// convertSecurityConfig converts consumer.SecurityConfig to config.SecurityConfig
func convertSecurityConfig(security *SecurityConfig) *config.SecurityConfig {
	if security == nil {
		return nil
	}

	return &config.SecurityConfig{
		Protocol:         security.Protocol,
		SASLMechanism:    security.SASLMechanism,
		SASLUsername:     security.SASLUsername,
		SASLPassword:     security.SASLPassword,
		SSLKeyLocation:   security.SSLKeyLocation,
		SSLCertLocation:  security.SSLCertLocation,
		SSLKeyPassword:   security.SSLKeyPassword,
		TruststoreLocation: "", // Not in SecurityConfig
		TruststorePassword: "", // Not in SecurityConfig
	}
}

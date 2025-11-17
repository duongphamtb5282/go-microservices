package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"auth-service/src/domain/events"
	"backend-core/logging"
	"backend-core/messaging/kafka/config"
	"backend-core/messaging/kafka/producer"
	sharedEvents "backend-shared/events"

	"github.com/segmentio/kafka-go"
)

// KafkaEventBus implements the EventBus interface using backend-core with fallback
type KafkaEventBus struct {
	producer producer.Producer
	writer   *kafka.Writer // Fallback writer
	logger   *logging.Logger
}

// NewKafkaEventBus creates a new KafkaEventBus using backend-core with kafka-go fallback
func NewKafkaEventBus(brokers []string, logger *logging.Logger) *KafkaEventBus {
	// Create Kafka config for backend-core
	kafkaConfig := &config.KafkaConfig{
		BootstrapServers: brokers,
		ClientID:         "auth-service-producer",
		Producer: &config.ProducerConfig{
			Acks:            "all",
			Retries:         3,
			BatchSize:       16384,
			LingerMs:        10,
			CompressionType: "snappy",
		},
	}

	// Try to create producer using backend-core
	producerInstance, err := producer.NewKafkaProducer(kafkaConfig, logger)
	if err != nil {
		logger.Error("Failed to create Kafka producer with backend-core",
			"error", err,
			"brokers", fmt.Sprintf("%v", brokers),
			"client_id", "auth-service-producer",
			"error_details", err.Error())

		// Log the specific error for debugging
		logger.Warn("Backend-core Kafka producer failed, trying fallback with kafka-go",
			"error_details", err.Error())

		// Create fallback writer using kafka-go
		writer := &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        sharedEvents.EventTopics.UserEvents,
			Balancer:     &kafka.LeastBytes{},
			Async:        false, // Synchronous writes for reliability
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    1, // Send immediately
		}

		logger.Info("✅ Kafka fallback writer created successfully",
			"brokers", fmt.Sprintf("%v", brokers),
			"topic", sharedEvents.EventTopics.UserEvents)

		return &KafkaEventBus{
			producer: nil,
			writer:   writer,
			logger:   logger,
		}
	}

	logger.Info("✅ Kafka producer created successfully with backend-core",
		"brokers", fmt.Sprintf("%v", brokers),
		"client_id", "auth-service-producer")

	return &KafkaEventBus{
		producer: producerInstance,
		writer:   nil,
		logger:   logger,
	}
}

// Publish publishes an event to Kafka
func (b *KafkaEventBus) Publish(event interface{}) error {
	ctx := context.Background()

	switch e := event.(type) {
	case events.UserCreated:
		return b.publishUserCreated(ctx, e)
	case events.UserActivated:
		return b.publishUserActivated(ctx, e)
	default:
		b.logger.Warn("Unknown event type", "type", fmt.Sprintf("%T", event))
		return fmt.Errorf("unknown event type: %T", event)
	}
}

// publishUserCreated publishes a UserCreated event
func (b *KafkaEventBus) publishUserCreated(ctx context.Context, event events.UserCreated) error {
	// Check if either producer or writer is available
	if b.producer == nil && b.writer == nil {
		b.logger.Warn("Neither Kafka producer nor writer available, skipping event publication",
			"user_id", event.UserID(),
			"username", event.Username(),
			"email", event.Email())
		return fmt.Errorf("kafka producer and writer not available")
	}

	// Create shared event using backend-shared
	sharedEvent := sharedEvents.NewUserCreatedEvent(
		event.UserID(),
		event.Username(),
		event.Email(),
		map[string]interface{}{
			"source":  "auth-service",
			"version": "1.0.0",
		},
	)

	// Convert to JSON
	data, err := json.Marshal(sharedEvent)
	if err != nil {
		b.logger.Error("Failed to marshal user created event", "error", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Try backend-core producer first, then fallback to kafka-go writer
	if b.producer != nil {
		// Use backend-core producer
		producerMessage := &producer.ProducerMessage{
			Topic: sharedEvents.EventTopics.UserEvents, // Use shared topic constant
			Key:   []byte(event.UserID()),
			Value: data,
			Headers: map[string]string{
				"event_type": sharedEvent.EventType,
				"event_id":   sharedEvent.EventID,
				"user_id":    event.UserID(),
				"timestamp":  sharedEvent.Timestamp.Format(time.RFC3339),
			},
		}

		// Publish to Kafka using backend-core
		if err := b.producer.Send(ctx, producerMessage); err != nil {
			b.logger.Error("Failed to publish user created event to Kafka with backend-core",
				"error", err,
				"user_id", event.UserID(),
				"event_id", sharedEvent.EventID)
			return fmt.Errorf("failed to publish event with backend-core: %w", err)
		}
	} else if b.writer != nil {
		// Use kafka-go writer fallback
		message := kafka.Message{
			Key:   []byte(event.UserID()),
			Value: data,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(sharedEvent.EventType)},
				{Key: "event_id", Value: []byte(sharedEvent.EventID)},
				{Key: "user_id", Value: []byte(event.UserID())},
				{Key: "timestamp", Value: []byte(sharedEvent.Timestamp.Format(time.RFC3339))},
			},
		}

		// Publish to Kafka using kafka-go
		if err := b.writer.WriteMessages(ctx, message); err != nil {
			b.logger.Error("Failed to publish user created event to Kafka with kafka-go",
				"error", err,
				"user_id", event.UserID(),
				"event_id", sharedEvent.EventID)
			return fmt.Errorf("failed to publish event with kafka-go: %w", err)
		}
	}

	b.logger.Info("✅ User created event published to Kafka",
		"user_id", event.UserID(),
		"username", event.Username(),
		"email", event.Email(),
		"topic", sharedEvents.EventTopics.UserEvents,
		"event_id", sharedEvent.EventID)

	return nil
}

// Close closes the Kafka producer and writer
func (b *KafkaEventBus) Close() error {
	var errs []error

	if b.producer != nil {
		if err := b.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close producer: %w", err))
		}
	}

	if b.writer != nil {
		if err := b.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Kafka resources: %v", errs)
	}

	return nil
}

// publishUserActivated publishes a UserActivated event
func (b *KafkaEventBus) publishUserActivated(ctx context.Context, event events.UserActivated) error {
	// Create shared event
	sharedEvent := sharedEvents.NewDomainEvent(
		"UserActivated",
		event.UserID(),
		map[string]interface{}{
			"user_id":  event.UserID(),
			"username": event.Username(),
			"email":    event.Email(),
		},
	)

	// Convert to JSON
	data, err := json.Marshal(sharedEvent)
	if err != nil {
		b.logger.Error("Failed to marshal user activated event", "error", err)
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// TODO: Use backend-core Kafka producer
	// For now, just log the event
	b.logger.Info("User activated event published",
		"user_id", event.UserID(),
		"username", event.Username(),
		"email", event.Email(),
		"event_data", string(data))

	return nil
}

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend-core/logging"
	"backend-core/messaging/kafka/config"
	"backend-core/messaging/kafka/consumer"
	"backend-shared/events"
)

// EventHandler defines the interface for handling events
type EventHandler interface {
	HandleUserCreatedEvent(event *events.UserCreatedEvent, requestID, correlationID string) error
	HandleUserRegisteredEvent(event *events.UserRegisteredEvent, requestID, correlationID string) error
	HandleUserActivatedEvent(event *events.UserActivatedEvent, requestID, correlationID string) error
	HandleUserLoginEvent(event *events.UserLoginEvent, requestID, correlationID string) error
}

// KafkaConsumer handles Kafka message consumption using backend-core
type KafkaConsumer struct {
	consumer consumer.Consumer
	logger   *logging.Logger
}

// NewKafkaConsumer creates a new Kafka consumer using backend-core
func NewKafkaConsumer(brokers []string, groupID string, topics []string, logger *logging.Logger) (*KafkaConsumer, error) {
	// Create Kafka config for backend-core
	kafkaConfig := &config.KafkaConfig{
		BootstrapServers: brokers,
		ClientID:         "notification-service-consumer",
		Consumer: &config.ConsumerConfig{
			GroupID:           groupID,
			AutoOffsetReset:   "earliest",
			EnableAutoCommit:  true,
			SessionTimeout:    30 * time.Second,
			HeartbeatInterval: 3 * time.Second,
			MaxPollRecords:    500,
			MaxPollInterval:   5 * time.Minute, // Required field for MaxProcessingTime
			FetchMinBytes:     1,
			FetchMaxWait:      500 * time.Millisecond,
			IsolationLevel:    "read_uncommitted",
		},
	}

	// Create consumer using backend-core
	consumerInstance, err := consumer.NewKafkaConsumer(kafkaConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer: consumerInstance,
		logger:   logger,
	}, nil
}

// extractCorrelationIDs extracts request and correlation IDs from Kafka message headers
func (c *KafkaConsumer) extractCorrelationIDs(headers map[string]string) (requestID, correlationID string) {
	for key, value := range headers {
		switch key {
		case "X-Request-ID", "request_id":
			if requestID == "" {
				requestID = value
			}
		case "X-Correlation-ID", "correlation_id":
			if correlationID == "" {
				correlationID = value
			}
		}
	}

	// If correlation ID is empty, use request ID
	if correlationID == "" {
		correlationID = requestID
	}

	return requestID, correlationID
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// ConsumeMessages starts consuming messages from Kafka using backend-core
func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, topics []string, handler EventHandler) error {
	c.logger.Info("starting message consumption with backend-core",
		"topics", topics,
		"client_id", "notification-service-consumer")

	// Subscribe to topics
	if err := c.consumer.Subscribe(topics); err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stopping message consumption")
			return nil
		default:
			// Poll for messages using backend-core
			messages, err := c.consumer.Poll(ctx, 1000) // 1 second timeout
			if err != nil {
				if err == context.Canceled {
					c.logger.Info("context canceled, stopping consumption")
					return nil
				}
				c.logger.Error("failed to poll messages", "error", err)
				continue
			}

			// Process each message
			for _, message := range messages {
				if err := c.processMessage(message, handler); err != nil {
					c.logger.Error("failed to process message", "error", err)
				}
			}
		}
	}
}

// processMessage processes a single Kafka message using backend-core ConsumerMessage
func (c *KafkaConsumer) processMessage(message *consumer.ConsumerMessage, handler EventHandler) error {
	// Extract correlation IDs from message headers
	requestID, correlationID := c.extractCorrelationIDs(message.Headers)

	c.logger.Info("received message",
		"topic", message.Topic,
		"partition", message.Partition,
		"offset", message.Offset,
		"key", string(message.Key),
		"request_id", requestID,
		"correlation_id", correlationID)

	// Parse the message
	var eventData map[string]interface{}
	if err := json.Unmarshal(message.Value, &eventData); err != nil {
		c.logger.Error("failed to parse message", "error", err)
		return err
	}

	// Debug: Log the actual message structure
	c.logger.Info("parsed message data", "data", eventData)

	// Extract event type from shared events structure
	eventType, ok := eventData["event_type"].(string)
	if !ok {
		// Fallback to "type" field for backward compatibility
		if fallbackType, fallbackOk := eventData["type"].(string); fallbackOk {
			eventType = fallbackType
		} else {
			c.logger.Error("missing event_type field", "available_fields", getKeys(eventData))
			return fmt.Errorf("missing event_type field")
		}
	}

	// Route to appropriate handler based on shared event types
	switch eventType {
	case "user.created":
		return c.handleUserCreatedEvent(eventData, handler, requestID, correlationID)
	case "user.registered":
		return c.handleUserRegisteredEvent(eventData, handler, requestID, correlationID)
	case "user.activated":
		return c.handleUserActivatedEvent(eventData, handler, requestID, correlationID)
	case "user.login":
		return c.handleUserLoginEvent(eventData, handler, requestID, correlationID)
	// Backward compatibility
	case "UserCreated":
		return c.handleUserCreatedEvent(eventData, handler, requestID, correlationID)
	case "user_registered":
		return c.handleUserRegisteredEvent(eventData, handler, requestID, correlationID)
	case "user_activated":
		return c.handleUserActivatedEvent(eventData, handler, requestID, correlationID)
	case "user_login":
		return c.handleUserLoginEvent(eventData, handler, requestID, correlationID)
	default:
		c.logger.Warn("unknown event type", "event_type", eventType)
		return &EventProcessingError{Message: "Unknown event type: " + eventType}
	}
}

// handleUserCreatedEvent handles user created events
func (c *KafkaConsumer) handleUserCreatedEvent(data map[string]interface{}, handler EventHandler, requestID, correlationID string) error {
	// Handle both shared event format and legacy format
	var eventID, eventType, aggregateID, userID, username, email string
	var timestamp time.Time
	var version int = 1
	var metadata map[string]interface{}

	// Check if this is the new shared event format (nested data)
	if nestedData, ok := data["data"].(map[string]interface{}); ok {
		// This is the new format with nested data
		eventID = getStringFromMap(data, "id")
		eventType = getStringFromMap(data, "type")
		aggregateID = getStringFromMap(nestedData, "user_id")
		userID = getStringFromMap(nestedData, "user_id")
		username = getStringFromMap(nestedData, "username")
		email = getStringFromMap(nestedData, "email")

		// Parse timestamp
		if timestampStr, ok := data["timestamp"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				timestamp = parsedTime
			}
		}

		// Parse metadata
		if meta, ok := data["metadata"].(map[string]interface{}); ok {
			metadata = meta
		}

		// Parse version
		if v, ok := data["version"].(string); ok {
			if parsedVersion, err := fmt.Sscanf(v, "%d", &version); err == nil && parsedVersion > 0 {
				// version is already set
			}
		}
	} else {
		// This is the legacy format (flat structure)
		eventID = getStringFromMap(data, "event_id")
		eventType = getStringFromMap(data, "event_type")
		aggregateID = getStringFromMap(data, "aggregate_id")
		userID = getStringFromMap(data, "user_id")
		username = getStringFromMap(data, "username")
		email = getStringFromMap(data, "email")

		// Parse timestamp
		if timestampStr, ok := data["timestamp"].(string); ok {
			if parsedTime, err := time.Parse(time.RFC3339, timestampStr); err == nil {
				timestamp = parsedTime
			}
		}

		// Parse metadata
		if meta, ok := data["metadata"].(map[string]interface{}); ok {
			metadata = meta
		}

		// Parse version
		if v, ok := data["version"].(float64); ok {
			version = int(v)
		}
	}

	// Create UserCreatedEvent
	event := &events.UserCreatedEvent{
		EventID:     eventID,
		EventType:   eventType,
		AggregateID: aggregateID,
		UserID:      userID,
		Username:    username,
		Email:       email,
		Timestamp:   timestamp,
		Version:     version,
		Metadata:    metadata,
	}

	// Use current time as fallback if timestamp is zero
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	c.logger.Info("processing user created event",
		"event_id", event.EventID,
		"event_type", event.EventType,
		"user_id", event.UserID,
		"username", event.Username,
		"email", event.Email,
		"timestamp", event.Timestamp.Format(time.RFC3339),
		"version", event.Version,
		"request_id", requestID,
		"correlation_id", correlationID)

	return handler.HandleUserCreatedEvent(event, requestID, correlationID)
}

// getStringFromMap safely extracts string values from map
func getStringFromMap(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// handleUserRegisteredEvent handles user registered events
func (c *KafkaConsumer) handleUserRegisteredEvent(data map[string]interface{}, handler EventHandler, requestID, correlationID string) error {
	var event events.UserRegisteredEvent
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", data)), &event); err != nil {
		c.logger.Error("failed to parse user registered event", "error", err)
		return err
	}

	c.logger.Info("processing user registered event",
		"event_id", event.EventID,
		"user_id", event.UserID,
		"username", event.Username,
		"email", event.Email,
		"timestamp", event.Timestamp.Format(time.RFC3339))

	return handler.HandleUserRegisteredEvent(&event, requestID, correlationID)
}

// handleUserActivatedEvent handles user activated events
func (c *KafkaConsumer) handleUserActivatedEvent(data map[string]interface{}, handler EventHandler, requestID, correlationID string) error {
	var event events.UserActivatedEvent
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", data)), &event); err != nil {
		c.logger.Error("failed to parse user activated event", "error", err)
		return err
	}

	c.logger.Info("processing user activated event",
		"event_id", event.EventID,
		"user_id", event.UserID,
		"username", event.Username,
		"email", event.Email,
		"timestamp", event.Timestamp.Format(time.RFC3339))

	return handler.HandleUserActivatedEvent(&event, requestID, correlationID)
}

// handleUserLoginEvent handles user login events
func (c *KafkaConsumer) handleUserLoginEvent(data map[string]interface{}, handler EventHandler, requestID, correlationID string) error {
	var event events.UserLoginEvent
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", data)), &event); err != nil {
		c.logger.Error("failed to parse user login event", "error", err)
		return err
	}

	c.logger.Info("processing user login event",
		"event_id", event.EventID,
		"user_id", event.UserID,
		"username", event.Username,
		"email", event.Email,
		"ip_address", event.IPAddress,
		"user_agent", event.UserAgent,
		"timestamp", event.Timestamp.Format(time.RFC3339))

	return handler.HandleUserLoginEvent(&event, requestID, correlationID)
}

// EventProcessingError represents an error in event processing
type EventProcessingError struct {
	Message string
}

func (e *EventProcessingError) Error() string {
	return e.Message
}

// getKeys returns the keys of a map
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

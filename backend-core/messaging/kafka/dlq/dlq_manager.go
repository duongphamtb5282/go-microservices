package dlq

import (
	"context"
	"fmt"
	"time"

	"backend-core/logging"
	"backend-core/messaging/kafka/producer"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// DLQMessage represents a message in the DLQ
type DLQMessage struct {
	OriginalTopic     string            `json:"original_topic"`
	OriginalPartition int32             `json:"original_partition"`
	OriginalOffset    int64             `json:"original_offset"`
	OriginalKey       []byte            `json:"original_key"`
	OriginalValue     []byte            `json:"original_value"`
	OriginalHeaders   map[string]string `json:"original_headers"`
	Error             string            `json:"error"`
	Reason            string            `json:"reason"`
	FailedAt          time.Time         `json:"failed_at"`
	RetryCount        int               `json:"retry_count"`
}

// DLQManager manages dead letter queue operations
type DLQManager struct {
	consumer *kafka.Consumer
	producer producer.Producer
	dlqTopic string
	logger   *logging.Logger
}

// NewDLQManager creates a new DLQ manager
func NewDLQManager(
	consumer *kafka.Consumer,
	producer producer.Producer,
	dlqTopic string,
	logger *logging.Logger,
) *DLQManager {
	return &DLQManager{
		consumer: consumer,
		producer: producer,
		dlqTopic: dlqTopic,
		logger:   logger,
	}
}

// Monitor monitors the DLQ and sends alerts
func (m *DLQManager) Monitor(ctx context.Context) error {
	// Subscribe to DLQ topic
	err := m.consumer.SubscribeTopics([]string{m.dlqTopic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to DLQ topic: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := m.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				m.logger.Error("DLQ consumer error", logging.Error(err))
				continue
			}

			m.handleDLQMessage(msg)
		}
	}
}

// handleDLQMessage processes DLQ messages for alerting/reporting
func (m *DLQManager) handleDLQMessage(message *kafka.Message) {
	dlqMsg := m.parseDLQMessage(message)

	// Log DLQ message
	m.logger.Error("Message in DLQ",
		logging.String("original_topic", dlqMsg.OriginalTopic),
		logging.Int64("original_offset", dlqMsg.OriginalOffset),
		logging.String("reason", dlqMsg.Reason),
		logging.String("error", dlqMsg.Error))

	// Send alert (integrate with alerting system)
	// m.alertManager.SendAlert(...)

	// Store in database for analysis
	// m.dlqRepository.Save(dlqMsg)
}

// parseDLQMessage parses DLQ message from headers
func (m *DLQManager) parseDLQMessage(message *kafka.Message) *DLQMessage {
	dlqMsg := &DLQMessage{
		OriginalKey:   message.Key,
		OriginalValue: message.Value,
		FailedAt:      time.Now(),
	}

	headers := make(map[string]string)
	for _, header := range message.Headers {
		key := header.Key
		value := string(header.Value)
		headers[key] = value

		switch key {
		case "original-topic":
			dlqMsg.OriginalTopic = value
		case "original-partition":
			fmt.Sscanf(value, "%d", &dlqMsg.OriginalPartition)
		case "original-offset":
			fmt.Sscanf(value, "%d", &dlqMsg.OriginalOffset)
		case "dlq-reason":
			dlqMsg.Reason = value
		case "error":
			dlqMsg.Error = value
		case "retry-attempt":
			fmt.Sscanf(value, "%d", &dlqMsg.RetryCount)
		}
	}

	dlqMsg.OriginalHeaders = headers
	return dlqMsg
}

// Replay replays messages from DLQ back to original topic
func (m *DLQManager) Replay(ctx context.Context, filter func(*DLQMessage) bool) error {
	// Subscribe to DLQ topic from beginning
	err := m.consumer.SubscribeTopics([]string{m.dlqTopic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to DLQ topic: %w", err)
	}

	replayCount := 0

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("DLQ replay completed",
				logging.Int("replayed_count", replayCount))
			return ctx.Err()
		default:
			msg, err := m.consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				if err.(kafka.Error).Code() == kafka.ErrPartitionEOF {
					// End of partition, we're done
					m.logger.Info("DLQ replay completed",
						logging.Int("replayed_count", replayCount))
					return nil
				}
				m.logger.Error("Failed to read DLQ message", logging.Error(err))
				continue
			}

			dlqMsg := m.parseDLQMessage(msg)

			// Apply filter
			if filter != nil && !filter(dlqMsg) {
				continue
			}

			// Replay message
			if err := m.replayMessage(dlqMsg); err != nil {
				m.logger.Error("Failed to replay message",
					logging.String("original_topic", dlqMsg.OriginalTopic),
					logging.Error(err))
				continue
			}

			replayCount++
			m.logger.Info("Message replayed",
				logging.String("original_topic", dlqMsg.OriginalTopic),
				logging.Int64("original_offset", dlqMsg.OriginalOffset))
		}
	}
}

// replayMessage sends a DLQ message back to its original topic
func (m *DLQManager) replayMessage(dlqMsg *DLQMessage) error {
	headers := make(map[string]string)
	for k, v := range dlqMsg.OriginalHeaders {
		headers[k] = v
	}

	// Add replay metadata
	headers["replayed-from-dlq"] = "true"
	headers["replayed-at"] = time.Now().Format(time.RFC3339)

	// Create producer message
	producerMsg := &producer.ProducerMessage{
		Topic:   dlqMsg.OriginalTopic,
		Key:     dlqMsg.OriginalKey,
		Value:   dlqMsg.OriginalValue,
		Headers: headers,
	}

	// Send message
	err := m.producer.Send(context.Background(), producerMsg)
	if err != nil {
		return fmt.Errorf("failed to send replayed message: %w", err)
	}

	return nil
}

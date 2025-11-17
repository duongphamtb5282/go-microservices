package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"backend-core/logging"
	"backend-core/messaging/kafka/consumer"
	"backend-core/messaging/kafka/producer"
)

// ConsumerRetryConfig holds consumer retry configuration
type ConsumerRetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int

	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration

	// BackoffMultiplier is the backoff multiplier
	BackoffMultiplier float64

	// Jitter adds randomization (0.0 to 1.0)
	Jitter float64

	// EnableDLQ enables dead letter queue
	EnableDLQ bool

	// DLQTopic is the dead letter queue topic name
	DLQTopic string

	// RetryTopicPrefix is the prefix for retry topics
	// Creates topics like: {prefix}.retry.{attempt}.{original-topic}
	RetryTopicPrefix string
}

// DefaultConsumerRetryConfig returns sensible defaults
func DefaultConsumerRetryConfig() *ConsumerRetryConfig {
	return &ConsumerRetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        60 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            0.2,
		EnableDLQ:         true,
		DLQTopic:          "dlq",
		RetryTopicPrefix:  "retry",
	}
}

// RetryableMessageHandler wraps a message handler with retry logic
type RetryableMessageHandler struct {
	handler  MessageHandler
	config   *ConsumerRetryConfig
	producer producer.Producer
	logger   *logging.Logger
}

// MessageHandler is the function that processes messages
type MessageHandler func(ctx context.Context, message *consumer.ConsumerMessage) error

// NewRetryableMessageHandler creates a handler with retry logic
func NewRetryableMessageHandler(
	handler MessageHandler,
	config *ConsumerRetryConfig,
	producer producer.Producer,
	logger *logging.Logger,
) *RetryableMessageHandler {
	if config == nil {
		config = DefaultConsumerRetryConfig()
	}

	return &RetryableMessageHandler{
		handler:  handler,
		config:   config,
		producer: producer,
		logger:   logger,
	}
}

// Handle processes a message with retry logic
func (h *RetryableMessageHandler) Handle(ctx context.Context, message *consumer.ConsumerMessage) error {
	// Get retry attempt from headers
	attempt := h.getRetryAttempt(message)

	h.logger.Info("Processing message",
		logging.String("topic", message.Topic),
		logging.Int64("offset", message.Offset),
		logging.Int("attempt", attempt))

	// Try to process the message
	err := h.handler(ctx, message)
	if err == nil {
		h.logger.Info("Message processed successfully",
			logging.String("topic", message.Topic),
			logging.Int64("offset", message.Offset))
		return nil
	}

	// Check if error is retryable
	if !IsRetryable(err) {
		h.logger.Error("Non-retryable error, sending to DLQ",
			logging.String("topic", message.Topic),
			logging.Int64("offset", message.Offset),
			logging.Error(err))
		return h.sendToDLQ(message, err, "non_retryable")
	}

	// Check if max attempts exceeded
	if attempt >= h.config.MaxAttempts {
		h.logger.Error("Max retry attempts exceeded, sending to DLQ",
			logging.String("topic", message.Topic),
			logging.Int64("offset", message.Offset),
			logging.Int("attempts", attempt),
			logging.Error(err))
		return h.sendToDLQ(message, err, "max_retries_exceeded")
	}

	// Schedule retry
	h.logger.Warn("Processing failed, scheduling retry",
		logging.String("topic", message.Topic),
		logging.Int64("offset", message.Offset),
		logging.Int("attempt", attempt),
		logging.Error(err))

	return h.scheduleRetry(message, attempt+1, err)
}

// getRetryAttempt extracts retry attempt from message headers
func (h *RetryableMessageHandler) getRetryAttempt(message *consumer.ConsumerMessage) int {
	if attemptStr, ok := message.Headers["retry-attempt"]; ok {
		var attempt int
		fmt.Sscanf(attemptStr, "%d", &attempt)
		return attempt
	}
	return 0
}

// scheduleRetry sends message to retry topic with backoff
func (h *RetryableMessageHandler) scheduleRetry(
	message *consumer.ConsumerMessage,
	attempt int,
	originalError error,
) error {
	// Calculate backoff
	backoff := calculateBackoff(attempt, h.config)

	// Create retry topic name
	retryTopic := fmt.Sprintf("%s.retry.%d.%s",
		h.config.RetryTopicPrefix,
		attempt,
		message.Topic)

	// Add metadata headers
	headers := make(map[string]string)
	for k, v := range message.Headers {
		headers[k] = v
	}
	headers["retry-attempt"] = fmt.Sprintf("%d", attempt)
	headers["original-topic"] = message.Topic
	headers["scheduled-at"] = time.Now().Format(time.RFC3339)
	headers["retry-after"] = backoff.String()
	headers["last-error"] = originalError.Error()

	// Create producer message
	producerMsg := &producer.ProducerMessage{
		Topic:   retryTopic,
		Key:     message.Key,
		Value:   message.Value,
		Headers: headers,
	}

	// Send to retry topic
	err := h.producer.Send(context.Background(), producerMsg)
	if err != nil {
		h.logger.Error("Failed to send to retry topic",
			logging.String("retry_topic", retryTopic),
			logging.Error(err))
		return err
	}

	h.logger.Info("Message scheduled for retry",
		logging.String("retry_topic", retryTopic),
		logging.Int("attempt", attempt),
		logging.Duration("backoff", backoff))

	return nil
}

// sendToDLQ sends message to dead letter queue
func (h *RetryableMessageHandler) sendToDLQ(
	message *consumer.ConsumerMessage,
	originalError error,
	reason string,
) error {
	if !h.config.EnableDLQ {
		return fmt.Errorf("DLQ disabled, dropping message: %v", originalError)
	}

	// Add DLQ metadata
	headers := make(map[string]string)
	for k, v := range message.Headers {
		headers[k] = v
	}
	headers["dlq-reason"] = reason
	headers["dlq-timestamp"] = time.Now().Format(time.RFC3339)
	headers["original-topic"] = message.Topic
	headers["original-partition"] = fmt.Sprintf("%d", message.Partition)
	headers["original-offset"] = fmt.Sprintf("%d", message.Offset)
	headers["error"] = originalError.Error()

	// Create producer message
	producerMsg := &producer.ProducerMessage{
		Topic:   h.config.DLQTopic,
		Key:     message.Key,
		Value:   message.Value,
		Headers: headers,
	}

	// Send to DLQ
	err := h.producer.Send(context.Background(), producerMsg)
	if err != nil {
		h.logger.Error("Failed to send to DLQ",
			logging.String("dlq_topic", h.config.DLQTopic),
			logging.Error(err))
		return err
	}

	h.logger.Warn("Message sent to DLQ",
		logging.String("dlq_topic", h.config.DLQTopic),
		logging.String("reason", reason),
		logging.String("original_topic", message.Topic))

	return nil
}

// calculateBackoff calculates backoff with jitter
func calculateBackoff(attempt int, config *ConsumerRetryConfig) time.Duration {
	// Exponential backoff
	backoff := float64(config.InitialBackoff) *
		math.Pow(config.BackoffMultiplier, float64(attempt-1))

	// Cap at max
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	// Add jitter
	if config.Jitter > 0 {
		jitterRange := backoff * config.Jitter
		jitter := rand.Float64()*jitterRange*2 - jitterRange
		backoff += jitter

		if backoff < 0 {
			backoff = 0
		}
	}

	return time.Duration(backoff)
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	// List of non-retryable errors
	nonRetryable := []string{
		"invalid",
		"malformed",
		"unauthorized",
		"forbidden",
		"not found",
		"duplicate",
		"validation",
	}

	errMsg := strings.ToLower(err.Error())
	for _, pattern := range nonRetryable {
		if strings.Contains(errMsg, pattern) {
			return false
		}
	}

	return true
}

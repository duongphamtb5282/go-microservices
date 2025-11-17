package producer

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"backend-core/logging"
	"backend-core/messaging/kafka/config"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

// KafkaProducer implements the Producer interface using Confluent Kafka
type KafkaProducer struct {
	producer *kafka.Producer
	config   *config.ProducerConfig
	logger   *logging.Logger
	mu       sync.RWMutex
	closed   bool
	delivery chan kafka.Event
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg *config.KafkaConfig, logger *logging.Logger) (Producer, error) {
	if cfg.Producer == nil {
		return nil, fmt.Errorf("producer configuration is required")
	}

	// Create Confluent Kafka configuration map
	configMap := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.BootstrapServers, ","),
		"client.id":         cfg.ClientID,
	}

	// Set required acks
	switch cfg.Producer.Acks {
	case "0":
		configMap["acks"] = 0
	case "1":
		configMap["acks"] = 1
	case "all", "-1":
		configMap["acks"] = "all"
	default:
		configMap["acks"] = "all"
	}

	// Set retries
	configMap["retries"] = cfg.Producer.Retries

	// Set compression
	switch cfg.Producer.CompressionType {
	case "none":
		configMap["compression.type"] = "none"
	case "gzip":
		configMap["compression.type"] = "gzip"
	case "snappy":
		configMap["compression.type"] = "snappy"
	case "lz4":
		configMap["compression.type"] = "lz4"
	case "zstd":
		configMap["compression.type"] = "zstd"
	default:
		configMap["compression.type"] = "snappy"
	}

	// Set batching
	configMap["linger.ms"] = cfg.Producer.LingerMs
	configMap["batch.size"] = cfg.Producer.BatchSize
	configMap["message.max.bytes"] = cfg.Producer.MaxRequestSize

	// Set idempotence
	if cfg.Producer.EnableIdempotent {
		configMap["enable.idempotence"] = true
	}

	// Configure security
	if err := configureSecurity(&configMap, cfg.Security); err != nil {
		return nil, fmt.Errorf("failed to configure security: %w", err)
	}

	// Create producer
	producer, err := kafka.NewProducer(&configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	kafkaProducer := &KafkaProducer{
		producer: producer,
		config:   cfg.Producer,
		logger:   logger,
		delivery: make(chan kafka.Event, 100),
	}

	// Start delivery report handler
	go kafkaProducer.handleDeliveryReports()

	return kafkaProducer, nil
}

// Send sends a message to Kafka
func (p *KafkaProducer) Send(ctx context.Context, message *ProducerMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	// Create Kafka message
	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &message.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:   message.Key,
		Value: message.Value,
	}

	// Add headers
	if len(message.Headers) > 0 {
		headers := make([]kafka.Header, 0, len(message.Headers))
		for key, value := range message.Headers {
			headers = append(headers, kafka.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
		kafkaMessage.Headers = headers
	}

	// Send message
	err := p.producer.Produce(kafkaMessage, p.delivery)
	if err != nil {
		p.logger.Error("Failed to produce message",
			zap.String("topic", message.Topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to produce message: %w", err)
	}

	// Wait for delivery report (synchronous behavior)
	select {
	case e := <-p.delivery:
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				p.logger.Error("Message delivery failed",
					zap.String("topic", *ev.TopicPartition.Topic),
					zap.Int32("partition", ev.TopicPartition.Partition),
					zap.Int64("offset", int64(ev.TopicPartition.Offset)),
					zap.Error(ev.TopicPartition.Error),
				)
				return fmt.Errorf("message delivery failed: %w", ev.TopicPartition.Error)
			}
			p.logger.Debug("Message sent successfully",
				zap.String("topic", *ev.TopicPartition.Topic),
				zap.Int32("partition", ev.TopicPartition.Partition),
				zap.Int64("offset", int64(ev.TopicPartition.Offset)),
			)
		case kafka.Error:
			p.logger.Error("Producer error", zap.Error(ev))
			return fmt.Errorf("producer error: %w", ev)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

// SendBatch sends multiple messages to Kafka
func (p *KafkaProducer) SendBatch(ctx context.Context, messages []*ProducerMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	if len(messages) == 0 {
		return nil
	}

	// Send all messages
	for _, message := range messages {
		kafkaMessage := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &message.Topic,
				Partition: kafka.PartitionAny,
			},
			Key:   message.Key,
			Value: message.Value,
		}

		// Add headers
		if len(message.Headers) > 0 {
			headers := make([]kafka.Header, 0, len(message.Headers))
			for key, value := range message.Headers {
				headers = append(headers, kafka.Header{
					Key:   key,
					Value: []byte(value),
				})
			}
			kafkaMessage.Headers = headers
		}

		if err := p.producer.Produce(kafkaMessage, p.delivery); err != nil {
			p.logger.Error("Failed to produce batch message",
				zap.String("topic", message.Topic),
				zap.Error(err),
			)
			return fmt.Errorf("failed to produce batch message: %w", err)
		}
	}

	// Wait for all messages to be delivered
	delivered := 0
	for delivered < len(messages) {
		select {
		case e := <-p.delivery:
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.logger.Error("Batch message delivery failed",
						zap.String("topic", *ev.TopicPartition.Topic),
						zap.Error(ev.TopicPartition.Error),
					)
					return fmt.Errorf("batch message delivery failed: %w", ev.TopicPartition.Error)
				}
				delivered++
			case kafka.Error:
				p.logger.Error("Producer error during batch", zap.Error(ev))
				return fmt.Errorf("producer error: %w", ev)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	p.logger.Debug("Batch messages sent successfully",
		zap.Int("count", len(messages)),
	)

	return nil
}

// SendAsync sends a message asynchronously
func (p *KafkaProducer) SendAsync(ctx context.Context, message *ProducerMessage) error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &message.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:   message.Key,
		Value: message.Value,
	}

	// Add headers
	if len(message.Headers) > 0 {
		headers := make([]kafka.Header, 0, len(message.Headers))
		for key, value := range message.Headers {
			headers = append(headers, kafka.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
		kafkaMessage.Headers = headers
	}

	// Use nil delivery channel for truly async
	err := p.producer.Produce(kafkaMessage, nil)
	if err != nil {
		p.logger.Error("Failed to produce async message",
			zap.String("topic", message.Topic),
			zap.Error(err),
		)
		return fmt.Errorf("failed to produce async message: %w", err)
	}

	p.logger.Debug("Async message queued",
		zap.String("topic", message.Topic),
	)

	return nil
}

// Flush flushes any pending messages
func (p *KafkaProducer) Flush() error {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		return fmt.Errorf("producer is closed")
	}
	p.mu.RUnlock()

	remaining := p.producer.Flush(15 * 1000) // 15 seconds timeout
	if remaining > 0 {
		p.logger.Warn("Flush timeout, some messages may not be delivered",
			zap.Int("remaining", remaining),
		)
	}

	p.logger.Debug("Flush completed")
	return nil
}

// Close closes the producer
func (p *KafkaProducer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	// Flush remaining messages
	remaining := p.producer.Flush(15 * 1000)
	if remaining > 0 {
		p.logger.Warn("Some messages not flushed during close",
			zap.Int("remaining", remaining),
		)
	}

	// Close producer
	p.producer.Close()
	close(p.delivery)

	p.closed = true
	p.logger.Info("Producer closed successfully")
	return nil
}

// handleDeliveryReports handles delivery reports from the producer
func (p *KafkaProducer) handleDeliveryReports() {
	for {
		select {
		case e, ok := <-p.delivery:
			if !ok {
				return
			}
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.logger.Error("Message delivery failed",
						zap.String("topic", *ev.TopicPartition.Topic),
						zap.Error(ev.TopicPartition.Error),
					)
				} else {
					p.logger.Debug("Message delivered successfully",
						zap.String("topic", *ev.TopicPartition.Topic),
						zap.Int32("partition", ev.TopicPartition.Partition),
						zap.Int64("offset", int64(ev.TopicPartition.Offset)),
					)
				}
			case kafka.Error:
				p.logger.Error("Producer error", zap.Error(ev))
			}
		}
	}
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

package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"backend-core/logging"
	"backend-core/messaging/kafka/config"
	"backend-core/messaging/kafka/consumer"
	"backend-core/messaging/kafka/management"
	"backend-core/messaging/kafka/producer"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

// KafkaManager implements the KafkaManager interface
type KafkaManager struct {
	config         *config.KafkaConfig
	logger         *logging.Logger
	adminClient    *kafka.AdminClient
	producers      map[string]producer.Producer
	consumers      map[string]consumer.Consumer
	consumerGroups map[string]consumer.ConsumerGroup
	mu             sync.RWMutex
	closed         bool
}

// NewKafkaManager creates a new Kafka manager
func NewKafkaManager(cfg *config.KafkaConfig, logger *logging.Logger) (management.KafkaManager, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create admin client configuration
	adminConfigMap := kafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.BootstrapServers, ","),
		"client.id":         cfg.ClientID,
	}

	// Configure security
	if err := configureSecurity(&adminConfigMap, cfg.Security); err != nil {
		return nil, fmt.Errorf("failed to configure security: %w", err)
	}

	// Create admin client
	adminClient, err := kafka.NewAdminClient(&adminConfigMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	manager := &KafkaManager{
		config:         cfg,
		logger:         logger,
		adminClient:    adminClient,
		producers:      make(map[string]producer.Producer),
		consumers:      make(map[string]consumer.Consumer),
		consumerGroups: make(map[string]consumer.ConsumerGroup),
	}

	return manager, nil
}

// CreateProducer creates a new producer
func (m *KafkaManager) CreateProducer(producerConfig *producer.Config) (producer.Producer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	// Create a new config with producer settings
	cfg := &config.KafkaConfig{
		BootstrapServers: producerConfig.BootstrapServers,
		ClientID:         producerConfig.ClientID,
		Security:         convertProducerSecurityConfig(producerConfig.SecurityConfig),
		Producer: &config.ProducerConfig{
			Acks:            producerConfig.Acks,
			Retries:         producerConfig.Retries,
			BatchSize:       producerConfig.BatchSize,
			LingerMs:        producerConfig.LingerMs,
			CompressionType: producerConfig.CompressionType,
			MaxRequestSize:  1048576,          // 1MB
			DeliveryTimeout: 30 * time.Second, // Set delivery timeout
		},
		Consumer: &config.ConsumerConfig{
			GroupID:           "auth-service-group",
			AutoOffsetReset:   "earliest",
			EnableAutoCommit:  true,
			SessionTimeout:    30 * time.Second,
			HeartbeatInterval: 10 * time.Second,
			MaxPollRecords:    100,
			MaxPollInterval:   5 * time.Minute,
			FetchMinBytes:     1,
			FetchMaxWait:      500 * time.Millisecond,
			IsolationLevel:    "read_uncommitted",
		},
	}

	producer, err := producer.NewKafkaProducer(cfg, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	producerID := fmt.Sprintf("producer_%d", time.Now().UnixNano())
	m.producers[producerID] = producer

	m.logger.Info("Producer created", zap.String("producer_id", producerID))
	return producer, nil
}

// CreateConsumer creates a new consumer
func (m *KafkaManager) CreateConsumer(consumerConfig *consumer.Config) (consumer.Consumer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	// Create a new config with consumer settings
	cfg := &config.KafkaConfig{
		BootstrapServers: consumerConfig.BootstrapServers,
		ClientID:         consumerConfig.ClientID,
		Security:         convertConsumerSecurityConfig(consumerConfig.SecurityConfig),
		Consumer: &config.ConsumerConfig{
			GroupID:           consumerConfig.GroupID,
			AutoOffsetReset:   consumerConfig.AutoOffsetReset,
			EnableAutoCommit:  consumerConfig.EnableAutoCommit,
			SessionTimeout:    consumerConfig.SessionTimeout,
			HeartbeatInterval: 10 * time.Second,
			MaxPollRecords:    100,
			MaxPollInterval:   5 * time.Minute,
			FetchMinBytes:     1,
			FetchMaxWait:      500 * time.Millisecond,
			IsolationLevel:    "read_uncommitted",
		},
	}

	kafkaConsumer, err := consumer.NewKafkaConsumer(cfg, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	consumerID := fmt.Sprintf("consumer_%d", time.Now().UnixNano())
	m.consumers[consumerID] = kafkaConsumer

	m.logger.Info("Consumer created", zap.String("consumer_id", consumerID))
	return kafkaConsumer, nil
}

// CreateConsumerGroup creates a consumer group
func (m *KafkaManager) CreateConsumerGroup(config *management.ConsumerGroupConfig) (consumer.ConsumerGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	// Convert management config to consumer config
	consumerConfig := convertConsumerGroupConfig(config)

	// Create consumer group implementation
	consumerGroup, err := consumer.NewKafkaConsumerGroup(consumerConfig, m.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	groupID := fmt.Sprintf("group_%d", time.Now().UnixNano())
	m.consumerGroups[groupID] = consumerGroup

	m.logger.Info("Consumer group created", zap.String("group_id", groupID))
	return consumerGroup, nil
}

// GetClusterInfo returns cluster information
func (m *KafkaManager) GetClusterInfo() (*management.ClusterInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	// Get cluster metadata
	metadata, err := m.adminClient.GetMetadata(nil, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	// Extract broker information
	brokerAddresses := make([]string, 0, len(metadata.Brokers))
	for _, broker := range metadata.Brokers {
		brokerAddresses = append(brokerAddresses, fmt.Sprintf("%s:%d", broker.Host, broker.Port))
	}

	// Extract topics
	topics := make([]string, 0, len(metadata.Topics))
	totalPartitions := 0
	totalReplicas := 0

	for topicName, topicMetadata := range metadata.Topics {
		topics = append(topics, topicName)
		for _, partitionMetadata := range topicMetadata.Partitions {
			totalPartitions++
			totalReplicas += len(partitionMetadata.Replicas)
		}
	}

	clusterInfo := &management.ClusterInfo{
		Brokers:    brokerAddresses,
		Controller: "", // Would need additional logic to determine controller
		Topics:     topics,
		Partitions: totalPartitions,
		Replicas:   totalReplicas,
	}

	return clusterInfo, nil
}

// CreateTopic creates a new topic
func (m *KafkaManager) CreateTopic(topicConfig *management.TopicConfig) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return fmt.Errorf("manager is closed")
	}

	// Convert config map to the required format
	configEntries := make(map[string]string)
	for k, v := range topicConfig.Config {
		configEntries[k] = v
	}

	// Create topic specification
	topicSpec := kafka.TopicSpecification{
		Topic:             topicConfig.Name,
		NumPartitions:     topicConfig.Partitions,
		ReplicationFactor: topicConfig.ReplicationFactor,
		Config:            configEntries,
	}

	// Create topic
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := m.adminClient.CreateTopics(ctx, []kafka.TopicSpecification{topicSpec})
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	// Check result
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to create topic %s: %w", result.Topic, result.Error)
		}
	}

	m.logger.Info("Topic created",
		zap.String("topic", topicConfig.Name),
		zap.Int("partitions", topicConfig.Partitions),
		zap.Int("replication_factor", topicConfig.ReplicationFactor),
	)

	return nil
}

// DeleteTopic deletes a topic
func (m *KafkaManager) DeleteTopic(topic string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return fmt.Errorf("manager is closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := m.adminClient.DeleteTopics(ctx, []string{topic})
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	// Check result
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError {
			return fmt.Errorf("failed to delete topic %s: %w", result.Topic, result.Error)
		}
	}

	m.logger.Info("Topic deleted", zap.String("topic", topic))
	return nil
}

// ListTopics lists all topics
func (m *KafkaManager) ListTopics() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	metadata, err := m.adminClient.GetMetadata(nil, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	topics := make([]string, 0, len(metadata.Topics))
	for topicName := range metadata.Topics {
		topics = append(topics, topicName)
	}

	return topics, nil
}

// GetTopicMetadata returns topic metadata
func (m *KafkaManager) GetTopicMetadata(topic string) (*management.TopicMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.closed {
		return nil, fmt.Errorf("manager is closed")
	}

	metadata, err := m.adminClient.GetMetadata(&topic, true, 5000)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	topicMetadata, ok := metadata.Topics[topic]
	if !ok {
		return nil, fmt.Errorf("topic %s not found", topic)
	}

	partitionInfos := make([]management.PartitionInfo, 0, len(topicMetadata.Partitions))
	for partitionID, partitionMetadata := range topicMetadata.Partitions {
		partitionIDInt32 := int32(partitionID)
		partitionInfo := management.PartitionInfo{
			ID:       partitionIDInt32,
			Leader:   partitionMetadata.Leader,
			Replicas: partitionMetadata.Replicas,
			ISR:      partitionMetadata.Isrs, // ISR is available in confluent-kafka-go
			Offsets:  make(map[string]int64),
		}
		partitionInfos = append(partitionInfos, partitionInfo)
	}

	metadataResult := &management.TopicMetadata{
		Name:       topic,
		Partitions: partitionInfos,
		Config:     make(map[string]string), // Would need additional call to get config
	}

	return metadataResult, nil
}

// Close closes the manager and all its resources
func (m *KafkaManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}

	var errs []error

	// Close all producers
	for id, producer := range m.producers {
		if err := producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close producer %s: %w", id, err))
		}
	}

	// Close all consumers
	for id, consumer := range m.consumers {
		if err := consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close consumer %s: %w", id, err))
		}
	}

	// Close all consumer groups
	for id, group := range m.consumerGroups {
		if err := group.Stop(); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop consumer group %s: %w", id, err))
		}
	}

	// Close admin client
	m.adminClient.Close()

	m.closed = true

	if len(errs) > 0 {
		return fmt.Errorf("errors closing manager: %v", errs)
	}

	m.logger.Info("Kafka manager closed successfully")
	return nil
}

// Helper functions

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

func convertProducerSecurityConfig(security *producer.SecurityConfig) *config.SecurityConfig {
	if security == nil {
		return nil
	}

	return &config.SecurityConfig{
		Protocol:         security.SecurityProtocol,
		SASLMechanism:    security.SASLMechanism,
		SASLUsername:     security.SASLUsername,
		SASLPassword:     security.SASLPassword,
		SSLKeyLocation:   security.SSLKeyLocation,
		SSLCertLocation:  security.SSLCertLocation,
		SSLKeyPassword:   security.SSLKeyPassword,
		TruststoreLocation: "",
		TruststorePassword: "",
	}
}

func convertConsumerSecurityConfig(security *consumer.SecurityConfig) *config.SecurityConfig {
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
		TruststoreLocation: "",
		TruststorePassword: "",
	}
}

func convertConsumerGroupConfig(config *management.ConsumerGroupConfig) *consumer.ConsumerGroupConfig {
	if config == nil {
		return nil
	}

	return &consumer.ConsumerGroupConfig{
		GroupID:          config.GroupID,
		BootstrapServers: config.BootstrapServers,
		Topics:           config.Topics,
		Config:           config.Config,
		ClientID:         "kafka-manager",
		AutoOffsetReset:  "latest",
		EnableAutoCommit: true,
		SessionTimeout:   30 * time.Second,
		SecurityConfig:   nil, // Will be set if needed
	}
}

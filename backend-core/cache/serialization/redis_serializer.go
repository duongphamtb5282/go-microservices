package serialization

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gomodule/redigo/redis"
)

// RedisSerializer provides serialization/deserialization for Redis
type RedisSerializer interface {
	Serialize(key string, value interface{}) (string, error)
	Deserialize(key string, data string, target interface{}) error
	GetSerializerType() string
}

// JSONSerializer implements JSON-based serialization
type JSONSerializer struct {
	prefix string
}

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer(prefix string) *JSONSerializer {
	return &JSONSerializer{
		prefix: prefix,
	}
}

// Serialize converts a Go value to JSON string for Redis storage
func (j *JSONSerializer) Serialize(key string, value interface{}) (string, error) {
	// Add metadata for deserialization
	serializedData := SerializedData{
		Type:      reflect.TypeOf(value).String(),
		Timestamp: time.Now().Unix(),
		Data:      value,
	}

	jsonData, err := json.Marshal(serializedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	return string(jsonData), nil
}

// Deserialize converts JSON string from Redis back to Go value
func (j *JSONSerializer) Deserialize(key string, data string, target interface{}) error {
	var serializedData SerializedData
	if err := json.Unmarshal([]byte(data), &serializedData); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Convert the data to the target type
	jsonData, err := json.Marshal(serializedData.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data for conversion: %w", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target type: %w", err)
	}

	return nil
}

// GetSerializerType returns the serializer type
func (j *JSONSerializer) GetSerializerType() string {
	return "json"
}

// SerializedData represents the structure of serialized data in Redis
type SerializedData struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// BinarySerializer implements binary serialization using gob
type BinarySerializer struct {
	prefix string
}

// NewBinarySerializer creates a new binary serializer
func NewBinarySerializer(prefix string) *BinarySerializer {
	return &BinarySerializer{
		prefix: prefix,
	}
}

// Serialize converts a Go value to binary data for Redis storage
func (b *BinarySerializer) Serialize(key string, value interface{}) (string, error) {
	// For binary serialization, we'll use base64 encoding
	// This is a simplified implementation - in production, use gob or protobuf
	jsonData, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	// In a real implementation, you would use gob.Encode or protobuf here
	return string(jsonData), nil
}

// Deserialize converts binary data from Redis back to Go value
func (b *BinarySerializer) Deserialize(key string, data string, target interface{}) error {
	// For binary deserialization, we'll use JSON as a fallback
	// In a real implementation, you would use gob.Decode or protobuf here
	return json.Unmarshal([]byte(data), target)
}

// GetSerializerType returns the serializer type
func (b *BinarySerializer) GetSerializerType() string {
	return "binary"
}

// MessagePackSerializer implements MessagePack serialization
type MessagePackSerializer struct {
	prefix string
}

// NewMessagePackSerializer creates a new MessagePack serializer
func NewMessagePackSerializer(prefix string) *MessagePackSerializer {
	return &MessagePackSerializer{
		prefix: prefix,
	}
}

// Serialize converts a Go value to MessagePack for Redis storage
func (m *MessagePackSerializer) Serialize(key string, value interface{}) (string, error) {
	// This is a placeholder - in production, use msgpack library
	jsonData, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	return string(jsonData), nil
}

// Deserialize converts MessagePack data from Redis back to Go value
func (m *MessagePackSerializer) Deserialize(key string, data string, target interface{}) error {
	// This is a placeholder - in production, use msgpack library
	return json.Unmarshal([]byte(data), target)
}

// GetSerializerType returns the serializer type
func (m *MessagePackSerializer) GetSerializerType() string {
	return "msgpack"
}

// SerializerFactory creates serializers based on configuration
type SerializerFactory struct{}

// NewSerializerFactory creates a new serializer factory
func NewSerializerFactory() *SerializerFactory {
	return &SerializerFactory{}
}

// CreateSerializer creates a serializer based on the specified type
func (f *SerializerFactory) CreateSerializer(serializerType, prefix string) RedisSerializer {
	switch serializerType {
	case "json":
		return NewJSONSerializer(prefix)
	case "binary":
		return NewBinarySerializer(prefix)
	case "msgpack":
		return NewMessagePackSerializer(prefix)
	default:
		return NewJSONSerializer(prefix)
	}
}

// RedisSerializationManager manages serialization for Redis operations
type RedisSerializationManager struct {
	serializer RedisSerializer
	conn       redis.Conn
}

// NewRedisSerializationManager creates a new serialization manager
func NewRedisSerializationManager(serializer RedisSerializer, conn redis.Conn) *RedisSerializationManager {
	return &RedisSerializationManager{
		serializer: serializer,
		conn:       conn,
	}
}

// Set stores a value in Redis with serialization
func (r *RedisSerializationManager) Set(key string, value interface{}, expiration time.Duration) error {
	serializedData, err := r.serializer.Serialize(key, value)
	if err != nil {
		return fmt.Errorf("failed to serialize value: %w", err)
	}

	_, err = r.conn.Do("SET", key, serializedData)
	if err != nil {
		return fmt.Errorf("failed to set key in Redis: %w", err)
	}

	if expiration > 0 {
		_, err = r.conn.Do("EXPIRE", key, int(expiration.Seconds()))
		if err != nil {
			return fmt.Errorf("failed to set expiration: %w", err)
		}
	}

	return nil
}

// Get retrieves a value from Redis with deserialization
func (r *RedisSerializationManager) Get(key string, target interface{}) error {
	data, err := redis.String(r.conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return fmt.Errorf("key not found: %s", key)
		}
		return fmt.Errorf("failed to get key from Redis: %w", err)
	}

	return r.serializer.Deserialize(key, data, target)
}

// Delete removes a key from Redis
func (r *RedisSerializationManager) Delete(key string) error {
	_, err := r.conn.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("failed to delete key from Redis: %w", err)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisSerializationManager) Exists(key string) (bool, error) {
	exists, err := redis.Bool(r.conn.Do("EXISTS", key))
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}

	return exists, nil
}

// GetTTL returns the time to live for a key
func (r *RedisSerializationManager) GetTTL(key string) (time.Duration, error) {
	ttl, err := redis.Int64(r.conn.Do("TTL", key))
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	if ttl == -1 {
		return -1, nil // Key exists but has no expiration
	}

	if ttl == -2 {
		return 0, fmt.Errorf("key does not exist")
	}

	return time.Duration(ttl) * time.Second, nil
}

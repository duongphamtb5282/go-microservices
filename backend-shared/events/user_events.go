package events

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// UserCreatedEvent represents the event when a user is created
type UserCreatedEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     int                    `json:"version"`
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID, username, email string, metadata map[string]interface{}) *UserCreatedEvent {
	return &UserCreatedEvent{
		EventID:     generateEventID(),
		EventType:   "user.created",
		AggregateID: userID,
		UserID:      userID,
		Username:    username,
		Email:       email,
		Timestamp:   time.Now().UTC(),
		Metadata:    metadata,
		Version:     1,
	}
}

// UserRegisteredEvent represents the event when a user is registered
type UserRegisteredEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     int                    `json:"version"`
}

// NewUserRegisteredEvent creates a new user registered event
func NewUserRegisteredEvent(userID, username, email string, metadata map[string]interface{}) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		EventID:     generateEventID(),
		EventType:   "user.registered",
		AggregateID: userID,
		UserID:      userID,
		Username:    username,
		Email:       email,
		Timestamp:   time.Now().UTC(),
		Metadata:    metadata,
		Version:     1,
	}
}

// UserActivatedEvent represents the event when a user is activated
type UserActivatedEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     int                    `json:"version"`
}

// NewUserActivatedEvent creates a new user activated event
func NewUserActivatedEvent(userID, username, email string, metadata map[string]interface{}) *UserActivatedEvent {
	return &UserActivatedEvent{
		EventID:     generateEventID(),
		EventType:   "user.activated",
		AggregateID: userID,
		UserID:      userID,
		Username:    username,
		Email:       email,
		Timestamp:   time.Now().UTC(),
		Metadata:    metadata,
		Version:     1,
	}
}

// UserLoginEvent represents the event when a user logs in
type UserLoginEvent struct {
	EventID     string                 `json:"event_id"`
	EventType   string                 `json:"event_type"`
	AggregateID string                 `json:"aggregate_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     int                    `json:"version"`
}

// NewUserLoginEvent creates a new user login event
func NewUserLoginEvent(userID, username, email, ipAddress, userAgent string, metadata map[string]interface{}) *UserLoginEvent {
	return &UserLoginEvent{
		EventID:     generateEventID(),
		EventType:   "user.login",
		AggregateID: userID,
		UserID:      userID,
		Username:    username,
		Email:       email,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Timestamp:   time.Now().UTC(),
		Metadata:    metadata,
		Version:     1,
	}
}

// EventTopics defines the Kafka topics for events
var EventTopics = struct {
	UserEvents string
	AuthEvents string
	AuditLogs  string
}{
	UserEvents: "user.events",
	AuthEvents: "auth.events",
	AuditLogs:  "audit.logs",
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Generate 16 random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)

	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant bits

	// Convert to hex string
	hexStr := hex.EncodeToString(bytes)

	// Format as UUID
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32])
}

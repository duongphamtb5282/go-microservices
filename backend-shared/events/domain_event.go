package events

import (
	"context"
	"time"

	"backend-shared/models"
)

// DomainEvent represents a domain event
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     int                    `json:"version"`
}

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType, aggregateID string, data map[string]interface{}) DomainEvent {
	return DomainEvent{
		ID:          models.NewEntityID().String(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        data,
		Metadata:    make(map[string]interface{}),
		Timestamp:   time.Now(),
		Version:     1,
	}
}

// WithMetadata adds metadata to the event
func (e DomainEvent) WithMetadata(key string, value interface{}) DomainEvent {
	e.Metadata[key] = value
	return e
}

// WithVersion sets the version of the event
func (e DomainEvent) WithVersion(version int) DomainEvent {
	e.Version = version
	return e
}

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
	PublishBatch(ctx context.Context, events []DomainEvent) error
}

// EventSubscriber defines the interface for subscribing to domain events
type EventSubscriber interface {
	Subscribe(eventType string, handler DomainEventHandler) error
	Unsubscribe(eventType string, handler DomainEventHandler) error
}

// DomainEventHandler defines the interface for handling domain events
type DomainEventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
	GetEventType() string
}

// DomainEventBus defines the interface for the domain event bus
type DomainEventBus interface {
	EventPublisher
	EventSubscriber
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

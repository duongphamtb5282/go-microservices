package events

import (
	"time"

	"github.com/google/uuid"
)

// Event represents a base event structure
type Event struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Version       string                 `json:"version"`
	Source        string                 `json:"source"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          interface{}            `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	CausationID   string                 `json:"causation_id,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType, source string, data interface{}) *Event {
	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Version:   "1.0",
		Source:    source,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// SetVersion sets the event version
func (e *Event) SetVersion(version string) {
	e.Version = version
}

// SetCorrelationID sets the correlation ID
func (e *Event) SetCorrelationID(correlationID string) {
	e.CorrelationID = correlationID
}

// SetCausationID sets the causation ID
func (e *Event) SetCausationID(causationID string) {
	e.CausationID = causationID
}

// AddMetadata adds metadata to the event
func (e *Event) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// GetID returns the event ID
func (e *Event) GetID() string {
	return e.ID
}

// GetType returns the event type
func (e *Event) GetType() string {
	return e.Type
}

// GetVersion returns the event version
func (e *Event) GetVersion() string {
	return e.Version
}

// GetSource returns the event source
func (e *Event) GetSource() string {
	return e.Source
}

// GetTimestamp returns the event timestamp
func (e *Event) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetData returns the event data
func (e *Event) GetData() interface{} {
	return e.Data
}

// GetMetadata returns the event metadata
func (e *Event) GetMetadata() map[string]interface{} {
	return e.Metadata
}

// GetCorrelationID returns the correlation ID
func (e *Event) GetCorrelationID() string {
	return e.CorrelationID
}

// GetCausationID returns the causation ID
func (e *Event) GetCausationID() string {
	return e.CausationID
}

// IsValid checks if the event is valid
func (e *Event) IsValid() bool {
	return e.ID != "" && e.Type != "" && e.Source != ""
}

// GetAge returns the age of the event
func (e *Event) GetAge() time.Duration {
	return time.Since(e.Timestamp)
}

// IsExpired checks if the event is expired based on TTL
func (e *Event) IsExpired(ttl time.Duration) bool {
	return e.GetAge() > ttl
}

// EventHandler represents an event handler interface
type EventHandler interface {
	// Handle handles an event
	Handle(event *Event) error

	// CanHandle checks if the handler can handle the event type
	CanHandle(eventType string) bool

	// GetHandlerName returns the handler name
	GetHandlerName() string
}

// EventBus represents an event bus interface
type EventBus interface {
	// Publish publishes an event
	Publish(event *Event) error

	// Subscribe subscribes to an event type
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe unsubscribes from an event type
	Unsubscribe(eventType string, handlerName string) error

	// GetSubscribers returns subscribers for an event type
	GetSubscribers(eventType string) []EventHandler

	// Start starts the event bus
	Start() error

	// Stop stops the event bus
	Stop() error

	// IsRunning checks if the event bus is running
	IsRunning() bool
}

// EventStore represents an event store interface
type EventStore interface {
	// Save saves an event
	Save(event *Event) error

	// Get gets an event by ID
	Get(eventID string) (*Event, error)

	// GetByType gets events by type
	GetByType(eventType string) ([]*Event, error)

	// GetBySource gets events by source
	GetBySource(source string) ([]*Event, error)

	// GetByCorrelationID gets events by correlation ID
	GetByCorrelationID(correlationID string) ([]*Event, error)

	// GetByDateRange gets events within a date range
	GetByDateRange(start, end time.Time) ([]*Event, error)

	// Delete deletes an event
	Delete(eventID string) error
}

// EventProcessor represents an event processor interface
type EventProcessor interface {
	// Process processes an event
	Process(event *Event) error

	// GetProcessorName returns the processor name
	GetProcessorName() string

	// GetSupportedEventTypes returns supported event types
	GetSupportedEventTypes() []string

	// IsHealthy checks if the processor is healthy
	IsHealthy() bool
}

// EventFilter represents an event filter
type EventFilter struct {
	EventTypes    []string               `json:"event_types,omitempty"`
	Sources       []string               `json:"sources,omitempty"`
	StartTime     *time.Time             `json:"start_time,omitempty"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewEventFilter creates a new event filter
func NewEventFilter() *EventFilter {
	return &EventFilter{
		EventTypes: make([]string, 0),
		Sources:    make([]string, 0),
		Metadata:   make(map[string]interface{}),
	}
}

// AddEventType adds an event type to the filter
func (f *EventFilter) AddEventType(eventType string) {
	f.EventTypes = append(f.EventTypes, eventType)
}

// AddSource adds a source to the filter
func (f *EventFilter) AddSource(source string) {
	f.Sources = append(f.Sources, source)
}

// SetTimeRange sets the time range for the filter
func (f *EventFilter) SetTimeRange(start, end time.Time) {
	f.StartTime = &start
	f.EndTime = &end
}

// SetCorrelationID sets the correlation ID for the filter
func (f *EventFilter) SetCorrelationID(correlationID string) {
	f.CorrelationID = correlationID
}

// AddMetadata adds metadata to the filter
func (f *EventFilter) AddMetadata(key string, value interface{}) {
	if f.Metadata == nil {
		f.Metadata = make(map[string]interface{})
	}
	f.Metadata[key] = value
}

// Matches checks if an event matches the filter
func (f *EventFilter) Matches(event *Event) bool {
	// Check event types
	if len(f.EventTypes) > 0 {
		found := false
		for _, eventType := range f.EventTypes {
			if event.Type == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check sources
	if len(f.Sources) > 0 {
		found := false
		for _, source := range f.Sources {
			if event.Source == source {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if f.StartTime != nil && event.Timestamp.Before(*f.StartTime) {
		return false
	}
	if f.EndTime != nil && event.Timestamp.After(*f.EndTime) {
		return false
	}

	// Check correlation ID
	if f.CorrelationID != "" && event.CorrelationID != f.CorrelationID {
		return false
	}

	// Check metadata
	if len(f.Metadata) > 0 {
		for key, value := range f.Metadata {
			if event.Metadata[key] != value {
				return false
			}
		}
	}

	return true
}

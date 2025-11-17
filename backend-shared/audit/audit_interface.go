package audit

import (
	"time"
)

// Auditable represents an entity that can be audited
type Auditable interface {
	// GetAuditEntity returns the audit entity
	GetAuditEntity() *AuditEntity

	// GetEntityID returns the entity ID
	GetEntityID() string

	// GetEntityType returns the entity type
	GetEntityType() string

	// UpdateAudit updates the audit fields
	UpdateAudit(modifiedBy string)

	// GetAuditInfo returns audit information
	GetAuditInfo() *AuditInfo
}

// AuditRepository represents a repository for audit operations
type AuditRepository interface {
	// SaveAuditInfo saves audit information
	SaveAuditInfo(info *AuditInfo) error

	// GetAuditTrail gets the audit trail for an entity
	GetAuditTrail(entityID, entityType string) (*AuditTrail, error)

	// GetAuditInfoByEntity gets audit info by entity
	GetAuditInfoByEntity(entityID, entityType string) ([]AuditInfo, error)

	// GetAuditInfoByUser gets audit info by user
	GetAuditInfoByUser(userID string) ([]AuditInfo, error)

	// GetAuditInfoByAction gets audit info by action
	GetAuditInfoByAction(action string) ([]AuditInfo, error)

	// GetAuditInfoByDateRange gets audit info within a date range
	GetAuditInfoByDateRange(start, end time.Time) ([]AuditInfo, error)

	// DeleteAuditInfo deletes audit information
	DeleteAuditInfo(entityID, entityType string) error
}

// AuditService represents a service for audit operations
type AuditService interface {
	// TrackCreate tracks entity creation
	TrackCreate(entity Auditable, userID string) error

	// TrackUpdate tracks entity update
	TrackUpdate(entity Auditable, userID string, changes map[string]interface{}) error

	// TrackDelete tracks entity deletion
	TrackDelete(entity Auditable, userID string) error

	// GetAuditTrail gets the audit trail for an entity
	GetAuditTrail(entityID, entityType string) (*AuditTrail, error)

	// GetAuditInfo gets audit information
	GetAuditInfo(entityID, entityType string) ([]AuditInfo, error)
}

// AuditEventHandler represents an event handler for audit events
type AuditEventHandler interface {
	// HandleAuditEvent handles audit events
	HandleAuditEvent(event *AuditEvent) error

	// CanHandle checks if the handler can handle the event
	CanHandle(eventType string) bool
}

// AuditEvent represents an audit event
type AuditEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	EntityID      string                 `json:"entity_id"`
	EntityType    string                 `json:"entity_type"`
	Action        string                 `json:"action"`
	UserID        string                 `json:"user_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Changes       []Change               `json:"changes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// NewAuditEvent creates a new audit event
func NewAuditEvent(eventType, entityID, entityType, action, userID string) *AuditEvent {
	return &AuditEvent{
		EventID:    generateEventID(),
		EventType:  eventType,
		EntityID:   entityID,
		EntityType: entityType,
		Action:     action,
		UserID:     userID,
		Timestamp:  time.Now(),
		Changes:    make([]Change, 0),
		Metadata:   make(map[string]interface{}),
	}
}

// AddChange adds a change to the audit event
func (e *AuditEvent) AddChange(field string, oldValue, newValue interface{}, operation string) {
	change := Change{
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operation: operation,
	}
	e.Changes = append(e.Changes, change)
}

// AddMetadata adds metadata to the audit event
func (e *AuditEvent) AddMetadata(key string, value interface{}) {
	e.Metadata[key] = value
}

// SetCorrelationID sets the correlation ID
func (e *AuditEvent) SetCorrelationID(correlationID string) {
	e.CorrelationID = correlationID
}

// GetEventID returns the event ID
func (e *AuditEvent) GetEventID() string {
	return e.EventID
}

// GetEventType returns the event type
func (e *AuditEvent) GetEventType() string {
	return e.EventType
}

// GetEntityID returns the entity ID
func (e *AuditEvent) GetEntityID() string {
	return e.EntityID
}

// GetEntityType returns the entity type
func (e *AuditEvent) GetEntityType() string {
	return e.EntityType
}

// GetAction returns the action
func (e *AuditEvent) GetAction() string {
	return e.Action
}

// GetUserID returns the user ID
func (e *AuditEvent) GetUserID() string {
	return e.UserID
}

// GetTimestamp returns the timestamp
func (e *AuditEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetChanges returns the changes
func (e *AuditEvent) GetChanges() []Change {
	return e.Changes
}

// GetMetadata returns the metadata
func (e *AuditEvent) GetMetadata() map[string]interface{} {
	return e.Metadata
}

// GetCorrelationID returns the correlation ID
func (e *AuditEvent) GetCorrelationID() string {
	return e.CorrelationID
}

// HasChanges checks if there are any changes
func (e *AuditEvent) HasChanges() bool {
	return len(e.Changes) > 0
}

// IsCreateEvent checks if this is a create event
func (e *AuditEvent) IsCreateEvent() bool {
	return e.Action == "create"
}

// IsUpdateEvent checks if this is an update event
func (e *AuditEvent) IsUpdateEvent() bool {
	return e.Action == "update"
}

// IsDeleteEvent checks if this is a delete event
func (e *AuditEvent) IsDeleteEvent() bool {
	return e.Action == "delete"
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return "audit_" + time.Now().Format("20060102150405") + "_" + generateRandomString(8)
}

// generateRandomString generates a random string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

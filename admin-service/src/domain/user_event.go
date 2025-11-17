package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventType represents the type of user event
type EventType string

const (
	EventTypeUserCreated EventType = "user_created"
	EventTypeUserUpdated EventType = "user_updated"
	EventTypeUserDeleted EventType = "user_deleted"
)

// UserEvent represents a user-related event in the system
type UserEvent struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID      uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	EventType   EventType              `json:"event_type" gorm:"type:varchar(50);not null;index"`
	Email       string                 `json:"email" gorm:"type:varchar(255);index"`
	Username    string                 `json:"username" gorm:"type:varchar(100)"`
	FirstName   string                 `json:"first_name" gorm:"type:varchar(100)"`
	LastName    string                 `json:"last_name" gorm:"type:varchar(100)"`
	ServiceName string                 `json:"service_name" gorm:"type:varchar(100);not null;index"`
	PerformedBy string                 `json:"performed_by" gorm:"type:varchar(255);not null"`
	EventTime   time.Time              `json:"event_time" gorm:"not null;default:now();index:idx_user_events_event_time,sort:desc"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at" gorm:"not null;default:now()"`
}

// TableName returns the table name for UserEvent
func (UserEvent) TableName() string {
	return "user_events"
}

// UserEventRepository defines the interface for user event persistence
type UserEventRepository interface {
	// Create creates a new user event
	Create(ctx context.Context, event *UserEvent) error

	// GetByID retrieves a user event by ID
	GetByID(ctx context.Context, id uuid.UUID) (*UserEvent, error)

	// GetByUserID retrieves all events for a specific user
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserEvent, int64, error)

	// GetByEventType retrieves events by type
	GetByEventType(ctx context.Context, eventType EventType, limit, offset int) ([]*UserEvent, int64, error)

	// GetByUserIDAndEventType retrieves events for a user filtered by type
	GetByUserIDAndEventType(ctx context.Context, userID uuid.UUID, eventType EventType, limit, offset int) ([]*UserEvent, int64, error)

	// GetByDateRange retrieves events within a date range
	GetByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*UserEvent, int64, error)

	// GetByServiceName retrieves events by service name
	GetByServiceName(ctx context.Context, serviceName string, limit, offset int) ([]*UserEvent, int64, error)

	// Delete deletes a user event
	Delete(ctx context.Context, id uuid.UUID) error

	// Count returns the total count of events
	Count(ctx context.Context) (int64, error)

	// CountByUserID returns the count of events for a specific user
	CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error)
}

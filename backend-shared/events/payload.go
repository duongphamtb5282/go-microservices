package events

import (
	"time"
)

// UserCreatedPayload represents the payload for user created event
type UserCreatedPayload struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	CreatedBy string                 `json:"created_by"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserCreatedPayload creates a new user created payload
func NewUserCreatedPayload(userID, username, email, createdBy string) *UserCreatedPayload {
	return &UserCreatedPayload{
		UserID:    userID,
		Username:  username,
		Email:     email,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the payload
func (p *UserCreatedPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// UserUpdatedPayload represents the payload for user updated event
type UserUpdatedPayload struct {
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	ModifiedBy string                 `json:"modified_by"`
	ModifiedAt time.Time              `json:"modified_at"`
	Changes    map[string]interface{} `json:"changes"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserUpdatedPayload creates a new user updated payload
func NewUserUpdatedPayload(userID, username, email, modifiedBy string, changes map[string]interface{}) *UserUpdatedPayload {
	return &UserUpdatedPayload{
		UserID:     userID,
		Username:   username,
		Email:      email,
		ModifiedBy: modifiedBy,
		ModifiedAt: time.Now(),
		Changes:    changes,
		Metadata:   make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the payload
func (p *UserUpdatedPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// UserDeletedPayload represents the payload for user deleted event
type UserDeletedPayload struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	DeletedBy string                 `json:"deleted_by"`
	DeletedAt time.Time              `json:"deleted_at"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserDeletedPayload creates a new user deleted payload
func NewUserDeletedPayload(userID, username, email, deletedBy, reason string) *UserDeletedPayload {
	return &UserDeletedPayload{
		UserID:    userID,
		Username:  username,
		Email:     email,
		DeletedBy: deletedBy,
		DeletedAt: time.Now(),
		Reason:    reason,
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the payload
func (p *UserDeletedPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// UserLoginPayload represents the payload for user login event
type UserLoginPayload struct {
	UserID    string                 `json:"user_id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	LoginTime time.Time              `json:"login_time"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserLoginPayload creates a new user login payload
func NewUserLoginPayload(userID, username, email, ipAddress, userAgent, sessionID string) *UserLoginPayload {
	return &UserLoginPayload{
		UserID:    userID,
		Username:  username,
		Email:     email,
		LoginTime: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the payload
func (p *UserLoginPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// UserLogoutPayload represents the payload for user logout event
type UserLogoutPayload struct {
	UserID     string                 `json:"user_id"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	LogoutTime time.Time              `json:"logout_time"`
	SessionID  string                 `json:"session_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewUserLogoutPayload creates a new user logout payload
func NewUserLogoutPayload(userID, username, email, sessionID string) *UserLogoutPayload {
	return &UserLogoutPayload{
		UserID:     userID,
		Username:   username,
		Email:      email,
		LogoutTime: time.Now(),
		SessionID:  sessionID,
		Metadata:   make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the payload
func (p *UserLogoutPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// AuditPayload represents the payload for audit events
type AuditPayload struct {
	EntityID   string                 `json:"entity_id"`
	EntityType string                 `json:"entity_type"`
	Action     string                 `json:"action"`
	UserID     string                 `json:"user_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewAuditPayload creates a new audit payload
func NewAuditPayload(entityID, entityType, action, userID string) *AuditPayload {
	return &AuditPayload{
		EntityID:   entityID,
		EntityType: entityType,
		Action:     action,
		UserID:     userID,
		Timestamp:  time.Now(),
		Changes:    make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}
}

// AddChange adds a change to the audit payload
func (p *AuditPayload) AddChange(field string, oldValue, newValue interface{}) {
	if p.Changes == nil {
		p.Changes = make(map[string]interface{})
	}
	p.Changes[field] = map[string]interface{}{
		"old_value": oldValue,
		"new_value": newValue,
	}
}

// AddMetadata adds metadata to the audit payload
func (p *AuditPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// SystemEventPayload represents the payload for system events
type SystemEventPayload struct {
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Message   string                 `json:"message"`
	Level     string                 `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewSystemEventPayload creates a new system event payload
func NewSystemEventPayload(eventType, source, message, level string) *SystemEventPayload {
	return &SystemEventPayload{
		EventType: eventType,
		Source:    source,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the system event payload
func (p *SystemEventPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

// NotificationPayload represents the payload for notification events
type NotificationPayload struct {
	RecipientID   string                 `json:"recipient_id"`
	RecipientType string                 `json:"recipient_type"`
	Channel       string                 `json:"channel"`
	Subject       string                 `json:"subject"`
	Message       string                 `json:"message"`
	Priority      string                 `json:"priority"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// NewNotificationPayload creates a new notification payload
func NewNotificationPayload(recipientID, recipientType, channel, subject, message, priority string) *NotificationPayload {
	return &NotificationPayload{
		RecipientID:   recipientID,
		RecipientType: recipientType,
		Channel:       channel,
		Subject:       subject,
		Message:       message,
		Priority:      priority,
		Timestamp:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}
}

// AddMetadata adds metadata to the notification payload
func (p *NotificationPayload) AddMetadata(key string, value interface{}) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]interface{})
	}
	p.Metadata[key] = value
}

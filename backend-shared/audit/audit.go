package audit

import (
	"time"
)

// AuditEntity represents the base audit entity with common audit fields
type AuditEntity struct {
	CreatedBy  string    `json:"created_by" db:"created_by" validate:"required"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" validate:"required"`
	ModifiedBy string    `json:"modified_by" db:"modified_by" validate:"required"`
	ModifiedAt time.Time `json:"modified_at" db:"modified_at" validate:"required"`
}

// NewAuditEntity creates a new audit entity with the provided user
func NewAuditEntity(createdBy string) AuditEntity {
	now := time.Now()
	return AuditEntity{
		CreatedBy:  createdBy,
		CreatedAt:  now,
		ModifiedBy: createdBy,
		ModifiedAt: now,
	}
}

// UpdateAudit updates the audit fields with the provided user
func (a *AuditEntity) UpdateAudit(modifiedBy string) {
	a.ModifiedBy = modifiedBy
	a.ModifiedAt = time.Now()
}

// GetCreatedBy returns the created by field
func (a *AuditEntity) GetCreatedBy() string {
	return a.CreatedBy
}

// GetCreatedAt returns the created at field
func (a *AuditEntity) GetCreatedAt() time.Time {
	return a.CreatedAt
}

// GetModifiedBy returns the modified by field
func (a *AuditEntity) GetModifiedBy() string {
	return a.ModifiedBy
}

// GetModifiedAt returns the modified at field
func (a *AuditEntity) GetModifiedAt() time.Time {
	return a.ModifiedAt
}

// IsNew checks if the entity is new (created and modified at the same time)
func (a *AuditEntity) IsNew() bool {
	return a.CreatedAt.Equal(a.ModifiedAt)
}

// GetAge returns the age of the entity
func (a *AuditEntity) GetAge() time.Duration {
	return time.Since(a.CreatedAt)
}

// GetLastModifiedAge returns the age since last modification
func (a *AuditEntity) GetLastModifiedAge() time.Duration {
	return time.Since(a.ModifiedAt)
}

// AuditInfo represents audit information for tracking changes
type AuditInfo struct {
	EntityID   string                 `json:"entity_id"`
	EntityType string                 `json:"entity_type"`
	Action     string                 `json:"action"`
	UserID     string                 `json:"user_id"`
	Timestamp  time.Time              `json:"timestamp"`
	Changes    []Change               `json:"changes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Change represents a field change
type Change struct {
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
	Operation string      `json:"operation"` // "create", "update", "delete"
}

// NewAuditInfo creates a new audit info
func NewAuditInfo(entityID, entityType, action, userID string) *AuditInfo {
	return &AuditInfo{
		EntityID:   entityID,
		EntityType: entityType,
		Action:     action,
		UserID:     userID,
		Timestamp:  time.Now(),
		Changes:    make([]Change, 0),
		Metadata:   make(map[string]interface{}),
	}
}

// AddChange adds a change to the audit info
func (a *AuditInfo) AddChange(field string, oldValue, newValue interface{}, operation string) {
	change := Change{
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		Operation: operation,
	}
	a.Changes = append(a.Changes, change)
}

// AddMetadata adds metadata to the audit info
func (a *AuditInfo) AddMetadata(key string, value interface{}) {
	a.Metadata[key] = value
}

// GetChanges returns all changes
func (a *AuditInfo) GetChanges() []Change {
	return a.Changes
}

// GetMetadata returns all metadata
func (a *AuditInfo) GetMetadata() map[string]interface{} {
	return a.Metadata
}

// HasChanges checks if there are any changes
func (a *AuditInfo) HasChanges() bool {
	return len(a.Changes) > 0
}

// AuditTrail represents a complete audit trail for an entity
type AuditTrail struct {
	EntityID   string      `json:"entity_id"`
	EntityType string      `json:"entity_type"`
	History    []AuditInfo `json:"history"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// NewAuditTrail creates a new audit trail
func NewAuditTrail(entityID, entityType string) *AuditTrail {
	now := time.Now()
	return &AuditTrail{
		EntityID:   entityID,
		EntityType: entityType,
		History:    make([]AuditInfo, 0),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// AddAuditInfo adds audit info to the trail
func (t *AuditTrail) AddAuditInfo(info *AuditInfo) {
	t.History = append(t.History, *info)
	t.UpdatedAt = time.Now()
}

// GetLatestAuditInfo returns the latest audit info
func (t *AuditTrail) GetLatestAuditInfo() *AuditInfo {
	if len(t.History) == 0 {
		return nil
	}
	return &t.History[len(t.History)-1]
}

// GetAuditInfoByAction returns audit info by action
func (t *AuditTrail) GetAuditInfoByAction(action string) []AuditInfo {
	var result []AuditInfo
	for _, info := range t.History {
		if info.Action == action {
			result = append(result, info)
		}
	}
	return result
}

// GetAuditInfoByUser returns audit info by user
func (t *AuditTrail) GetAuditInfoByUser(userID string) []AuditInfo {
	var result []AuditInfo
	for _, info := range t.History {
		if info.UserID == userID {
			result = append(result, info)
		}
	}
	return result
}

// GetAuditInfoByDateRange returns audit info within a date range
func (t *AuditTrail) GetAuditInfoByDateRange(start, end time.Time) []AuditInfo {
	var result []AuditInfo
	for _, info := range t.History {
		if info.Timestamp.After(start) && info.Timestamp.Before(end) {
			result = append(result, info)
		}
	}
	return result
}

// GetHistoryCount returns the number of audit entries
func (t *AuditTrail) GetHistoryCount() int {
	return len(t.History)
}

// IsEmpty checks if the audit trail is empty
func (t *AuditTrail) IsEmpty() bool {
	return len(t.History) == 0
}

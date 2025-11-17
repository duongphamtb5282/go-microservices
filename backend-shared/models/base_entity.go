package models

import (
	"time"
)

// BaseEntity represents a base entity with common fields
type BaseEntity struct {
	ID         EntityID  `json:"id"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedBy string    `json:"modified_by"`
	ModifiedAt time.Time `json:"modified_at"`
}

// NewBaseEntity creates a new BaseEntity
func NewBaseEntity(createdBy string) BaseEntity {
	now := time.Now()
	return BaseEntity{
		ID:         NewEntityID(),
		CreatedBy:  createdBy,
		CreatedAt:  now,
		ModifiedBy: createdBy,
		ModifiedAt: now,
	}
}

// NewBaseEntityFromExisting creates a BaseEntity from existing data
func NewBaseEntityFromExisting(id, createdBy string, createdAt, modifiedAt time.Time, modifiedBy string) (BaseEntity, error) {
	entityID, err := NewEntityIDFromString(id)
	if err != nil {
		return BaseEntity{}, err
	}

	return BaseEntity{
		ID:         entityID,
		CreatedBy:  createdBy,
		CreatedAt:  createdAt,
		ModifiedBy: modifiedBy,
		ModifiedAt: modifiedAt,
	}, nil
}

// GetID returns the entity ID
func (e *BaseEntity) GetID() EntityID {
	return e.ID
}

// GetIDString returns the entity ID as string
func (e *BaseEntity) GetIDString() string {
	return e.ID.String()
}

// UpdateAudit updates the audit information
func (e *BaseEntity) UpdateAudit(modifiedBy string) {
	e.ModifiedBy = modifiedBy
	e.ModifiedAt = time.Now()
}

// IsNew checks if the entity is new (created and modified at the same time)
func (e *BaseEntity) IsNew() bool {
	return e.CreatedAt.Equal(e.ModifiedAt)
}

// GetAge returns the age of the entity
func (e *BaseEntity) GetAge() time.Duration {
	return time.Since(e.CreatedAt)
}

// GetLastModifiedAge returns the age since last modification
func (e *BaseEntity) GetLastModifiedAge() time.Duration {
	return time.Since(e.ModifiedAt)
}

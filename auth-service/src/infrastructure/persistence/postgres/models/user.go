package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the database with audit fields
type User struct {
	// Override audit field names to match database schema
	CreatedBy  string    `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt  time.Time `gorm:"column:created_at;not null" json:"created_at"`
	ModifiedBy string    `gorm:"column:updated_by;not null" json:"modified_by"`
	ModifiedAt time.Time `gorm:"column:updated_at;not null" json:"modified_at"`

	ID            string     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Username      string     `gorm:"uniqueIndex;not null;size:20" json:"username"`
	Email         string     `gorm:"uniqueIndex;not null;size:254" json:"email"`
	PasswordHash  string     `gorm:"not null" json:"-"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	LoginAttempts int        `gorm:"default:0" json:"login_attempts"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// NewUser creates a new user with audit information
func NewUser(username, email, passwordHash, createdBy string) *User {
	now := time.Now()
	return &User{
		CreatedBy:     createdBy,
		CreatedAt:     now,
		ModifiedBy:    createdBy,
		ModifiedAt:    now,
		ID:            "", // Will be set by database
		Username:      username,
		Email:         email,
		PasswordHash:  passwordHash,
		IsActive:      true,
		LoginAttempts: 0,
	}
}

// BeforeCreate is called before creating a user - entity listener for audit
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Ensure audit fields are set
	if u.CreatedBy == "" {
		u.CreatedBy = "system"
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	if u.ModifiedBy == "" {
		u.ModifiedBy = u.CreatedBy
	}
	if u.ModifiedAt.IsZero() {
		u.ModifiedAt = u.CreatedAt
	}
	return nil
}

// BeforeUpdate is called before updating a user - entity listener for audit
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Update audit fields
	u.ModifiedBy = "system"
	u.ModifiedAt = time.Now()
	return nil
}

// GetAuditInfo returns basic audit information
func (u *User) GetAuditInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":         u.ID,
		"type":       "user",
		"created_by": u.CreatedBy,
		"updated_by": u.ModifiedBy,
	}
}

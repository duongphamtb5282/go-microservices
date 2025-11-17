package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	Username      string     `json:"username" gorm:"uniqueIndex;not null"`
	Email         string     `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash  string     `json:"-" gorm:"not null"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	LoginAttempts int        `json:"login_attempts" gorm:"default:0"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedBy     string     `json:"created_by"`
	UpdatedBy     string     `json:"updated_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

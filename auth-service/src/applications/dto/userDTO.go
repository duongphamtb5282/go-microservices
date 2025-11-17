package dto

import (
	"time"

	"auth-service/src/domain/entities"
	"backend-shared/audit"
)

// UserDTO represents a user data transfer object with generic audit fields
type UserDTO struct {
	ID                string     `json:"id"`
	Username          string     `json:"username"`
	Email             string     `json:"email"`
	IsActive          bool       `json:"is_active"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty"`
	LoginAttempts     int        `json:"login_attempts"`
	audit.AuditEntity            // Embedded audit entity from backend-shared
}

// NewUserDTO creates a new UserDTO from a User entity using generic audit initialization
func NewUserDTO(user *entities.User) *UserDTO {
	// Create audit entity using backend-shared
	auditInfo := user.AuditInfo()
	auditEntity := audit.AuditEntity{
		CreatedBy:  auditInfo.CreatedBy,
		CreatedAt:  auditInfo.CreatedAt,
		ModifiedBy: auditInfo.CreatedBy, // Use CreatedBy as fallback
		ModifiedAt: auditInfo.CreatedAt, // Use CreatedAt as fallback
	}

	return &UserDTO{
		ID:            user.ID().String(),
		Username:      user.Username().String(),
		Email:         user.Email().String(),
		IsActive:      user.IsActive(),
		LastLoginAt:   user.LastLoginAt(),
		LoginAttempts: user.LoginAttempts(),
		AuditEntity:   auditEntity,
	}
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=100,password"`
}

// CreateUserResponse represents a response for user creation
type CreateUserResponse struct {
	User    *UserDTO `json:"user"`
	Message string   `json:"message"`
}

// GetUserResponse represents a response for getting a user
type GetUserResponse struct {
	User *UserDTO `json:"user"`
}

// ActivateUserRequest represents a request to activate a user
type ActivateUserRequest struct {
	UserID string `json:"user_id" validate:"required"`
}

// ActivateUserResponse represents a response for user activation
type ActivateUserResponse struct {
	Message string `json:"message"`
}

// UserListRequest represents a request to list users
type UserListRequest struct {
	Page  int `form:"page" validate:"min=1"`
	Limit int `form:"limit" validate:"min=1,max=100"`
}

// UserListResponse represents a response for listing users
type UserListResponse struct {
	Users []*UserDTO `json:"users"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

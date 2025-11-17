package dto

import (
	"time"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" binding:"required"`
	Password string `json:"password" validate:"required,min=8" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
	User         UserDTO `json:"user"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,username" binding:"required"`
	Email    string `json:"email" validate:"required,email" binding:"required"`
	Password string `json:"password" validate:"required,password" binding:"required"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	User    UserDTO `json:"user"`
	Message string  `json:"message"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" binding:"required"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required" binding:"required"`
	NewPassword string `json:"new_password" validate:"required,password" binding:"required"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required" binding:"required"`
	NewPassword     string `json:"new_password" validate:"required,password" binding:"required"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required" binding:"required"`
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	Token string `json:"token" validate:"required" binding:"required"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required" binding:"required"`
}

// ResendVerificationRequest represents a resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email" binding:"required"`
}

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserCreateRequest represents a user creation request
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,username" binding:"required"`
	Email    string `json:"email" validate:"required,email" binding:"required"`
	Password string `json:"password" validate:"required,password" binding:"required"`
}

// UserUpdateRequest represents a user update request
type UserUpdateRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,username"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// UserListRequest represents a user list request
type UserListRequest struct {
	Page   int    `form:"page" validate:"min=1" binding:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100" binding:"min=1,max=100"`
	Search string `form:"search" validate:"omitempty,min=1"`
}

// UserListResponse represents a user list response
type UserListResponse struct {
	Users []UserDTO `json:"users"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

// UserPathRequest represents a user path parameter request
type UserPathRequest struct {
	ID string `uri:"id" validate:"required,uuid" binding:"required"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
	Code    string                 `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

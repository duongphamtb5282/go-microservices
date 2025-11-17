package dto

import (
	"time"
)

// RoleDTO represents a role data transfer object
type RoleDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleCreateRequest represents a role creation request
type RoleCreateRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100" binding:"required"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// RoleUpdateRequest represents a role update request
type RoleUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// RoleListRequest represents a role list request
type RoleListRequest struct {
	Page   int    `form:"page" validate:"min=1" binding:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100" binding:"min=1,max=100"`
	Search string `form:"search" validate:"omitempty,min=1"`
}

// RoleListResponse represents a role list response
type RoleListResponse struct {
	Roles []RoleDTO `json:"roles"`
	Total int64     `json:"total"`
	Page  int       `json:"page"`
	Limit int       `json:"limit"`
}

// RolePathRequest represents a role path parameter request
type RolePathRequest struct {
	ID string `uri:"id" validate:"required,uuid" binding:"required"`
}

// PermissionDTO represents a permission data transfer object
type PermissionDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PermissionCreateRequest represents a permission creation request
type PermissionCreateRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100" binding:"required"`
	Resource    string `json:"resource" validate:"required,min=3,max=100" binding:"required"`
	Action      string `json:"action" validate:"required,min=3,max=50" binding:"required"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// PermissionUpdateRequest represents a permission update request
type PermissionUpdateRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Resource    *string `json:"resource,omitempty" validate:"omitempty,min=3,max=100"`
	Action      *string `json:"action,omitempty" validate:"omitempty,min=3,max=50"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// PermissionListRequest represents a permission list request
type PermissionListRequest struct {
	Page   int    `form:"page" validate:"min=1" binding:"min=1"`
	Limit  int    `form:"limit" validate:"min=1,max=100" binding:"min=1,max=100"`
	Search string `form:"search" validate:"omitempty,min=1"`
}

// PermissionListResponse represents a permission list response
type PermissionListResponse struct {
	Permissions []PermissionDTO `json:"permissions"`
	Total       int64           `json:"total"`
	Page        int             `json:"page"`
	Limit       int             `json:"limit"`
}

// PermissionPathRequest represents a permission path parameter request
type PermissionPathRequest struct {
	ID string `uri:"id" validate:"required,uuid" binding:"required"`
}

// AssignRoleRequest represents a role assignment request
type AssignRoleRequest struct {
	UserID string `json:"user_id" validate:"required,uuid" binding:"required"`
	RoleID string `json:"role_id" validate:"required,uuid" binding:"required"`
}

// RemoveRoleRequest represents a role removal request
type RemoveRoleRequest struct {
	UserID string `json:"user_id" validate:"required,uuid" binding:"required"`
	RoleID string `json:"role_id" validate:"required,uuid" binding:"required"`
}

// AssignPermissionRequest represents a permission assignment request
type AssignPermissionRequest struct {
	RoleID       string `json:"role_id" validate:"required,uuid" binding:"required"`
	PermissionID string `json:"permission_id" validate:"required,uuid" binding:"required"`
}

// RemovePermissionRequest represents a permission removal request
type RemovePermissionRequest struct {
	RoleID       string `json:"role_id" validate:"required,uuid" binding:"required"`
	PermissionID string `json:"permission_id" validate:"required,uuid" binding:"required"`
}

// UserRolesRequest represents a user roles request
type UserRolesRequest struct {
	UserID string `uri:"user_id" validate:"required,uuid" binding:"required"`
}

// UserRolesResponse represents a user roles response
type UserRolesResponse struct {
	UserID string    `json:"user_id"`
	Roles  []RoleDTO `json:"roles"`
}

// RolePermissionsRequest represents a role permissions request
type RolePermissionsRequest struct {
	RoleID string `uri:"role_id" validate:"required,uuid" binding:"required"`
}

// RolePermissionsResponse represents a role permissions response
type RolePermissionsResponse struct {
	RoleID      string          `json:"role_id"`
	Permissions []PermissionDTO `json:"permissions"`
}

// UserPermissionsRequest represents a user permissions request
type UserPermissionsRequest struct {
	UserID string `uri:"user_id" validate:"required,uuid" binding:"required"`
}

// UserPermissionsResponse represents a user permissions response
type UserPermissionsResponse struct {
	UserID      string          `json:"user_id"`
	Permissions []PermissionDTO `json:"permissions"`
}

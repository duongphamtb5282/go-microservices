package authorization

import (
	"context"
	"time"

	"backend-shared/models"

	"github.com/google/uuid"
)

// Role represents a role in the system
type Role struct {
	models.BaseEntity `json:",inline"`
	Name              string `json:"name" db:"name"`
	Description       string `json:"description" db:"description"`
	IsActive          bool   `json:"is_active" db:"is_active"`
}

// GetUUID returns the ID as uuid.UUID for backward compatibility
func (r *Role) GetUUID() uuid.UUID {
	id, _ := uuid.Parse(r.ID.String())
	return id
}

// SetUUID sets the ID from uuid.UUID
func (r *Role) SetUUID(id uuid.UUID) {
	entityID, _ := models.NewEntityIDFromString(id.String())
	r.ID = entityID
}

// Permission represents a permission in the system
type Permission struct {
	models.BaseEntity `json:",inline"`
	Name              string `json:"name" db:"name"`
	Resource          string `json:"resource" db:"resource"`
	Action            string `json:"action" db:"action"`
	Description       string `json:"description" db:"description"`
	IsActive          bool   `json:"is_active" db:"is_active"`
}

// GetUUID returns the ID as uuid.UUID for backward compatibility
func (p *Permission) GetUUID() uuid.UUID {
	id, _ := uuid.Parse(p.ID.String())
	return id
}

// SetUUID sets the ID from uuid.UUID
func (p *Permission) SetUUID(id uuid.UUID) {
	entityID, _ := models.NewEntityIDFromString(id.String())
	p.ID = entityID
}

// UserRole represents the relationship between users and roles
type UserRole struct {
	models.BaseEntity `json:",inline"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	RoleID            uuid.UUID `json:"role_id" db:"role_id"`
	AssignedAt        time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy        string    `json:"assigned_by" db:"assigned_by"`
}

// GetUUID returns the ID as uuid.UUID for backward compatibility
func (ur *UserRole) GetUUID() uuid.UUID {
	id, _ := uuid.Parse(ur.ID.String())
	return id
}

// SetUUID sets the ID from uuid.UUID
func (ur *UserRole) SetUUID(id uuid.UUID) {
	entityID, _ := models.NewEntityIDFromString(id.String())
	ur.ID = entityID
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	models.BaseEntity `json:",inline"`
	RoleID            uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID      uuid.UUID `json:"permission_id" db:"permission_id"`
	AssignedAt        time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy        string    `json:"assigned_by" db:"assigned_by"`
}

// GetUUID returns the ID as uuid.UUID for backward compatibility
func (rp *RolePermission) GetUUID() uuid.UUID {
	id, _ := uuid.Parse(rp.ID.String())
	return id
}

// SetUUID sets the ID from uuid.UUID
func (rp *RolePermission) SetUUID(id uuid.UUID) {
	entityID, _ := models.NewEntityIDFromString(id.String())
	rp.ID = entityID
}

// RoleRepository defines the interface for role persistence
type RoleRepository interface {
	// Create creates a new role
	Create(ctx context.Context, role *Role) error

	// GetByID retrieves a role by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)

	// GetByName retrieves a role by name
	GetByName(ctx context.Context, name string) (*Role, error)

	// GetAll retrieves all roles
	GetAll(ctx context.Context) ([]*Role, error)

	// GetActiveRoles retrieves all active roles
	GetActiveRoles(ctx context.Context) ([]*Role, error)

	// Update updates a role
	Update(ctx context.Context, role *Role) error

	// Delete deletes a role
	Delete(ctx context.Context, id uuid.UUID) error

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)

	// AssignRoleToUser assigns a role to a user
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, assignedBy string) error

	// RemoveRoleFromUser removes a role from a user
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
}

// PermissionRepository defines the interface for permission persistence
type PermissionRepository interface {
	// Create creates a new permission
	Create(ctx context.Context, permission *Permission) error

	// GetByID retrieves a permission by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)

	// GetByName retrieves a permission by name
	GetByName(ctx context.Context, name string) (*Permission, error)

	// GetByResourceAndAction retrieves a permission by resource and action
	GetByResourceAndAction(ctx context.Context, resource, action string) (*Permission, error)

	// GetAll retrieves all permissions
	GetAll(ctx context.Context) ([]*Permission, error)

	// GetActivePermissions retrieves all active permissions
	GetActivePermissions(ctx context.Context) ([]*Permission, error)

	// Update updates a permission
	Update(ctx context.Context, permission *Permission) error

	// Delete deletes a permission
	Delete(ctx context.Context, id uuid.UUID) error

	// GetRolePermissions retrieves all permissions for a role
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)

	// GetUserPermissions retrieves all permissions for a user (through roles)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*Permission, error)

	// AssignPermissionToRole assigns a permission to a role
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID, assignedBy string) error

	// RemovePermissionFromRole removes a permission from a role
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error

	// CheckUserPermission checks if a user has a specific permission
	CheckUserPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error)
}

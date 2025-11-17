package role

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type RoleApi struct {
	roleHandler *handlers.RoleHandler
	logger      *logging.Logger
}

func NewRoleApi(roleHandler *handlers.RoleHandler, logger *logging.Logger) *RoleApi {
	return &RoleApi{
		roleHandler: roleHandler,
		logger:      logger,
	}
}

// CreateRole handles role creation
func (r *RoleApi) CreateRole(c *gin.Context) {
	r.roleHandler.CreateRole(c)
}

// UpdateRole handles role update
func (r *RoleApi) UpdateRole(c *gin.Context) {
	r.roleHandler.UpdateRole(c)
}

// DeleteRole handles role deletion
func (r *RoleApi) DeleteRole(c *gin.Context) {
	r.roleHandler.DeleteRole(c)
}

// AssignRole handles role assignment
func (r *RoleApi) AssignRole(c *gin.Context) {
	r.roleHandler.AssignRole(c)
}

// RemoveRole handles role removal
func (r *RoleApi) RemoveRole(c *gin.Context) {
	r.roleHandler.RemoveRole(c)
}

// UpdateRolePermissions handles role permissions update
func (r *RoleApi) UpdateRolePermissions(c *gin.Context) {
	r.roleHandler.UpdateRolePermissions(c)
}

// GetRoleList handles getting role list
func (r *RoleApi) GetRoleList(c *gin.Context) {
	r.roleHandler.GetRoleList(c)
}

// GetRole handles getting role by ID
func (r *RoleApi) GetRole(c *gin.Context) {
	r.roleHandler.GetRole(c)
}

// SearchRoles handles role search
func (r *RoleApi) SearchRoles(c *gin.Context) {
	r.roleHandler.SearchRoles(c)
}

// GetUserRoles handles getting user roles
func (r *RoleApi) GetUserRoles(c *gin.Context) {
	r.roleHandler.GetUserRoles(c)
}

// GetRoleStats handles getting role statistics
func (r *RoleApi) GetRoleStats(c *gin.Context) {
	r.roleHandler.GetRoleStats(c)
}

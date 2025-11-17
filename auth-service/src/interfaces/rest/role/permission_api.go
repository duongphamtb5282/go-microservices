package role

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type PermissionApi struct {
	permissionHandler *handlers.PermissionHandler
	logger            *logging.Logger
}

func NewPermissionApi(permissionHandler *handlers.PermissionHandler, logger *logging.Logger) *PermissionApi {
	return &PermissionApi{
		permissionHandler: permissionHandler,
		logger:            logger,
	}
}

// CreatePermission handles permission creation
func (p *PermissionApi) CreatePermission(c *gin.Context) {
	p.permissionHandler.CreatePermission(c)
}

// UpdatePermission handles permission update
func (p *PermissionApi) UpdatePermission(c *gin.Context) {
	p.permissionHandler.UpdatePermission(c)
}

// DeletePermission handles permission deletion
func (p *PermissionApi) DeletePermission(c *gin.Context) {
	p.permissionHandler.DeletePermission(c)
}

// AssignPermission handles permission assignment
func (p *PermissionApi) AssignPermission(c *gin.Context) {
	p.permissionHandler.AssignPermission(c)
}

// RemovePermission handles permission removal
func (p *PermissionApi) RemovePermission(c *gin.Context) {
	p.permissionHandler.RemovePermission(c)
}

// GetPermissionList handles getting permission list
func (p *PermissionApi) GetPermissionList(c *gin.Context) {
	p.permissionHandler.GetPermissionList(c)
}

// GetPermission handles getting permission by ID
func (p *PermissionApi) GetPermission(c *gin.Context) {
	p.permissionHandler.GetPermission(c)
}

// SearchPermissions handles permission search
func (p *PermissionApi) SearchPermissions(c *gin.Context) {
	p.permissionHandler.SearchPermissions(c)
}

// GetRolePermissions handles getting role permissions
func (p *PermissionApi) GetRolePermissions(c *gin.Context) {
	p.permissionHandler.GetRolePermissions(c)
}

// GetUserPermissions handles getting user permissions
func (p *PermissionApi) GetUserPermissions(c *gin.Context) {
	p.permissionHandler.GetUserPermissions(c)
}

// GetPermissionStats handles getting permission statistics
func (p *PermissionApi) GetPermissionStats(c *gin.Context) {
	p.permissionHandler.GetPermissionStats(c)
}

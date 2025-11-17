package role

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type PermissionRouter struct{}

// InitPermissionRouter initializes permission management routes (auth required)
func (s *PermissionRouter) InitPermissionRouter(PrivateGroup *gin.RouterGroup) {
	permissionRouter := PrivateGroup.Group("permission").Use(middleware.OperationRecord())
	permissionRouterWithoutRecord := PrivateGroup.Group("permission")
	{
		permissionRouter.POST("create", permissionApi.CreatePermission)       // Create permission
		permissionRouter.PUT("update/:id", permissionApi.UpdatePermission)    // Update permission
		permissionRouter.DELETE("delete/:id", permissionApi.DeletePermission) // Delete permission
		permissionRouter.POST("assign/:id", permissionApi.AssignPermission)   // Assign permission to role
		permissionRouter.POST("remove/:id", permissionApi.RemovePermission)   // Remove permission from role
	}
	{
		permissionRouterWithoutRecord.GET("list", permissionApi.GetPermissionList)          // Get permission list
		permissionRouterWithoutRecord.GET(":id", permissionApi.GetPermission)               // Get permission by ID
		permissionRouterWithoutRecord.GET("search", permissionApi.SearchPermissions)        // Search permissions
		permissionRouterWithoutRecord.GET("role/:roleId", permissionApi.GetRolePermissions) // Get role permissions
		permissionRouterWithoutRecord.GET("user/:userId", permissionApi.GetUserPermissions) // Get user permissions
		permissionRouterWithoutRecord.GET("stats", permissionApi.GetPermissionStats)        // Get permission statistics
	}
}

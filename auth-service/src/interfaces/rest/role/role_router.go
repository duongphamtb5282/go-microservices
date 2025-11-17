package role

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type RoleRouter struct{}

// InitRoleRouter initializes role management routes (auth required)
func (s *RoleRouter) InitRoleRouter(PrivateGroup *gin.RouterGroup) {
	roleRouter := PrivateGroup.Group("role").Use(middleware.OperationRecord())
	roleRouterWithoutRecord := PrivateGroup.Group("role")
	{
		roleRouter.POST("create", roleApi.CreateRole)                     // Create role
		roleRouter.PUT("update/:id", roleApi.UpdateRole)                  // Update role
		roleRouter.DELETE("delete/:id", roleApi.DeleteRole)               // Delete role
		roleRouter.POST("assign/:id", roleApi.AssignRole)                 // Assign role to user
		roleRouter.POST("remove/:id", roleApi.RemoveRole)                 // Remove role from user
		roleRouter.POST("permissions/:id", roleApi.UpdateRolePermissions) // Update role permissions
	}
	{
		roleRouterWithoutRecord.GET("list", roleApi.GetRoleList)          // Get role list
		roleRouterWithoutRecord.GET(":id", roleApi.GetRole)               // Get role by ID
		roleRouterWithoutRecord.GET("search", roleApi.SearchRoles)        // Search roles
		roleRouterWithoutRecord.GET("user/:userId", roleApi.GetUserRoles) // Get user roles
		roleRouterWithoutRecord.GET("stats", roleApi.GetRoleStats)        // Get role statistics
	}
}

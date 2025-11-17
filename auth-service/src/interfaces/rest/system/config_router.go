package system

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type ConfigRouter struct{}

// InitConfigRouter initializes configuration routes (auth required)
func (s *ConfigRouter) InitConfigRouter(PrivateGroup *gin.RouterGroup) {
	configRouter := PrivateGroup.Group("config").Use(middleware.OperationRecord())
	configRouterWithoutRecord := PrivateGroup.Group("config")
	{
		configRouter.POST("update", configApi.UpdateConfig)   // Update configuration
		configRouter.POST("reset", configApi.ResetConfig)     // Reset configuration
		configRouter.POST("backup", configApi.BackupConfig)   // Backup configuration
		configRouter.POST("restore", configApi.RestoreConfig) // Restore configuration
	}
	{
		configRouterWithoutRecord.GET("", configApi.GetConfig)             // Get configuration
		configRouterWithoutRecord.GET("public", configApi.GetPublicConfig) // Get public configuration
		configRouterWithoutRecord.GET("version", configApi.GetVersion)     // Get service version
		configRouterWithoutRecord.GET("info", configApi.GetServiceInfo)    // Get service information
	}
}

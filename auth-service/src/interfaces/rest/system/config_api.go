package system

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type ConfigApi struct {
	configHandler *handlers.ConfigHandler
	logger        *logging.Logger
}

func NewConfigApi(configHandler *handlers.ConfigHandler, logger *logging.Logger) *ConfigApi {
	return &ConfigApi{
		configHandler: configHandler,
		logger:        logger,
	}
}

// UpdateConfig handles configuration update
func (c *ConfigApi) UpdateConfig(ctx *gin.Context) {
	c.configHandler.UpdateConfig(ctx)
}

// ResetConfig handles configuration reset
func (c *ConfigApi) ResetConfig(ctx *gin.Context) {
	c.configHandler.ResetConfig(ctx)
}

// BackupConfig handles configuration backup
func (c *ConfigApi) BackupConfig(ctx *gin.Context) {
	c.configHandler.BackupConfig(ctx)
}

// RestoreConfig handles configuration restore
func (c *ConfigApi) RestoreConfig(ctx *gin.Context) {
	c.configHandler.RestoreConfig(ctx)
}

// GetConfig handles getting configuration
func (c *ConfigApi) GetConfig(ctx *gin.Context) {
	c.configHandler.GetConfig(ctx)
}

// GetPublicConfig handles getting public configuration
func (c *ConfigApi) GetPublicConfig(ctx *gin.Context) {
	c.configHandler.GetPublicConfig(ctx)
}

// GetVersion handles getting service version
func (c *ConfigApi) GetVersion(ctx *gin.Context) {
	c.configHandler.GetVersion(ctx)
}

// GetServiceInfo handles getting service information
func (c *ConfigApi) GetServiceInfo(ctx *gin.Context) {
	c.configHandler.GetServiceInfo(ctx)
}

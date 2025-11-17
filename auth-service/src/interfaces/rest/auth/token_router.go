package auth

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type TokenRouter struct{}

// InitTokenRouter initializes token management routes (auth required)
func (s *TokenRouter) InitTokenRouter(PrivateGroup *gin.RouterGroup) {
	tokenRouter := PrivateGroup.Group("token").Use(middleware.OperationRecord())
	tokenRouterWithoutRecord := PrivateGroup.Group("token")
	{
		tokenRouter.POST("generate", tokenApi.GenerateToken)     // Generate new token
		tokenRouter.POST("revoke", tokenApi.RevokeToken)         // Revoke specific token
		tokenRouter.POST("revoke-all", tokenApi.RevokeAllTokens) // Revoke all user tokens
		tokenRouter.POST("validate", tokenApi.ValidateToken)     // Validate token
	}
	{
		tokenRouterWithoutRecord.GET("list", tokenApi.GetTokenList)   // Get user token list
		tokenRouterWithoutRecord.GET("info", tokenApi.GetTokenInfo)   // Get token information
		tokenRouterWithoutRecord.GET("stats", tokenApi.GetTokenStats) // Get token statistics
	}
}

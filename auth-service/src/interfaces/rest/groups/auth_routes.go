package groups

import (
	"auth-service/src/interfaces/rest/handlers"

	"github.com/gin-gonic/gin"
)

// AuthRoutes defines authentication-related routes
type AuthRoutes struct {
	authHandler *handlers.AuthHandler
}

// NewAuthRoutes creates a new auth routes group
func NewAuthRoutes(authHandler *handlers.AuthHandler) *AuthRoutes {
	return &AuthRoutes{
		authHandler: authHandler,
	}
}

// RegisterRoutes registers all authentication routes
func (r *AuthRoutes) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", r.authHandler.Login)
		auth.POST("/register", r.authHandler.Register)
		auth.POST("/logout", r.authHandler.Logout)
		auth.POST("/forgot-password", r.authHandler.ForgotPassword)
		auth.POST("/reset-password", r.authHandler.ResetPassword)
		auth.POST("/verify-email", r.authHandler.VerifyEmail)
		auth.POST("/change-password", r.authHandler.ChangePassword)
		auth.POST("/refresh-token", r.authHandler.RefreshToken)
	}
}

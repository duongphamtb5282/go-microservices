package auth

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type AuthRouter struct{}

// InitAuthPublicRouter initializes public authentication routes (no auth required)
func (s *AuthRouter) InitAuthPublicRouter(PublicGroup *gin.RouterGroup) {
	authPublic := PublicGroup.Group("auth")
	{
		authPublic.POST("login", authApi.Login)                    // User login
		authPublic.POST("register", authApi.Register)              // User registration
		authPublic.POST("forgot-password", authApi.ForgotPassword) // Forgot password
		authPublic.POST("reset-password", authApi.ResetPassword)   // Reset password
		authPublic.POST("verify-email", authApi.VerifyEmail)       // Email verification
	}
}

// InitAuthRouter initializes private authentication routes (auth required)
func (s *AuthRouter) InitAuthRouter(PrivateGroup *gin.RouterGroup) {
	authRouter := PrivateGroup.Group("auth").Use(middleware.OperationRecord())
	authRouterWithoutRecord := PrivateGroup.Group("auth")
	{
		authRouter.POST("logout", authApi.Logout)                  // User logout
		authRouter.POST("change-password", authApi.ChangePassword) // Change password
		authRouter.POST("refresh-token", authApi.RefreshToken)     // Refresh token
		authRouter.POST("revoke-token", authApi.RevokeToken)       // Revoke token
		authRouter.POST("logout-all", authApi.LogoutAll)           // Logout from all devices
	}
	{
		authRouterWithoutRecord.GET("profile", authApi.GetProfile)    // Get user profile
		authRouterWithoutRecord.PUT("profile", authApi.UpdateProfile) // Update user profile
		authRouterWithoutRecord.GET("sessions", authApi.GetSessions)  // Get active sessions
	}
}

package user

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type ProfileRouter struct{}

// InitProfileRouter initializes user profile routes (auth required)
func (s *ProfileRouter) InitProfileRouter(PrivateGroup *gin.RouterGroup) {
	profileRouter := PrivateGroup.Group("profile").Use(middleware.OperationRecord())
	profileRouterWithoutRecord := PrivateGroup.Group("profile")
	{
		profileRouter.PUT("update", profileApi.UpdateProfile)                 // Update user profile
		profileRouter.PUT("change-password", profileApi.ChangePassword)       // Change own password
		profileRouter.PUT("update-avatar", profileApi.UpdateAvatar)           // Update profile avatar
		profileRouter.PUT("update-preferences", profileApi.UpdatePreferences) // Update user preferences
		profileRouter.DELETE("delete-account", profileApi.DeleteAccount)      // Delete own account
	}
	{
		profileRouterWithoutRecord.GET("", profileApi.GetProfile)                // Get user profile
		profileRouterWithoutRecord.GET("preferences", profileApi.GetPreferences) // Get user preferences
		profileRouterWithoutRecord.GET("activity", profileApi.GetActivity)       // Get user activity
		profileRouterWithoutRecord.GET("sessions", profileApi.GetSessions)       // Get user sessions
	}
}

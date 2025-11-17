package user

import (
	"auth-service/src/interfaces/rest/middleware"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

// InitUserPublicRouter initializes public user routes (no auth required)
func (s *UserRouter) InitUserPublicRouter(PublicGroup *gin.RouterGroup) {
	userPublic := PublicGroup.Group("user")
	{
		userPublic.POST("register", userApi.Register)                      // User registration
		userPublic.POST("verify-email", userApi.VerifyEmail)               // Email verification
		userPublic.POST("resend-verification", userApi.ResendVerification) // Resend verification email
	}
}

// InitUserRouter initializes private user routes (auth required)
func (s *UserRouter) InitUserRouter(PrivateGroup *gin.RouterGroup) {
	userRouter := PrivateGroup.Group("user").Use(middleware.OperationRecord())
	userRouterWithoutRecord := PrivateGroup.Group("user")
	{
		userRouter.POST("create", userApi.CreateUser)                  // Create user (admin)
		userRouter.PUT("update/:id", userApi.UpdateUser)               // Update user
		userRouter.DELETE("delete/:id", userApi.DeleteUser)            // Delete user
		userRouter.POST("activate/:id", userApi.ActivateUser)          // Activate user
		userRouter.POST("deactivate/:id", userApi.DeactivateUser)      // Deactivate user
		userRouter.POST("change-password/:id", userApi.ChangePassword) // Change user password
		userRouter.POST("reset-password/:id", userApi.ResetPassword)   // Reset user password
	}
	{
		userRouterWithoutRecord.GET("list", userApi.GetUserList)   // Get user list
		userRouterWithoutRecord.GET(":id", userApi.GetUser)        // Get user by ID
		userRouterWithoutRecord.GET("search", userApi.SearchUsers) // Search users
		userRouterWithoutRecord.GET("stats", userApi.GetUserStats) // Get user statistics
	}
}

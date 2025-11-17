package user

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type UserApi struct {
	userHandler *handlers.UserHandler
	logger      *logging.Logger
}

func NewUserApi(userHandler *handlers.UserHandler, logger *logging.Logger) *UserApi {
	return &UserApi{
		userHandler: userHandler,
		logger:      logger,
	}
}

// Register handles user registration
func (u *UserApi) Register(c *gin.Context) {
	u.userHandler.CreateUser(c)
}

// VerifyEmail handles email verification
func (u *UserApi) VerifyEmail(c *gin.Context) {
	u.userHandler.VerifyEmail(c)
}

// ResendVerification handles resending verification email
func (u *UserApi) ResendVerification(c *gin.Context) {
	u.userHandler.ResendVerification(c)
}

// CreateUser handles user creation (admin)
func (u *UserApi) CreateUser(c *gin.Context) {
	u.userHandler.CreateUser(c)
}

// UpdateUser handles user update
func (u *UserApi) UpdateUser(c *gin.Context) {
	u.userHandler.UpdateUser(c)
}

// DeleteUser handles user deletion
func (u *UserApi) DeleteUser(c *gin.Context) {
	u.userHandler.DeleteUser(c)
}

// ActivateUser handles user activation
func (u *UserApi) ActivateUser(c *gin.Context) {
	u.userHandler.ActivateUser(c)
}

// DeactivateUser handles user deactivation
func (u *UserApi) DeactivateUser(c *gin.Context) {
	u.userHandler.DeactivateUser(c)
}

// ChangePassword handles password change
func (u *UserApi) ChangePassword(c *gin.Context) {
	u.userHandler.ChangePassword(c)
}

// ResetPassword handles password reset
func (u *UserApi) ResetPassword(c *gin.Context) {
	u.userHandler.ResetPassword(c)
}

// GetUserList handles getting user list
func (u *UserApi) GetUserList(c *gin.Context) {
	u.userHandler.ListUsers(c)
}

// GetUser handles getting user by ID
func (u *UserApi) GetUser(c *gin.Context) {
	u.userHandler.GetUser(c)
}

// SearchUsers handles user search
func (u *UserApi) SearchUsers(c *gin.Context) {
	u.userHandler.SearchUsers(c)
}

// GetUserStats handles getting user statistics
func (u *UserApi) GetUserStats(c *gin.Context) {
	u.userHandler.GetUserStats(c)
}

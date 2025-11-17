package auth

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type AuthApi struct {
	authHandler *handlers.AuthHandler
	logger      *logging.Logger
}

func NewAuthApi(authHandler *handlers.AuthHandler, logger *logging.Logger) *AuthApi {
	return &AuthApi{
		authHandler: authHandler,
		logger:      logger,
	}
}

// Login handles user login
func (a *AuthApi) Login(c *gin.Context) {
	a.authHandler.Login(c)
}

// Register handles user registration
func (a *AuthApi) Register(c *gin.Context) {
	a.authHandler.Register(c)
}

// Logout handles user logout
func (a *AuthApi) Logout(c *gin.Context) {
	a.authHandler.Logout(c)
}

// ForgotPassword handles forgot password request
func (a *AuthApi) ForgotPassword(c *gin.Context) {
	a.authHandler.ForgotPassword(c)
}

// ResetPassword handles password reset
func (a *AuthApi) ResetPassword(c *gin.Context) {
	a.authHandler.ResetPassword(c)
}

// VerifyEmail handles email verification
func (a *AuthApi) VerifyEmail(c *gin.Context) {
	a.authHandler.VerifyEmail(c)
}

// ChangePassword handles password change
func (a *AuthApi) ChangePassword(c *gin.Context) {
	a.authHandler.ChangePassword(c)
}

// RefreshToken handles token refresh
func (a *AuthApi) RefreshToken(c *gin.Context) {
	a.authHandler.RefreshToken(c)
}

// RevokeToken handles token revocation
func (a *AuthApi) RevokeToken(c *gin.Context) {
	a.authHandler.RevokeToken(c)
}

// LogoutAll handles logout from all devices
func (a *AuthApi) LogoutAll(c *gin.Context) {
	a.authHandler.LogoutAll(c)
}

// GetProfile handles getting user profile
func (a *AuthApi) GetProfile(c *gin.Context) {
	a.authHandler.GetProfile(c)
}

// UpdateProfile handles updating user profile
func (a *AuthApi) UpdateProfile(c *gin.Context) {
	a.authHandler.UpdateProfile(c)
}

// GetSessions handles getting user sessions
func (a *AuthApi) GetSessions(c *gin.Context) {
	a.authHandler.GetSessions(c)
}

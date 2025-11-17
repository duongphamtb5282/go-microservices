package user

import (
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

type ProfileApi struct {
	profileHandler *handlers.ProfileHandler
	logger         *logging.Logger
}

func NewProfileApi(profileHandler *handlers.ProfileHandler, logger *logging.Logger) *ProfileApi {
	return &ProfileApi{
		profileHandler: profileHandler,
		logger:         logger,
	}
}

// UpdateProfile handles profile update
func (p *ProfileApi) UpdateProfile(c *gin.Context) {
	p.profileHandler.UpdateProfile(c)
}

// ChangePassword handles password change
func (p *ProfileApi) ChangePassword(c *gin.Context) {
	p.profileHandler.ChangePassword(c)
}

// UpdateAvatar handles avatar update
func (p *ProfileApi) UpdateAvatar(c *gin.Context) {
	p.profileHandler.UpdateAvatar(c)
}

// UpdatePreferences handles preferences update
func (p *ProfileApi) UpdatePreferences(c *gin.Context) {
	p.profileHandler.UpdatePreferences(c)
}

// DeleteAccount handles account deletion
func (p *ProfileApi) DeleteAccount(c *gin.Context) {
	p.profileHandler.DeleteAccount(c)
}

// GetProfile handles getting user profile
func (p *ProfileApi) GetProfile(c *gin.Context) {
	p.profileHandler.GetProfile(c)
}

// GetPreferences handles getting user preferences
func (p *ProfileApi) GetPreferences(c *gin.Context) {
	p.profileHandler.GetPreferences(c)
}

// GetActivity handles getting user activity
func (p *ProfileApi) GetActivity(c *gin.Context) {
	p.profileHandler.GetActivity(c)
}

// GetSessions handles getting user sessions
func (p *ProfileApi) GetSessions(c *gin.Context) {
	p.profileHandler.GetSessions(c)
}

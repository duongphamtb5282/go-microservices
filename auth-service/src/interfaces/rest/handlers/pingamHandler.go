package handlers

import (
	"net/http"
	"time"

	"auth-service/src/applications/services"
	"auth-service/src/infrastructure/identity/models"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// PingAMHandler handles PingAM-related HTTP requests
type PingAMHandler struct {
	pingamService *services.PingAMApplicationService
	logger        *logging.Logger
}

// NewPingAMHandler creates a new PingAM handler
func NewPingAMHandler(pingamService *services.PingAMApplicationService, logger *logging.Logger) *PingAMHandler {
	return &PingAMHandler{
		pingamService: pingamService,
		logger:        logger,
	}
}

// LoginWithPingAM handles PingAM authentication
func (h *PingAMHandler) LoginWithPingAM(c *gin.Context) {
	var credentials models.Credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		h.logger.Error("Invalid request body for PingAM login",
			logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("PingAM login attempt",
		logging.String("username", credentials.Username))

	authResult, err := h.pingamService.LoginWithPingAM(c.Request.Context(), credentials)
	if err != nil {
		h.logger.Error("PingAM login failed",
			logging.Error(err),
			logging.String("username", credentials.Username))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("PingAM login successful",
		logging.String("user_id", authResult.UserID),
		logging.String("username", authResult.Username))

	c.JSON(http.StatusOK, gin.H{
		"access_token":  authResult.AccessToken,
		"refresh_token": authResult.RefreshToken,
		"token_type":    authResult.TokenType,
		"expires_in":    authResult.ExpiresIn,
		"user_id":       authResult.UserID,
		"username":      authResult.Username,
		"scope":         authResult.Scope,
		"issued_at":     authResult.IssuedAt,
	})
}

// ValidateSession validates session with PingAM
func (h *PingAMHandler) ValidateSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session ID is required",
		})
		return
	}

	h.logger.Debug("Validating session with PingAM",
		logging.String("session_id", sessionID))

	session, err := h.pingamService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("Session validation failed",
			logging.Error(err),
			logging.String("session_id", sessionID))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid session",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("Session validated successfully",
		logging.String("session_id", sessionID),
		logging.String("user_id", session.UserID))

	c.JSON(http.StatusOK, gin.H{
		"session_id":  session.ID,
		"user_id":     session.UserID,
		"is_active":   session.IsActive,
		"expires_at":  session.ExpiresAt,
		"last_access": session.LastAccess,
		"created_at":  session.CreatedAt,
	})
}

// CheckPermission checks user permission
func (h *PingAMHandler) CheckPermission(c *gin.Context) {
	var request struct {
		UserID   string `json:"user_id" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body for permission check",
			logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("Checking permission with PingAM",
		logging.String("user_id", request.UserID),
		logging.String("resource", request.Resource),
		logging.String("action", request.Action))

	allowed, err := h.pingamService.CheckPermission(c.Request.Context(), request.UserID, request.Resource, request.Action)
	if err != nil {
		h.logger.Error("Permission check failed",
			logging.Error(err),
			logging.String("user_id", request.UserID),
			logging.String("resource", request.Resource),
			logging.String("action", request.Action))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Permission check failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("Permission check completed",
		logging.String("user_id", request.UserID),
		logging.String("resource", request.Resource),
		logging.String("action", request.Action),
		logging.Bool("allowed", allowed))

	c.JSON(http.StatusOK, gin.H{
		"allowed":  allowed,
		"user_id":  request.UserID,
		"resource": request.Resource,
		"action":   request.Action,
	})
}

// GetUserRoles retrieves user roles
func (h *PingAMHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	h.logger.Debug("Getting user roles from PingAM",
		logging.String("user_id", userID))

	roles, err := h.pingamService.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user roles",
			logging.Error(err),
			logging.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user roles",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("User roles retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"roles":   roles,
		"count":   len(roles),
	})
}

// GetUserPermissions retrieves user permissions
func (h *PingAMHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	h.logger.Debug("Getting user permissions from PingAM",
		logging.String("user_id", userID))

	permissions, err := h.pingamService.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user permissions",
			logging.Error(err),
			logging.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user permissions",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("User permissions retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// RefreshToken refreshes an access token
func (h *PingAMHandler) RefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body for token refresh",
			logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Refreshing token with PingAM")

	result, err := h.pingamService.RefreshToken(c.Request.Context(), request.RefreshToken)
	if err != nil {
		h.logger.Error("Token refresh failed",
			logging.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Token refresh failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Token refreshed successfully")

	c.JSON(http.StatusOK, gin.H{
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"token_type":    result.TokenType,
		"expires_in":    result.ExpiresIn,
		"scope":         result.Scope,
	})
}

// RevokeToken revokes an access token
func (h *PingAMHandler) RevokeToken(c *gin.Context) {
	var request struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.Error("Invalid request body for token revocation",
			logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Revoking token with PingAM")

	err := h.pingamService.RevokeToken(c.Request.Context(), request.Token)
	if err != nil {
		h.logger.Error("Token revocation failed",
			logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Token revocation failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Token revoked successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Token revoked successfully",
	})
}

// HealthCheck performs a health check against PingAM
func (h *PingAMHandler) HealthCheck(c *gin.Context) {
	h.logger.Debug("Performing PingAM health check")

	err := h.pingamService.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error("PingAM health check failed",
			logging.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "PingAM health check failed",
			"details": err.Error(),
		})
		return
	}

	h.logger.Debug("PingAM health check successful")

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "pingam",
		"timestamp": time.Now(),
	})
}

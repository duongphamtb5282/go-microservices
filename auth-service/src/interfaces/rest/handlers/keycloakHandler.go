package handlers

import (
	"net/http"

	"auth-service/src/applications/services"
	"auth-service/src/infrastructure/identity/models"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// KeycloakHandler handles Keycloak-related HTTP requests
type KeycloakHandler struct {
	service *services.KeycloakApplicationService
	logger  *logging.Logger
}

// NewKeycloakHandler creates a new Keycloak handler
func NewKeycloakHandler(service *services.KeycloakApplicationService, logger *logging.Logger) *KeycloakHandler {
	return &KeycloakHandler{
		service: service,
		logger:  logger,
	}
}

// Authenticate handles user authentication with Keycloak
func (h *KeycloakHandler) Authenticate(c *gin.Context) {
	var credentials models.Credentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.service.Authenticate(c.Request.Context(), credentials)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ValidateToken handles token validation
func (h *KeycloakHandler) ValidateToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	tokenInfo, err := h.service.ValidateToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	c.JSON(http.StatusOK, tokenInfo)
}

// GetUserProfile handles user profile retrieval
func (h *KeycloakHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	profile, err := h.service.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// CheckPermission handles permission checking
func (h *KeycloakHandler) CheckPermission(c *gin.Context) {
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	allowed, err := h.service.CheckPermission(c.Request.Context(), req.UserID, req.Resource, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Permission check failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"allowed": allowed})
}

// GetUserRoles handles user roles retrieval
func (h *KeycloakHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	roles, err := h.service.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetUserPermissions handles user permissions retrieval
func (h *KeycloakHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	permissions, err := h.service.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// RefreshToken handles token refresh
func (h *KeycloakHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token refresh failed"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RevokeToken handles token revocation
func (h *KeycloakHandler) RevokeToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := h.service.RevokeToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token revocation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token revoked successfully"})
}

// InitiateSSOLogin handles SSO login initiation
func (h *KeycloakHandler) InitiateSSOLogin(c *gin.Context) {
	provider := c.Query("provider")

	authURL, err := h.service.InitiateSSOLogin(c.Request.Context(), provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO initiation failed"})
		return
	}

	c.JSON(http.StatusOK, authURL)
}

// InitiateMFA handles MFA challenge initiation
func (h *KeycloakHandler) InitiateMFA(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	challenge, err := h.service.InitiateMFA(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MFA initiation failed"})
		return
	}

	c.JSON(http.StatusOK, challenge)
}

// VerifyMFA handles MFA verification
func (h *KeycloakHandler) VerifyMFA(c *gin.Context) {
	var req struct {
		ChallengeID string `json:"challenge_id" binding:"required"`
		Code        string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	verification, err := h.service.VerifyMFA(c.Request.Context(), req.ChallengeID, req.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "MFA verification failed"})
		return
	}

	c.JSON(http.StatusOK, verification)
}

// HealthCheck handles Keycloak health check
func (h *KeycloakHandler) HealthCheck(c *gin.Context) {
	err := h.service.HealthCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

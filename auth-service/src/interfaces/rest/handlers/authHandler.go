package handlers

import (
	"context"
	"fmt"
	"net/http"

	"auth-service/src/applications/commands"
	"auth-service/src/applications/services"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/identity/models"
	"auth-service/src/interfaces/rest/dto"
	"backend-core/logging"
	"backend-core/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IdentityProvider defines the interface for identity providers
type IdentityProvider interface {
	Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error)
	ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error)
	GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error)
}

// AuthHandler handles HTTP requests for authentication operations
type AuthHandler struct {
	userService      *services.UserApplicationService
	jwtManager       *security.JWTManager
	logger           *logging.Logger
	identityProvider IdentityProvider
	authConfig       *config.AuthorizationConfig
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(
	userService *services.UserApplicationService,
	jwtManager *security.JWTManager,
	logger *logging.Logger,
	identityProvider IdentityProvider,
	authConfig *config.AuthorizationConfig,
) *AuthHandler {
	return &AuthHandler{
		userService:      userService,
		jwtManager:       jwtManager,
		logger:           logger,
		identityProvider: identityProvider,
		authConfig:       authConfig,
	}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
		return
	}

	// Authenticate using the configured identity provider
	credentials := models.Credentials{
		Username: req.Email,
		Password: req.Password,
	}

	authResult, err := h.identityProvider.Authenticate(c.Request.Context(), credentials)
	if err != nil {
		h.logger.Error("Authentication failed",
			logging.Error(err),
			logging.String("username", req.Email))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token pair using authenticated user info
	accessToken, refreshToken, err := h.jwtManager.GenerateTokenPair(authResult.UserID, authResult.Username, "user") // TODO: Get role from auth result
	if err != nil {
		h.logger.Error("Failed to generate token pair", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	h.logger.Info("User logged in successfully",
		logging.String("user_id", authResult.UserID),
		logging.String("username", authResult.Username))

	// Create user DTO
	userDTO := dto.UserDTO{
		ID:       authResult.UserID,
		Username: authResult.Username,
		Email:    authResult.Username, // TODO: Get actual email from profile
	}

	response := dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds for access token
		User:         userDTO,
	}

	c.JSON(http.StatusOK, response)
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create user using UserApplicationService
	command := commands.NewCreateUserCommand(
		req.Username,
		req.Email,
		req.Password,
		"system", // TODO: Get from JWT token
	)

	// Execute command
	result, err := h.userService.CreateUser(c.Request.Context(), command)
	if err != nil {
		h.logger.Error("Failed to create user", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    result,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement logout logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ForgotPassword handles POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement forgot password logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent",
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement reset password logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// VerifyEmail handles POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement email verification logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ChangePassword handles POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement change password logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// RefreshToken handles POST /api/v1/auth/refresh-token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required"})
		return
	}

	// Validate refresh token and generate new token pair
	accessToken, refreshToken, err := h.jwtManager.RefreshTokenPair(req.RefreshToken)
	if err != nil {
		h.logger.Error("Failed to refresh token", logging.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	h.logger.Info("Token pair refreshed successfully")

	response := dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours in seconds for access token
	}

	c.JSON(http.StatusOK, response)
}

// RevokeToken handles POST /api/v1/auth/revoke-token
func (h *AuthHandler) RevokeToken(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement token revocation logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Token revoked successfully",
	})
}

// LogoutAll handles POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// TODO: Implement logout from all devices logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all devices successfully",
	})
}

// GetProfile handles GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// TODO: Implement get profile logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Get profile endpoint - TODO: Implement",
	})
}

// UpdateProfile handles PUT /api/v1/auth/profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// TODO: Implement update profile logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetSessions handles GET /api/v1/auth/sessions
func (h *AuthHandler) GetSessions(c *gin.Context) {
	// TODO: Implement get sessions logic
	sessions := []gin.H{
		{
			"id":            "session-1",
			"ip_address":    "192.168.1.1",
			"user_agent":    "Mozilla/5.0...",
			"created_at":    "2024-01-01T00:00:00Z",
			"last_accessed": "2024-01-01T00:00:00Z",
			"is_active":     true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// DatabaseIdentityProvider implements IdentityProvider using database authentication
type DatabaseIdentityProvider struct {
	userService *services.UserApplicationService
	logger      *logging.Logger
}

// NewDatabaseIdentityProvider creates a new database identity provider
func NewDatabaseIdentityProvider(userService *services.UserApplicationService, logger *logging.Logger) IdentityProvider {
	return &DatabaseIdentityProvider{
		userService: userService,
		logger:      logger,
	}
}

// Authenticate authenticates a user against the database
func (p *DatabaseIdentityProvider) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	// TODO: Implement proper database authentication
	// For now, accept any credentials for testing
	if credentials.Username == "" || credentials.Password == "" {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Generate a mock user ID
	userID := uuid.New().String()

	p.logger.Info("Database authentication successful",
		logging.String("username", credentials.Username),
		logging.String("user_id", userID))

	return &models.AuthResult{
		UserID:   userID,
		Username: credentials.Username,
		// TODO: Add roles and groups when AuthResult struct is extended
	}, nil
}

// ValidateToken validates a JWT token
func (p *DatabaseIdentityProvider) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	// TODO: Implement token validation
	return &models.TokenInfo{
		Active: true,
		UserID: "user-123", // TODO: Extract from token
	}, nil
}

// GetUserProfile retrieves user profile from database
func (p *DatabaseIdentityProvider) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	// TODO: Implement profile retrieval
	return &models.UserProfile{
		ID:       userID,
		Username: "user", // TODO: Get from database
		Email:    "user@example.com", // TODO: Get from database
	}, nil
}

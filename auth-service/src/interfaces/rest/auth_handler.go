package handlers

import (
	"auth-service/src/interfaces/rest/dto"
	"auth-service/src/applications/services"
	"backend-core/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
	logger      *logging.Logger
}

func NewAuthHandler(authService *services.AuthService, logger *logging.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with username/email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	loginReq := req.(*dto.LoginRequest)

	// Call auth service
	response, err := h.authService.Login(c.Request.Context(), loginReq)
	if err != nil {
		h.logger.Error("Login failed", logging.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user with username, email, and password
// @Tags auth
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequest true "Registration information"
// @Success 201 {object} dto.RegisterResponse "Registration successful"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	registerReq := req.(*dto.RegisterRequest)

	// Call auth service
	response, err := h.authService.Register(c.Request.Context(), registerReq)
	if err != nil {
		h.logger.Error("Registration failed", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Registration failed",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	logoutReq := req.(*dto.LogoutRequest)

	// Call auth service
	err := h.authService.Logout(c.Request.Context(), logoutReq)
	if err != nil {
		h.logger.Error("Logout failed", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Logout failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// ForgotPassword handles forgot password request
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	forgotReq := req.(*dto.ForgotPasswordRequest)

	// Call auth service
	err := h.authService.ForgotPassword(c.Request.Context(), forgotReq)
	if err != nil {
		h.logger.Error("Forgot password failed", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent",
	})
}

// ResetPassword handles password reset
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	resetReq := req.(*dto.ResetPasswordRequest)

	// Call auth service
	err := h.authService.ResetPassword(c.Request.Context(), resetReq)
	if err != nil {
		h.logger.Error("Password reset failed", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password reset failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

// VerifyEmail handles email verification
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	verifyReq := req.(*dto.VerifyEmailRequest)

	// Call auth service
	err := h.authService.VerifyEmail(c.Request.Context(), verifyReq)
	if err != nil {
		h.logger.Error("Email verification failed", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email verification failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	changeReq := req.(*dto.ChangePasswordRequest)

	// Get user ID from JWT token (TODO: implement proper JWT extraction)
	userID := "user-id-from-jwt" // This should come from JWT middleware

	// Call auth service
	err := h.authService.ChangePassword(c.Request.Context(), userID, changeReq)
	if err != nil {
		h.logger.Error("Password change failed", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password change failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	refreshReq := req.(*dto.RefreshTokenRequest)

	// Call auth service
	response, err := h.authService.RefreshToken(c.Request.Context(), refreshReq)
	if err != nil {
		h.logger.Error("Token refresh failed", logging.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Token refresh failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RevokeToken handles token revocation
func (h *AuthHandler) RevokeToken(c *gin.Context) {
	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	revokeReq := req.(*dto.LogoutRequest) // Reuse logout request structure

	// Call auth service
	err := h.authService.Logout(c.Request.Context(), revokeReq)
	if err != nil {
		h.logger.Error("Token revocation failed", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Token revocation failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token revoked successfully",
	})
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Get user ID from JWT token (TODO: implement proper JWT extraction)
	_ = "user-id-from-jwt" // This should come from JWT middleware

	// TODO: Implement logout from all devices
	// For now, just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out from all devices successfully",
	})
}

// GetProfile handles getting user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from JWT token (TODO: implement proper JWT extraction)
	userID := "user-id-from-jwt" // This should come from JWT middleware

	// Call auth service
	profile, err := h.authService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get profile", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get profile",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile handles updating user profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from JWT token (TODO: implement proper JWT extraction)
	userID := "user-id-from-jwt" // This should come from JWT middleware

	// Get validated request from middleware
	req, exists := c.Get("validated_model")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request validation failed",
		})
		return
	}

	updateReq := req.(*dto.UserUpdateRequest)

	// Call auth service
	profile, err := h.authService.UpdateUserProfile(c.Request.Context(), userID, updateReq)
	if err != nil {
		h.logger.Error("Failed to update profile", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// GetSessions handles getting user sessions
func (h *AuthHandler) GetSessions(c *gin.Context) {
	// Get user ID from JWT token (TODO: implement proper JWT extraction)
	_ = "user-id-from-jwt" // This should come from JWT middleware

	// TODO: Implement session management
	// For now, return empty sessions
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

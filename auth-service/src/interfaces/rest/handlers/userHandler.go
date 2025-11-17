package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"auth-service/src/applications/commands"
	"auth-service/src/applications/dto"
	"auth-service/src/applications/queries"
	"auth-service/src/applications/services"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserApplicationService
	logger      *logging.Logger
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService *services.UserApplicationService, logger *logging.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// CreateUser handles POST /api/v1/users with automatic validation
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest

	// Parse and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate using struct tags
	if err := h.validateCreateUserRequest(&req); err != nil {
		h.logger.Error("Validation failed", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Create command
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

	c.JSON(http.StatusCreated, result)
}

// GetUser handles GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Create query
	query := queries.NewGetUserQuery(userID)

	// Execute query
	result, err := h.userService.GetUser(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get user", logging.Error(err))
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Execute query
	result, err := h.userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		h.logger.Error("Failed to list users", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ActivateUser handles POST /api/v1/users/:id/activate
func (h *UserHandler) ActivateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Create command
	command := commands.NewActivateUserCommand(userID)

	// Execute command
	result, err := h.userService.ActivateUser(c.Request.Context(), command)
	if err != nil {
		h.logger.Error("Failed to activate user", logging.Error(err))
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// validateCreateUserRequest validates the create user request with comprehensive rules
func (h *UserHandler) validateCreateUserRequest(req *dto.CreateUserRequest) error {
	var errors []string

	// Validate username
	if err := h.validateUsername(req.Username); err != nil {
		errors = append(errors, fmt.Sprintf("username: %s", err.Error()))
	}

	// Validate email
	if err := h.validateEmail(req.Email); err != nil {
		errors = append(errors, fmt.Sprintf("email: %s", err.Error()))
	}

	// Validate password
	if err := h.validatePassword(req.Password); err != nil {
		errors = append(errors, fmt.Sprintf("password: %s", err.Error()))
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateUsername validates username according to business rules
func (h *UserHandler) validateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 20 {
		return fmt.Errorf("username must be at most 20 characters long")
	}

	// Check for alphanumeric characters only
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if !matched {
		return fmt.Errorf("username must contain only alphanumeric characters and underscores")
	}

	// Check for reserved usernames
	reservedUsernames := []string{"admin", "root", "system", "api", "www", "mail", "ftp", "test", "user", "guest"}
	for _, reserved := range reservedUsernames {
		if strings.ToLower(username) == reserved {
			return fmt.Errorf("username '%s' is reserved", username)
		}
	}

	return nil
}

// validateEmail validates email according to business rules
func (h *UserHandler) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 {
		return fmt.Errorf("email must be at most 254 characters long")
	}

	// Basic email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	// Check for common email providers (optional business rule)
	email = strings.ToLower(email)
	disallowedDomains := []string{"tempmail.com", "10minutemail.com", "guerrillamail.com"}
	for _, domain := range disallowedDomains {
		if strings.Contains(email, domain) {
			return fmt.Errorf("email domain '%s' is not allowed", domain)
		}
	}

	return nil
}

// validatePassword validates password according to business rules
func (h *UserHandler) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 100 {
		return fmt.Errorf("password must be at most 100 characters long")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	// Check for at least one special character
	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	// Check for common weak passwords
	weakPasswords := []string{"password", "123456", "12345678", "qwerty", "abc123", "password123", "admin", "letmein"}
	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			return fmt.Errorf("password is too common, please choose a stronger password")
		}
	}

	return nil
}

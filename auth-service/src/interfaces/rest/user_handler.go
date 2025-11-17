package handlers

import (
	"net/http"
	"strconv"

	"auth-service/src/applications"
	"auth-service/src/applications/commands"
	"auth-service/src/applications/dto"
	"auth-service/src/applications/queries"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	commandBus applications.CommandBus
	queryBus   applications.QueryBus
	logger     *logging.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	commandBus applications.CommandBus,
	queryBus applications.QueryBus,
	logger *logging.Logger,
) *UserHandler {
	return &UserHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// UserHandlerInstance creates a new user handler (without New prefix)
func UserHandlerInstance(
	commandBus interfaces.CommandBus,
	queryBus interfaces.QueryBus,
	logger *logging.Logger,
) *UserHandler {
	return &UserHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// CreateUser handles POST /api/v1/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create command
	command := commands.NewCreateUserCommand(
		req.Username,
		req.Email,
		req.Password,
		"system", // TODO: Get from JWT token
	)

	// Send command
	result, err := h.commandBus.Send(c.Request.Context(), command)
	if err != nil {
		h.logger.Error("Failed to create user", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	userDTO, ok := result.(dto.UserDTO)
	if !ok {
		h.logger.Error("Invalid result type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid result type"})
		return
	}

	// Convert to response
	response := dto.NewUserResponse(userDTO)

	c.JSON(http.StatusCreated, response)
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

	// Send query
	result, err := h.queryBus.Ask(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to get user", logging.Error(err))
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	userDTO, ok := result.(dto.UserDTO)
	if !ok {
		h.logger.Error("Invalid result type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid result type"})
		return
	}

	// Convert to response
	response := dto.NewUserResponse(userDTO)

	c.JSON(http.StatusOK, response)
}

// ListUsers handles GET /api/v1/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	// Create query
	query := queries.NewListUsersQuery(page, limit)

	// Send query
	result, err := h.queryBus.Ask(c.Request.Context(), query)
	if err != nil {
		h.logger.Error("Failed to list users", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	userListDTO, ok := result.(dto.UserListDTO)
	if !ok {
		h.logger.Error("Invalid result type")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid result type"})
		return
	}

	// Convert to response
	responses := dto.ToUserResponse(userListDTO.Users)
	response := dto.UserListResponse{
		Users: responses,
		Total: userListDTO.Total,
		Page:  userListDTO.Page,
		Limit: userListDTO.Limit,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create command
	command := commands.NewUpdateUserCommand(userID, "system") // TODO: Get from JWT token

	// Apply updates
	if req.Username != nil {
		command = command.WithUsername(*req.Username)
	}
	if req.Email != nil {
		command = command.WithEmail(*req.Email)
	}
	if req.Password != nil {
		command = command.WithPassword(*req.Password)
	}

	// Send command
	_, err := h.commandBus.Send(c.Request.Context(), command)
	if err != nil {
		h.logger.Error("Failed to update user", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser handles DELETE /api/v1/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Create command
	command := commands.NewDeleteUserCommand(userID, "system") // TODO: Get from JWT token

	// Send command
	_, err := h.commandBus.Send(c.Request.Context(), command)
	if err != nil {
		h.logger.Error("Failed to delete user", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// VerifyEmail handles email verification
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Verify email endpoint - TODO: implement"})
}

// ResendVerification handles resending verification email
func (h *UserHandler) ResendVerification(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Resend verification endpoint - TODO: implement"})
}

// ActivateUser handles user activation
func (h *UserHandler) ActivateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Activate user endpoint - TODO: implement"})
}

// DeactivateUser handles user deactivation
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Deactivate user endpoint - TODO: implement"})
}

// ChangePassword handles password change
func (h *UserHandler) ChangePassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Change password endpoint - TODO: implement"})
}

// ResetPassword handles password reset
func (h *UserHandler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Reset password endpoint - TODO: implement"})
}

// SearchUsers handles user search
func (h *UserHandler) SearchUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Search users endpoint - TODO: implement"})
}

// GetUserStats handles getting user statistics
func (h *UserHandler) GetUserStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get user stats endpoint - TODO: implement"})
}

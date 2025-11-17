package handlers

import (
	"net/http"
	"strconv"
	"time"

	"auth-service/src/applications/services"
	"auth-service/src/domain/entities"
	"auth-service/src/domain/valueObjects"
	"auth-service/src/applications/dto"

	"backend-core/middleware/exception"
	"backend-shared/errors"
	sharedModels "backend-shared/models"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// UserHandlerWithException demonstrates exception handling in handlers
type UserHandlerWithException struct {
	commandBus *services.CommandBus
	queryBus   *services.QueryBus
	logger     *logging.Logger
}

// NewUserHandlerWithException creates a new user handler with exception handling
func NewUserHandlerWithException(commandBus *services.CommandBus, queryBus *services.QueryBus, logger *logging.Logger) *UserHandlerWithException {
	return &UserHandlerWithException{
		commandBus: commandBus,
		queryBus:   queryBus,
		logger:     logger,
	}
}

// CreateUser creates a new user with proper exception handling
func (h *UserHandlerWithException) CreateUser(c *gin.Context) {
	var req models.UserCreateRequest

	// Parse and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use exception middleware helper for validation errors
		validationErr := errors.NewValidationError("request", req, "Invalid request format")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Validate business rules
	if err := h.validateUserRequest(&req); err != nil {
		exception.GinErrorResponse(c, err, http.StatusBadRequest)
		return
	}

	// Create user entity
	user, err := entity.NewUser(req.UserName, req.Email, req.PasswordHash, "system")
	if err != nil {
		exception.GinErrorResponse(c, err, http.StatusBadRequest)
		return
	}

	// Publish user created event
	userCreatedEvent := events.NewUserCreated(
		user.ID(),
		user.Username().String(),
		user.Email().String(),
	)

	// In a real implementation, you would publish this event
	h.logger.Info("User created event",
		logging.String("user_id", userCreatedEvent.UserID()),
		logging.String("username", userCreatedEvent.Username()),
		logging.String("email", userCreatedEvent.Email()),
	)

	// Convert to response format
	userResponse := models.UserResponse{
		ID:         user.ID().String(),
		UserName:   user.Username().String(),
		Email:      user.Email().String(),
		CreatedBy:  user.AuditInfo().CreatedBy,
		CreatedAt:  user.AuditInfo().CreatedAt,
		ModifiedBy: user.AuditInfo().CreatedBy, // Use CreatedBy as fallback
		ModifiedAt: user.AuditInfo().CreatedAt, // Use CreatedAt as fallback
	}

	// Return success response
	exception.GinSuccessResponse(c, userResponse, "User created successfully")
}

// GetUser retrieves a user by ID with exception handling
func (h *UserHandlerWithException) GetUser(c *gin.Context) {
	userID := c.Param("id")

	// Validate user ID format
	if userID == "" {
		validationErr := errors.NewValidationError("id", userID, "User ID is required")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Parse user ID
	entityID, err := sharedModels.NewEntityIDFromString(userID)
	if err != nil {
		validationErr := errors.NewValidationError("id", userID, "Invalid user ID format")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Mock user retrieval (in real implementation, use repository)
	user := h.mockGetUser(entityID)
	if user == nil {
		notFoundErr := errors.NewDomainError(errors.ErrCodeNotFound, "User not found")
		exception.GinErrorResponse(c, notFoundErr, http.StatusNotFound)
		return
	}

	// Convert to response format
	userResponse := models.UserResponse{
		ID:         user.ID().String(),
		UserName:   user.Username().String(),
		Email:      user.Email().String(),
		CreatedBy:  user.AuditInfo().CreatedBy,
		CreatedAt:  user.AuditInfo().CreatedAt,
		ModifiedBy: user.AuditInfo().CreatedBy, // Use CreatedBy as fallback
		ModifiedAt: user.AuditInfo().CreatedAt, // Use CreatedAt as fallback
	}

	// Return success response
	exception.GinSuccessResponse(c, userResponse, "User retrieved successfully")
}

// ListUsers lists users with pagination and exception handling
func (h *UserHandlerWithException) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		validationErr := errors.NewValidationError("page", pageStr, "Invalid page number")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		validationErr := errors.NewValidationError("limit", limitStr, "Invalid limit (must be 1-100)")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Mock user listing (in real implementation, use repository)
	users := h.mockListUsers(page, limit)

	// Return success response
	exception.GinSuccessResponse(c, users, "Users retrieved successfully")
}

// UpdateUser updates a user with exception handling
func (h *UserHandlerWithException) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	// Validate user ID
	if userID == "" {
		validationErr := errors.NewValidationError("id", userID, "User ID is required")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	var req models.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErr := errors.NewValidationError("request", req, "Invalid request format")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Mock user update (in real implementation, use repository)
	user := h.mockUpdateUser(userID, &req)
	if user == nil {
		notFoundErr := errors.NewDomainError(errors.ErrCodeNotFound, "User not found")
		exception.GinErrorResponse(c, notFoundErr, http.StatusNotFound)
		return
	}

	// Log user updated event
	h.logger.Info("User updated event",
		logging.String("user_id", user.ID().String()),
		logging.String("username", user.Username().String()),
		logging.String("email", user.Email().String()),
	)

	// Convert to response format
	userResponse := models.UserResponse{
		ID:         user.ID().String(),
		UserName:   user.Username().String(),
		Email:      user.Email().String(),
		CreatedBy:  user.AuditInfo().CreatedBy,
		CreatedAt:  user.AuditInfo().CreatedAt,
		ModifiedBy: user.AuditInfo().CreatedBy, // Use CreatedBy as fallback
		ModifiedAt: user.AuditInfo().CreatedAt, // Use CreatedAt as fallback
	}

	// Return success response
	exception.GinSuccessResponse(c, userResponse, "User retrieved successfully")
}

// DeleteUser deletes a user with exception handling
func (h *UserHandlerWithException) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Validate user ID
	if userID == "" {
		validationErr := errors.NewValidationError("id", userID, "User ID is required")
		exception.GinErrorResponse(c, validationErr, http.StatusBadRequest)
		return
	}

	// Mock user deletion (in real implementation, use repository)
	user := h.mockDeleteUser(userID)
	if user == nil {
		notFoundErr := errors.NewDomainError(errors.ErrCodeNotFound, "User not found")
		exception.GinErrorResponse(c, notFoundErr, http.StatusNotFound)
		return
	}

	// Publish user deleted event
	userDeletedEvent := events.NewUserDeleted(
		user.ID(),
		user.Username().String(),
		user.Email().String(),
	)

	h.logger.Info("User deleted event",
		logging.String("user_id", userDeletedEvent.UserID()),
		logging.String("username", userDeletedEvent.Username()),
		logging.String("email", userDeletedEvent.Email()),
	)

	// Return success response
	exception.GinSuccessResponse(c, gin.H{"message": "User deleted successfully"}, "User deleted successfully")
}

// Helper methods for validation and mocking

func (h *UserHandlerWithException) validateUserRequest(req *models.UserCreateRequest) error {
	// Validate username
	if len(req.UserName) < 3 {
		return errors.NewValidationError("user_name", req.UserName, "Username must be at least 3 characters")
	}

	// Validate email format
	emailVO, err := value_objects.NewEmail(req.Email)
	if err != nil {
		return errors.NewValidationError("email", req.Email, "Invalid email format")
	}

	// Validate password
	if len(req.PasswordHash) < 8 {
		return errors.NewValidationError("password_hash", "[REDACTED]", "Password must be at least 8 characters")
	}

	// Check if email already exists (mock)
	if h.mockEmailExists(emailVO.String()) {
		return errors.NewDomainError(errors.ErrCodeAlreadyExists, "Email already exists")
	}

	return nil
}

// Mock methods (replace with real repository calls)

func (h *UserHandlerWithException) mockGetUser(id sharedModels.EntityID) *entity.User {
	// Mock implementation - in real app, use repository
	if id.String() == "123e4567-e89b-12d3-a456-426614174000" {
		user, _ := entity.NewUser("john_doe", "john@example.com", "hashed_password", "system")
		return user
	}
	return nil
}

func (h *UserHandlerWithException) mockListUsers(page, limit int) models.UserListResponse {
	// Mock implementation - in real app, use repository
	users := []models.UserResponse{
		{
			ID:         "123e4567-e89b-12d3-a456-426614174000",
			UserName:   "john_doe",
			Email:      "john@example.com",
			CreatedBy:  "system",
			CreatedAt:  time.Now(),
			ModifiedBy: "system",   // Use system as fallback
			ModifiedAt: time.Now(), // Use current time as fallback
		},
	}

	return models.NewUserListResponse(users, page, limit, 1)
}

func (h *UserHandlerWithException) mockUpdateUser(userID string, req *models.UserUpdateRequest) *entity.User {
	// Mock implementation - in real app, use repository
	if userID == "123e4567-e89b-12d3-a456-426614174000" {
		user, _ := entity.NewUser("john_doe", "john@example.com", "hashed_password", "system")
		return user
	}
	return nil
}

func (h *UserHandlerWithException) mockDeleteUser(userID string) *entity.User {
	// Mock implementation - in real app, use repository
	if userID == "123e4567-e89b-12d3-a456-426614174000" {
		user, _ := entity.NewUser("john_doe", "john@example.com", "hashed_password", "system")
		return user
	}
	return nil
}

func (h *UserHandlerWithException) mockEmailExists(email string) bool {
	// Mock implementation - in real app, use repository
	return email == "existing@example.com"
}

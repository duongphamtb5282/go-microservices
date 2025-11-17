package middleware

import (
	"backend-core/logging"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware provides request validation functionality
type ValidationMiddleware struct {
	validator *validator.Validate
	logger    *logging.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(logger *logging.Logger) *ValidationMiddleware {
	validate := validator.New()

	// Register custom validators
	validate.RegisterValidation("password", validatePassword)
	validate.RegisterValidation("email", validateEmail)
	validate.RegisterValidation("username", validateUsername)

	return &ValidationMiddleware{
		validator: validate,
		logger:    logger,
	}
}

// ValidateRequest validates request body against a struct
func (v *ValidationMiddleware) ValidateRequest(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse JSON body
		if err := c.ShouldBindJSON(model); err != nil {
			v.logger.Error("Failed to parse JSON", logging.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate the struct
		if err := v.validator.Struct(model); err != nil {
			validationErrors := v.formatValidationErrors(err)
			v.logger.Error("Validation failed", logging.Any("errors", validationErrors))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated model in context for use in handlers
		c.Set("validated_model", model)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func (v *ValidationMiddleware) ValidateQuery(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind query parameters
		if err := c.ShouldBindQuery(model); err != nil {
			v.logger.Error("Failed to parse query parameters", logging.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid query parameters",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate the struct
		if err := v.validator.Struct(model); err != nil {
			validationErrors := v.formatValidationErrors(err)
			v.logger.Error("Query validation failed", logging.Any("errors", validationErrors))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Query validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_query", model)
		c.Next()
	}
}

// ValidatePath validates path parameters
func (v *ValidationMiddleware) ValidatePath(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind path parameters
		if err := c.ShouldBindUri(model); err != nil {
			v.logger.Error("Failed to parse path parameters", logging.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid path parameters",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate the struct
		if err := v.validator.Struct(model); err != nil {
			validationErrors := v.formatValidationErrors(err)
			v.logger.Error("Path validation failed", logging.Any("errors", validationErrors))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Path validation failed",
				"details": validationErrors,
			})
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_path", model)
		c.Next()
	}
}

// formatValidationErrors formats validation errors into a readable format
func (v *ValidationMiddleware) formatValidationErrors(err error) map[string]interface{} {
	errors := make(map[string]interface{})

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			errors[field] = map[string]interface{}{
				"tag":   e.Tag(),
				"value": e.Value(),
				"param": e.Param(),
			}
		}
	}

	return errors
}

// Custom validation functions

// validatePassword validates password strength
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Check minimum length
	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// validateEmail validates email format
func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Basic email validation
	if len(email) == 0 {
		return false
	}

	// Check for @ symbol
	if !strings.Contains(email, "@") {
		return false
	}

	// Split by @ and check parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]

	// Check local part
	if len(local) == 0 || len(local) > 64 {
		return false
	}

	// Check domain part
	if len(domain) == 0 || len(domain) > 255 {
		return false
	}

	// Check for valid domain format
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// validateUsername validates username format
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()

	// Check length
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	// Check for valid characters (alphanumeric and underscore)
	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_') {
			return false
		}
	}

	// Must start with letter or underscore
	firstChar := username[0]
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_') {
		return false
	}

	return true
}

// Response validation helpers

// ValidateResponse validates response data before sending
func (v *ValidationMiddleware) ValidateResponse(data interface{}) error {
	return v.validator.Struct(data)
}

// SanitizeResponse sanitizes response data
func (v *ValidationMiddleware) SanitizeResponse(data interface{}) interface{} {
	// Convert to JSON and back to remove any sensitive fields
	jsonData, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var sanitized interface{}
	if err := json.Unmarshal(jsonData, &sanitized); err != nil {
		return data
	}

	return sanitized
}

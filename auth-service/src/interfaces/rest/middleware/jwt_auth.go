package middleware

import (
	"net/http"
	"strings"

	"backend-core/logging"
	"backend-core/security"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware creates JWT authentication middleware using backend-core/security
func JWTAuthMiddleware(jwtManager *security.JWTManager, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "AUTH_HEADER_MISSING",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if token == "" {
			logger.Warn("Empty token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token is required",
				"code":  "EMPTY_TOKEN",
			})
			c.Abort()
			return
		}

		// Validate access token using backend-core JWT manager
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			logger.Warn("Access token validation failed",
				logging.Error(err),
				logging.String("token_prefix", token[:min(len(token), 10)]+"..."))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid, expired, or incorrect token type",
				"code":    "INVALID_ACCESS_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if token is expired
		if jwtManager.IsTokenExpired(token) {
			logger.Warn("Token expired",
				logging.String("user_id", claims.UserID),
				logging.String("username", claims.Username))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token expired",
				"code":  "TOKEN_EXPIRED",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		logger.Debug("JWT authentication successful",
			logging.String("user_id", claims.UserID),
			logging.String("username", claims.Username),
			logging.String("role", claims.Role))

		c.Next()
	}
}

// JWTAuthOptional creates optional JWT authentication middleware
// If token is present and valid, user info is set in context
// If token is missing or invalid, request continues without authentication
func JWTAuthOptional(jwtManager *security.JWTManager, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		token := tokenParts[1]
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			logger.Debug("Optional JWT access token validation failed", logging.Error(err))
			c.Next()
			return
		}

		// Valid token, set user info
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_role", claims.Role)
		c.Set("user_claims", claims)

		logger.Debug("Optional JWT authentication successful",
			logging.String("user_id", claims.UserID))

		c.Next()
	}
}

// RequireRole middleware ensures user has specific role
func RequireRole(requiredRole string, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			logger.Warn("User claims not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*security.Claims)
		if !ok {
			logger.Error("Invalid claims type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		if userClaims.Role != requiredRole {
			logger.Warn("Insufficient permissions",
				logging.String("user_id", userClaims.UserID),
				logging.String("user_role", userClaims.Role),
				logging.String("required_role", requiredRole))
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"code":          "FORBIDDEN",
				"required_role": requiredRole,
				"user_role":     userClaims.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware ensures user has at least one of the specified roles
func RequireAnyRole(allowedRoles []string, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			logger.Warn("User claims not found in context")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*security.Claims)
		if !ok {
			logger.Error("Invalid claims type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
				"code":  "INVALID_CLAIMS",
			})
			c.Abort()
			return
		}

		hasRole := false
		for _, role := range allowedRoles {
			if userClaims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			logger.Warn("Insufficient permissions",
				logging.String("user_id", userClaims.UserID),
				logging.String("user_role", userClaims.Role),
				logging.Any("allowed_roles", allowedRoles))
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Insufficient permissions",
				"code":          "FORBIDDEN",
				"allowed_roles": allowedRoles,
				"user_role":     userClaims.Role,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserClaims helper function to extract user claims from context
func GetUserClaims(c *gin.Context) (*security.Claims, error) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, ErrUserClaimsNotFound
	}

	userClaims, ok := claims.(*security.Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return userClaims, nil
}

// GetUserID helper function to extract user ID from context
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", ErrUserIDNotFound
	}

	id, ok := userID.(string)
	if !ok {
		return "", ErrInvalidUserID
	}

	return id, nil
}

// Custom errors
var (
	ErrUserClaimsNotFound = &AuthError{Code: "USER_CLAIMS_NOT_FOUND", Message: "User claims not found in context"}
	ErrInvalidClaims      = &AuthError{Code: "INVALID_CLAIMS", Message: "Invalid claims type"}
	ErrUserIDNotFound     = &AuthError{Code: "USER_ID_NOT_FOUND", Message: "User ID not found in context"}
	ErrInvalidUserID      = &AuthError{Code: "INVALID_USER_ID", Message: "Invalid user ID type"}
)

// AuthError represents an authentication error
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

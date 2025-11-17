package middleware

import (
	"net/http"
	"strings"

	"auth-service/src/applications/services"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// PingAMAuth middleware for PingAM authentication
func PingAMAuth(pingamService *services.PingAMApplicationService, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Error("Authorization header required")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Check if token starts with "Bearer "
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Error("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token with PingAM
		session, err := pingamService.ValidateSession(c.Request.Context(), token)
		if err != nil {
			logger.Error("Token validation failed",
				logging.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"code":    "INVALID_TOKEN",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", session.UserID)
		c.Set("session_id", session.ID)
		c.Set("session", session)
		c.Set("auth_source", "pingam")

		logger.Debug("User authenticated with PingAM",
			logging.String("user_id", session.UserID),
			logging.String("session_id", session.ID))

		c.Next()
	}
}

// PingAMAuthz middleware for PingAM authorization
func PingAMAuthz(pingamService *services.PingAMApplicationService, resource, action string, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Check permission with PingAM
		allowed, err := pingamService.CheckPermission(c.Request.Context(), userID.(string), resource, action)
		if err != nil {
			logger.Error("Permission check failed",
				logging.Error(err),
				logging.String("user_id", userID.(string)),
				logging.String("resource", resource),
				logging.String("action", action))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Permission check failed",
				"code":    "PERMISSION_CHECK_FAILED",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !allowed {
			logger.Warn("Access denied",
				logging.String("user_id", userID.(string)),
				logging.String("resource", resource),
				logging.String("action", action))
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "Access denied",
				"code":     "ACCESS_DENIED",
				"user_id":  userID,
				"resource": resource,
				"action":   action,
			})
			c.Abort()
			return
		}

		// Set authorization context
		c.Set("authorized_resource", resource)
		c.Set("authorized_action", action)

		logger.Debug("User authorized with PingAM",
			logging.String("user_id", userID.(string)),
			logging.String("resource", resource),
			logging.String("action", action))

		c.Next()
	}
}

// PingAMRoleAuthz middleware for role-based authorization
func PingAMRoleAuthz(pingamService *services.PingAMApplicationService, requiredRoles []string, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Get user roles from PingAM
		userRoles, err := pingamService.GetUserRoles(c.Request.Context(), userID.(string))
		if err != nil {
			logger.Error("Failed to get user roles",
				logging.Error(err),
				logging.String("user_id", userID.(string)))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get user roles",
				"code":    "ROLE_CHECK_FAILED",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			logger.Warn("Insufficient role permissions",
				logging.String("user_id", userID.(string)),
				logging.Any("user_roles", userRoles),
				logging.Any("required_roles", requiredRoles))
			c.JSON(http.StatusForbidden, gin.H{
				"error":          "Insufficient role permissions",
				"code":           "INSUFFICIENT_ROLE_PERMISSIONS",
				"user_id":        userID,
				"user_roles":     userRoles,
				"required_roles": requiredRoles,
			})
			c.Abort()
			return
		}

		// Set role context
		c.Set("user_roles", userRoles)
		c.Set("required_roles", requiredRoles)

		logger.Debug("User authorized with required role",
			logging.String("user_id", userID.(string)),
			logging.Any("user_roles", userRoles),
			logging.Any("required_roles", requiredRoles))

		c.Next()
	}
}

// PingAMPermissionAuthz middleware for permission-based authorization
func PingAMPermissionAuthz(pingamService *services.PingAMApplicationService, requiredPermissions []string, logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context
		userID, exists := c.Get("user_id")
		if !exists {
			logger.Error("User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
				"code":  "USER_NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Get user permissions from PingAM
		userPermissions, err := pingamService.GetUserPermissions(c.Request.Context(), userID.(string))
		if err != nil {
			logger.Error("Failed to get user permissions",
				logging.Error(err),
				logging.String("user_id", userID.(string)))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get user permissions",
				"code":    "PERMISSION_CHECK_FAILED",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Check if user has all required permissions
		hasAllPermissions := true
		missingPermissions := []string{}

		for _, requiredPermission := range requiredPermissions {
			hasPermission := false
			for _, userPermission := range userPermissions {
				if userPermission == requiredPermission {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				hasAllPermissions = false
				missingPermissions = append(missingPermissions, requiredPermission)
			}
		}

		if !hasAllPermissions {
			logger.Warn("Insufficient permissions",
				logging.String("user_id", userID.(string)),
				logging.Any("user_permissions", userPermissions),
				logging.Any("required_permissions", requiredPermissions),
				logging.Any("missing_permissions", missingPermissions))
			c.JSON(http.StatusForbidden, gin.H{
				"error":                "Insufficient permissions",
				"code":                 "INSUFFICIENT_PERMISSIONS",
				"user_id":              userID,
				"user_permissions":     userPermissions,
				"required_permissions": requiredPermissions,
				"missing_permissions":  missingPermissions,
			})
			c.Abort()
			return
		}

		// Set permission context
		c.Set("user_permissions", userPermissions)
		c.Set("required_permissions", requiredPermissions)

		logger.Debug("User authorized with required permissions",
			logging.String("user_id", userID.(string)),
			logging.Any("user_permissions", userPermissions),
			logging.Any("required_permissions", requiredPermissions))

		c.Next()
	}
}

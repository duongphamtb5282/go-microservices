package middleware

import (
	"context"
	"net/http"

	"auth-service/src/domain/authorization"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/identity/keycloak"
	"backend-core/logging"
	"backend-core/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UnifiedAuthorizationMiddleware handles authorization using JWT, JWT with DB, or Keycloak
type UnifiedAuthorizationMiddleware struct {
	authConfig      *config.AuthorizationConfig
	jwtManager      *security.JWTManager
	keycloakAdapter *keycloak.KeycloakAdapter
	roleRepo        authorization.RoleRepository
	permissionRepo  authorization.PermissionRepository
	logger          *logging.Logger
}

// NewUnifiedAuthorizationMiddleware creates a new unified authorization middleware
func NewUnifiedAuthorizationMiddleware(
	authConfig *config.AuthorizationConfig,
	jwtManager *security.JWTManager,
	keycloakAdapter *keycloak.KeycloakAdapter,
	roleRepo authorization.RoleRepository,
	permissionRepo authorization.PermissionRepository,
	logger *logging.Logger,
) *UnifiedAuthorizationMiddleware {
	return &UnifiedAuthorizationMiddleware{
		authConfig:      authConfig,
		jwtManager:      jwtManager,
		keycloakAdapter: keycloakAdapter,
		roleRepo:        roleRepo,
		permissionRepo:  permissionRepo,
		logger:          logger,
	}
}

// RequirePermission checks if user has required permission based on configured mode
func (m *UnifiedAuthorizationMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if authorization is enabled
		if !m.authConfig.Enabled {
			m.logger.Debug("Authorization is disabled, allowing request")
			c.Next()
			return
		}

		// Get user_id from context (set by JWT auth middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			m.logger.Warn("User ID not found in context for authorization check")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Invalid user ID type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
			})
			c.Abort()
			return
		}

		// Route to appropriate authorization method based on mode
		m.logger.Info("Authorization mode check",
			logging.String("mode", string(m.authConfig.Mode)),
			logging.String("resource", resource),
			logging.String("action", action))

		switch m.authConfig.Mode {
		case config.AuthorizationModeJWT:
			m.handleJWTAuthorization(c, userIDStr, resource, action)
		case config.AuthorizationModeJWTWithDB:
			m.handleJWTWithDBAuthorization(c, userIDStr, resource, action)
		case config.AuthorizationModeKeycloak:
			m.handleKeycloakAuthorization(c, userIDStr, resource, action)
		default:
			m.logger.Error("Unknown authorization mode", logging.String("mode", string(m.authConfig.Mode)))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid authorization configuration",
			})
			c.Abort()
		}
	}
}

// handleJWTAuthorization checks permissions from JWT claims only
func (m *UnifiedAuthorizationMiddleware) handleJWTAuthorization(c *gin.Context, userID, resource, action string) {
	m.logger.Info("Checking permission using JWT claims",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	// Get permissions from context (set by JWT middleware)
	permissions, exists := c.Get("permissions")
	if !exists || permissions == nil {
		m.logger.Warn("No permissions found in JWT token",
			logging.String("user_id", userID))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "No permissions found in token",
		})
		c.Abort()
		return
	}

	// Convert permissions to slice
	permSlice := convertToStringSlice(permissions)
	if permSlice == nil {
		m.logger.Error("Invalid permissions type in context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid token permissions format",
		})
		c.Abort()
		return
	}

	// Check if required permission exists
	if !hasPermission(permSlice, resource, action) {
		m.logger.Warn("Permission denied (JWT)",
			logging.String("user_id", userID),
			logging.String("required", resource+":"+action))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to perform this action",
			"details": gin.H{
				"required_permission": resource + ":" + action,
				"your_permissions":    permSlice,
			},
		})
		c.Abort()
		return
	}

	m.logger.Info("Permission granted (JWT)",
		logging.String("user_id", userID),
		logging.String("permission", resource+":"+action))
	c.Set("authorization_mode", "jwt")
	c.Next()
}

// handleJWTWithDBAuthorization checks permissions from database
func (m *UnifiedAuthorizationMiddleware) handleJWTWithDBAuthorization(c *gin.Context, userID, resource, action string) {
	m.logger.Info("Checking permission using JWT with Database",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	ctx := context.Background()

	// Parse user ID to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		m.logger.Error("Invalid user ID format", logging.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		c.Abort()
		return
	}

	// Check if we should use roles or permissions
	if m.authConfig.JWTWithDBAuth.UsePermissions {
		// Check permission from database
		hasPermission, err := m.permissionRepo.CheckUserPermission(ctx, userUUID, resource, action)
		if err != nil {
			m.logger.Error("Failed to check permission from database",
				logging.Error(err),
				logging.String("user_id", userID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to verify permissions",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			m.logger.Warn("Permission denied (JWT with DB)",
				logging.String("user_id", userID),
				logging.String("resource", resource),
				logging.String("action", action))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You do not have permission to perform this action",
			})
			c.Abort()
			return
		}
	}

	m.logger.Info("Permission granted (JWT with DB)",
		logging.String("user_id", userID))
	c.Set("authorization_mode", "jwt_with_db")
	c.Next()
}

// handleKeycloakAuthorization checks permissions with Keycloak
func (m *UnifiedAuthorizationMiddleware) handleKeycloakAuthorization(c *gin.Context, userID, resource, action string) {
	m.logger.Info("Checking permission with Keycloak",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	ctx := context.Background()

	// Check permission with Keycloak
	allowed, err := m.keycloakAdapter.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		m.logger.Error("Failed to check permission with Keycloak",
			logging.Error(err),
			logging.String("user_id", userID))

		// Fallback to JWT if configured
		if m.authConfig.KeycloakAuth.FallbackToJWT {
			m.logger.Info("Falling back to JWT authorization")
			m.handleJWTAuthorization(c, userID, resource, action)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Authorization Check Failed",
			"message": "Failed to verify permissions with Keycloak",
		})
		c.Abort()
		return
	}

	if !allowed {
		m.logger.Warn("Permission denied (Keycloak)",
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to perform this action",
		})
		c.Abort()
		return
	}

	m.logger.Info("Permission granted (Keycloak)",
		logging.String("user_id", userID))
	c.Set("authorization_mode", "keycloak")
	c.Next()
}

// RequireRole checks if user has a specific role
func (m *UnifiedAuthorizationMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.authConfig.Enabled {
			c.Next()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDStr := userID.(string)

		switch m.authConfig.Mode {
		case config.AuthorizationModeJWT:
			m.handleJWTRoleCheck(c, userIDStr, role)
		case config.AuthorizationModeJWTWithDB:
			m.handleJWTWithDBRoleCheck(c, userIDStr, role)
		case config.AuthorizationModeKeycloak:
			m.handleKeycloakRoleCheck(c, userIDStr, role)
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authorization mode"})
			c.Abort()
		}
	}
}

// handleJWTRoleCheck checks role from JWT claims
func (m *UnifiedAuthorizationMiddleware) handleJWTRoleCheck(c *gin.Context, userID, role string) {
	roles, exists := c.Get("roles")
	if !exists || roles == nil {
		m.logger.Warn("No roles found in JWT token", logging.String("user_id", userID))
		c.JSON(http.StatusForbidden, gin.H{"error": "No roles found in token"})
		c.Abort()
		return
	}

	roleSlice := convertToStringSlice(roles)
	if !hasRole(roleSlice, role) {
		m.logger.Warn("Required role not found (JWT)",
			logging.String("user_id", userID),
			logging.String("required_role", role))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Required role not found",
		})
		c.Abort()
		return
	}

	m.logger.Info("Role check passed (JWT)", logging.String("user_id", userID))
	c.Next()
}

// handleJWTWithDBRoleCheck checks role from database
func (m *UnifiedAuthorizationMiddleware) handleJWTWithDBRoleCheck(c *gin.Context, userID, role string) {
	ctx := context.Background()
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		c.Abort()
		return
	}

	roles, err := m.roleRepo.GetUserRoles(ctx, userUUID)
	if err != nil {
		m.logger.Error("Failed to get user roles from database", logging.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
		c.Abort()
		return
	}

	hasRequiredRole := false
	for _, r := range roles {
		if r.Name == role {
			hasRequiredRole = true
			break
		}
	}

	if !hasRequiredRole {
		c.JSON(http.StatusForbidden, gin.H{"error": "Required role not found"})
		c.Abort()
		return
	}

	c.Next()
}

// handleKeycloakRoleCheck checks role from Keycloak
func (m *UnifiedAuthorizationMiddleware) handleKeycloakRoleCheck(c *gin.Context, userID, role string) {
	ctx := context.Background()
	roles, err := m.keycloakAdapter.GetUserRoles(ctx, userID)
	if err != nil {
		m.logger.Error("Failed to get user roles from Keycloak", logging.Error(err))

		// Fallback to JWT if configured
		if m.authConfig.KeycloakAuth.FallbackToJWT {
			m.handleJWTRoleCheck(c, userID, role)
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify roles"})
		c.Abort()
		return
	}

	if !containsString(roles, role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Required role not found"})
		c.Abort()
		return
	}

	c.Next()
}

// Helper functions

func convertToStringSlice(data interface{}) []string {
	if strSlice, ok := data.([]string); ok {
		return strSlice
	}

	if interfaceSlice, ok := data.([]interface{}); ok {
		result := make([]string, 0, len(interfaceSlice))
		for _, item := range interfaceSlice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	return nil
}

func hasPermission(permissions []string, resource, action string) bool {
	requiredPermission := resource + ":" + action
	for _, perm := range permissions {
		if perm == requiredPermission || perm == resource+":*" || perm == "*:*" {
			return true
		}
	}
	return false
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

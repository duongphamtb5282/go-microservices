package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"auth-service/src/infrastructure/identity/keycloak"
	"auth-service/src/infrastructure/identity/pingam"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// AuthorizationMiddleware handles PingAM authorization checks
type AuthorizationMiddleware struct {
	pingamAdapter *pingam.PingAMAdapter
	logger        *logging.Logger
}

// NewAuthorizationMiddleware creates a new authorization middleware
func NewAuthorizationMiddleware(pingamAdapter *pingam.PingAMAdapter, logger *logging.Logger) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		pingamAdapter: pingamAdapter,
		logger:        logger,
	}
}

// RequirePermission returns a middleware that checks if user has required permission
func (m *AuthorizationMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID from context (set by authentication middleware)
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

		// Extract access token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Authorization header missing")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Missing authorization token",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Check permission with PingAM
		ctx := context.Background()
		userIDStr, ok := userID.(string)
		if !ok {
			m.logger.Error("Invalid user ID type in context")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "Invalid user ID",
			})
			c.Abort()
			return
		}

		m.logger.Info("Checking permission with PingAM",
			logging.String("user_id", userIDStr),
			logging.String("resource", resource),
			logging.String("action", action))

		// Call PingAM adapter to check permission
		allowed, err := m.pingamAdapter.CheckPermission(ctx, userIDStr, resource, action)
		if err != nil {
			m.logger.Error("Failed to check permission with PingAM",
				logging.Error(err),
				logging.String("user_id", userIDStr),
				logging.String("resource", resource),
				logging.String("action", action))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Authorization Check Failed",
				"message": "Failed to verify permissions",
			})
			c.Abort()
			return
		}

		// Check if permission is granted
		if !allowed {
			m.logger.Warn("Permission denied",
				logging.String("user_id", userIDStr),
				logging.String("resource", resource),
				logging.String("action", action))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You do not have permission to perform this action",
				"details": gin.H{
					"required_permission": resource + ":" + action,
				},
			})
			c.Abort()
			return
		}

		m.logger.Info("Permission granted",
			logging.String("user_id", userIDStr),
			logging.String("resource", resource),
			logging.String("action", action))

		// Store permission info in context for later use
		c.Set("permission_allowed", allowed)
		c.Set("resource", resource)
		c.Set("action", action)

		c.Next()
	}
}

// RequireAnyPermission checks if user has at least one of the required permissions
func (m *AuthorizationMiddleware) RequireAnyPermission(permissions map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		ctx := context.Background()
		userIDStr := userID.(string)

		// Check each permission
		for resource, action := range permissions {
			allowed, err := m.pingamAdapter.CheckPermission(ctx, userIDStr, resource, action)
			if err == nil && allowed {
				m.logger.Info("Permission granted (any)",
					logging.String("user_id", userIDStr),
					logging.String("resource", resource),
					logging.String("action", action))
				c.Set("permission_allowed", allowed)
				c.Next()
				return
			}
		}

		// No permission matched
		m.logger.Warn("No matching permission found",
			logging.String("user_id", userIDStr))
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You do not have permission to perform this action",
		})
		c.Abort()
	}
}

// RequireAllPermissions checks if user has all of the required permissions
func (m *AuthorizationMiddleware) RequireAllPermissions(permissions map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		ctx := context.Background()
		userIDStr := userID.(string)

		// Check each permission
		for resource, action := range permissions {
			allowed, err := m.pingamAdapter.CheckPermission(ctx, userIDStr, resource, action)
			if err != nil || !allowed {
				m.logger.Warn("Permission denied (all required)",
					logging.String("user_id", userIDStr),
					logging.String("resource", resource),
					logging.String("action", action))
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Forbidden",
					"message": "You do not have all required permissions",
				})
				c.Abort()
				return
			}
		}

		m.logger.Info("All permissions granted",
			logging.String("user_id", userIDStr))
		c.Next()
	}
}

// RequireRole checks if user has a specific role
func (m *AuthorizationMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		ctx := context.Background()
		userIDStr := userID.(string)

		// Get user roles from PingAM
		roles, err := m.pingamAdapter.GetUserRoles(ctx, userIDStr)
		if err != nil {
			m.logger.Error("Failed to get user roles",
				logging.Error(err),
				logging.String("user_id", userIDStr))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to verify roles",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		hasRole := false
		for _, userRole := range roles {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.logger.Warn("Required role not found",
				logging.String("user_id", userIDStr),
				logging.String("required_role", role))
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Required role not found",
				"details": gin.H{
					"required_role": role,
					"user_roles":    roles,
				},
			})
			c.Abort()
			return
		}

		m.logger.Info("Role check passed",
			logging.String("user_id", userIDStr),
			logging.String("role", role))
		c.Set("user_roles", roles)
		c.Next()
	}
}

// KeycloakAuthorizationMiddleware handles Keycloak role-based authorization
type KeycloakAuthorizationMiddleware struct {
	keycloakAdapter *keycloak.KeycloakAdapter
	logger          *logging.Logger
}

// NewKeycloakAuthorizationMiddleware creates a new Keycloak authorization middleware
func NewKeycloakAuthorizationMiddleware(keycloakAdapter *keycloak.KeycloakAdapter, logger *logging.Logger) *KeycloakAuthorizationMiddleware {
	return &KeycloakAuthorizationMiddleware{
		keycloakAdapter: keycloakAdapter,
		logger:          logger,
	}
}

// RequireValidKeycloakToken validates Keycloak JWT token and sets claims in context
func (m *KeycloakAuthorizationMiddleware) RequireValidKeycloakToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Keycloak token validation middleware called for path: %s\n", c.Request.URL.Path)

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Printf("DEBUG: No Authorization header\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Printf("DEBUG: Invalid Authorization header format\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			fmt.Printf("DEBUG: Empty token\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Token required",
			})
			c.Abort()
			return
		}

		// For now, we'll assume the token is valid and parse it
		// In a real implementation, you'd validate the token signature with Keycloak's public key
		// For this demo, we'll just decode the JWT payload
		parts := strings.Split(tokenString, ".")
		if len(parts) != 3 {
			fmt.Printf("DEBUG: Invalid JWT format\n")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token format",
			})
			c.Abort()
			return
		}

		// Decode the payload (second part)
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			fmt.Printf("DEBUG: Failed to decode JWT payload: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token",
			})
			c.Abort()
			return
		}

		// Parse JSON claims
		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			fmt.Printf("DEBUG: Failed to parse JWT claims: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set claims in context for role checking middleware
		c.Set("jwt_claims", claims)
		fmt.Printf("DEBUG: JWT claims set in context\n")

		c.Next()
	}
}

// RequireKeycloakRole checks if user has a specific Keycloak realm role from JWT token
func (m *KeycloakAuthorizationMiddleware) RequireKeycloakRole(requiredRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("DEBUG: Keycloak role authorization middleware called for path: %s\n", c.Request.URL.Path)
		fmt.Printf("DEBUG: Required roles: %v\n", requiredRoles)

		// Extract roles from JWT token claims (set by token validation middleware)
		claims, exists := c.Get("jwt_claims")
		if !exists {
			fmt.Printf("DEBUG: JWT claims not found in context\n")
			m.logger.Warn("JWT claims not found in context for Keycloak role authorization")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "JWT claims not found",
			})
			c.Abort()
			return
		}

		fmt.Printf("DEBUG: JWT claims found, checking roles\n")

		// Extract realm roles from JWT claims
		// Keycloak stores roles in realm_access.roles and resource_access.{client}.roles
		var userRoles []string

		// Check realm_access.roles
		if realmAccess, ok := claims.(map[string]interface{})["realm_access"].(map[string]interface{}); ok {
			if roles, ok := realmAccess["roles"].([]interface{}); ok {
				for _, role := range roles {
					if roleStr, ok := role.(string); ok {
						userRoles = append(userRoles, roleStr)
					}
				}
			}
		}

		// Check resource_access.{client}.roles (client-specific roles)
		if resourceAccess, ok := claims.(map[string]interface{})["resource_access"].(map[string]interface{}); ok {
			// Look for client-specific roles (client ID should match the one configured)
			for _, clientRoles := range resourceAccess {
				if clientRoleMap, ok := clientRoles.(map[string]interface{}); ok {
					if roles, ok := clientRoleMap["roles"].([]interface{}); ok {
						for _, role := range roles {
							if roleStr, ok := role.(string); ok {
								userRoles = append(userRoles, roleStr)
							}
						}
					}
				}
			}
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
			m.logger.Warn("Keycloak role authorization failed",
				logging.String("required_roles", fmt.Sprintf("%v", requiredRoles)),
				logging.String("user_roles", fmt.Sprintf("%v", userRoles)))

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "Insufficient permissions - required role not found",
				"details": gin.H{
					"required_roles": requiredRoles,
					"user_roles":     userRoles,
				},
			})
			c.Abort()
			return
		}

		m.logger.Info("Keycloak role authorization passed",
			logging.String("required_roles", fmt.Sprintf("%v", requiredRoles)),
			logging.String("user_roles", fmt.Sprintf("%v", userRoles)))

		// Store role information in context for later use
		c.Set("user_roles", userRoles)
		c.Set("authorized_roles", requiredRoles)

		c.Next()
	}
}

// RequireKeycloakPermission checks permissions using Keycloak Authorization Services
func (m *KeycloakAuthorizationMiddleware) RequireKeycloakPermission(resource, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user ID from JWT claims
		claims, exists := c.Get("jwt_claims")
		if !exists {
			m.logger.Warn("JWT claims not found in context for Keycloak permission check")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "JWT claims not found",
			})
			c.Abort()
			return
		}

		// Extract subject (user ID) from claims
		var userID string
		if sub, ok := claims.(map[string]interface{})["sub"].(string); ok {
			userID = sub
		} else {
			m.logger.Warn("Subject (sub) claim not found in JWT for permission check")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid JWT token - missing subject",
			})
			c.Abort()
			return
		}

		// Extract access token from Authorization header for Keycloak API calls
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			m.logger.Warn("Authorization header missing for Keycloak permission check")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Missing authorization token",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			m.logger.Warn("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Check permission with Keycloak Authorization Services
		ctx := context.Background()

		m.logger.Info("Checking permission with Keycloak Authorization Services",
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("scope", scope))

		// This would call Keycloak's Authorization Services API
		// For now, we'll implement a basic check - in production this would call:
		// POST /realms/{realm}/authz/protection/permission
		allowed, err := m.checkKeycloakPermission(ctx, token, resource, scope)
		if err != nil {
			m.logger.Error("Failed to check permission with Keycloak",
				logging.Error(err),
				logging.String("user_id", userID),
				logging.String("resource", resource),
				logging.String("scope", scope))

			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Authorization Check Failed",
				"message": "Failed to verify permissions with Keycloak",
			})
			c.Abort()
			return
		}

		if !allowed {
			m.logger.Warn("Keycloak permission denied",
				logging.String("user_id", userID),
				logging.String("resource", resource),
				logging.String("scope", scope))

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "You do not have permission to perform this action",
				"details": gin.H{
					"required_permission": resource + ":" + scope,
				},
			})
			c.Abort()
			return
		}

		m.logger.Info("Keycloak permission granted",
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("scope", scope))

		// Store permission info in context
		c.Set("permission_allowed", allowed)
		c.Set("resource", resource)
		c.Set("scope", scope)

		c.Next()
	}
}

// checkKeycloakPermission implements the actual Keycloak Authorization Services call
// This is a placeholder - implement based on your Keycloak Authorization Services setup
func (m *KeycloakAuthorizationMiddleware) checkKeycloakPermission(ctx context.Context, accessToken, resource, scope string) (bool, error) {
	// TODO: Implement actual Keycloak Authorization Services API call
	// For now, return true for demo purposes
	// In production, this would make an HTTP call to:
	// POST /realms/{realm}/authz/protection/permission
	// with the RPT token and permission request

	m.logger.Info("Keycloak permission check - placeholder implementation",
		logging.String("resource", resource),
		logging.String("scope", scope))

	// Placeholder: always allow for now
	// Replace with actual Keycloak Authorization Services integration
	return true, nil
}

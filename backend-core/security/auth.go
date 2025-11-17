package security

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// AuthManager handles authentication operations
type AuthManager struct {
	jwtManager *JWTManager
	bcryptCost int
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(jwtManager *JWTManager, bcryptCost int) *AuthManager {
	return &AuthManager{
		jwtManager: jwtManager,
		bcryptCost: bcryptCost,
	}
}

// HashPassword hashes a password using bcrypt
func (a *AuthManager) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), a.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func (a *AuthManager) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateToken generates a JWT token for a user
func (a *AuthManager) GenerateToken(userID, username, role string) (string, error) {
	return a.jwtManager.GenerateToken(userID, username, role)
}

// ValidateToken validates a JWT token
func (a *AuthManager) ValidateToken(tokenString string) (*Claims, error) {
	return a.jwtManager.ValidateToken(tokenString)
}

// ExtractTokenFromHeader extracts token from Authorization header
func (a *AuthManager) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// GetUserFromContext extracts user information from context
func (a *AuthManager) GetUserFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value("user").(*Claims)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	return claims, nil
}

// SetUserInContext sets user information in context
func (a *AuthManager) SetUserInContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, "user", claims)
}

// HasRole checks if the user has a specific role
func (a *AuthManager) HasRole(claims *Claims, role string) bool {
	return claims.Role == role
}

// HasAnyRole checks if the user has any of the specified roles
func (a *AuthManager) HasAnyRole(claims *Claims, roles []string) bool {
	for _, role := range roles {
		if claims.Role == role {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (a *AuthManager) IsAdmin(claims *Claims) bool {
	return claims.Role == "admin"
}

// IsUser checks if the user is a regular user
func (a *AuthManager) IsUser(claims *Claims) bool {
	return claims.Role == "user"
}

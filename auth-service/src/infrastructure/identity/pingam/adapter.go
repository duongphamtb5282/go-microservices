package pingam

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/infrastructure/identity/models"
	"backend-core/cache"
	"backend-core/logging"
)

// PingAMAdapter provides a service-specific interface to PingAM
type PingAMAdapter struct {
	client *PingAMClient
	cache  cache.Cache
	logger *logging.Logger
	config PingAMConfig
}

// NewPingAMAdapter creates a new PingAM adapter
func NewPingAMAdapter(client *PingAMClient, cache cache.Cache, config PingAMConfig, logger *logging.Logger) *PingAMAdapter {
	return &PingAMAdapter{
		client: client,
		cache:  cache,
		logger: logger,
		config: config,
	}
}

// Authenticate authenticates a user with PingAM
func (a *PingAMAdapter) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("auth:%s", credentials.Username)
	var cached models.AuthResult
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("Authentication result found in cache",
			logging.String("username", credentials.Username))
		return &cached, nil
	}

	// Authenticate with PingAM
	result, err := a.client.Authenticate(ctx, credentials)
	if err != nil {
		a.logger.Error("PingAM authentication failed",
			logging.Error(err),
			logging.String("username", credentials.Username))
		return nil, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, result, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache authentication result",
			logging.Error(err),
			logging.String("username", credentials.Username))
	}

	a.logger.Info("User authenticated successfully",
		logging.String("user_id", result.UserID),
		logging.String("username", result.Username))

	return result, nil
}

// ValidateToken validates a token with PingAM
func (a *PingAMAdapter) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("token:%s", token)
	var cached models.TokenInfo
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("Token validation result found in cache")
		return &cached, nil
	}

	// Validate with PingAM
	tokenInfo, err := a.client.ValidateToken(ctx, token)
	if err != nil {
		a.logger.Error("Token validation failed",
			logging.Error(err))
		return nil, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, tokenInfo, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache token validation result",
			logging.Error(err))
	}

	a.logger.Debug("Token validated successfully",
		logging.Bool("active", tokenInfo.Active),
		logging.String("user_id", tokenInfo.UserID))

	return tokenInfo, nil
}

// GetUserProfile retrieves user profile from PingAM
func (a *PingAMAdapter) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("profile:%s", userID)
	var cached models.UserProfile
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User profile found in cache",
			logging.String("user_id", userID))
		return &cached, nil
	}

	// Get from PingAM
	profile, err := a.client.GetUserProfile(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user profile from PingAM",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, profile, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache user profile",
			logging.Error(err),
			logging.String("user_id", userID))
	}

	a.logger.Debug("User profile retrieved successfully",
		logging.String("user_id", profile.ID),
		logging.String("username", profile.Username))

	return profile, nil
}

// CheckPermission checks user permission with PingAM
func (a *PingAMAdapter) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("permission:%s:%s:%s", userID, resource, action)
	var cached bool
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("Permission check result found in cache",
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
		return cached, nil
	}

	// Check with PingAM
	result, err := a.client.CheckPermissions(ctx, userID, resource, action)
	if err != nil {
		a.logger.Error("Permission check failed",
			logging.Error(err),
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
		return false, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, result.Allowed, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache permission result",
			logging.Error(err),
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
	}

	a.logger.Debug("Permission checked successfully",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("allowed", result.Allowed))

	return result.Allowed, nil
}

// GetUserRoles retrieves user roles from PingAM
func (a *PingAMAdapter) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("roles:%s", userID)
	var cached []string
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User roles found in cache",
			logging.String("user_id", userID))
		return cached, nil
	}

	// Get from PingAM
	roles, err := a.client.GetUserRoles(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user roles from PingAM",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, roles, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache user roles",
			logging.Error(err),
			logging.String("user_id", userID))
	}

	a.logger.Debug("User roles retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions retrieves user permissions from PingAM
func (a *PingAMAdapter) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("permissions:%s", userID)
	var cached []string
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User permissions found in cache",
			logging.String("user_id", userID))
		return cached, nil
	}

	// Get from PingAM
	permissions, err := a.client.GetUserPermissions(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user permissions from PingAM",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	// Cache the result
	if err := a.setCache(ctx, cacheKey, permissions, a.config.CacheTTL); err != nil {
		a.logger.Warn("Failed to cache user permissions",
			logging.Error(err),
			logging.String("user_id", userID))
	}

	a.logger.Debug("User permissions retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	return permissions, nil
}

// RefreshToken refreshes an access token
func (a *PingAMAdapter) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	result, err := a.client.RefreshToken(ctx, refreshToken)
	if err != nil {
		a.logger.Error("Token refresh failed",
			logging.Error(err))
		return nil, err
	}

	a.logger.Info("Token refreshed successfully")

	return result, nil
}

// RevokeToken revokes an access token
func (a *PingAMAdapter) RevokeToken(ctx context.Context, token string) error {
	err := a.client.RevokeToken(ctx, token)
	if err != nil {
		a.logger.Error("Token revocation failed",
			logging.Error(err))
		return err
	}

	// Clear any cached data for this token
	a.clearTokenCache(ctx, token)

	a.logger.Info("Token revoked successfully")

	return nil
}

// HealthCheck performs a health check against PingAM
func (a *PingAMAdapter) HealthCheck(ctx context.Context) error {
	err := a.client.HealthCheck(ctx)
	if err != nil {
		a.logger.Error("PingAM health check failed",
			logging.Error(err))
		return err
	}

	a.logger.Debug("PingAM health check successful")

	return nil
}

// Cache helper methods

func (a *PingAMAdapter) getFromCache(ctx context.Context, key string, dest interface{}) error {
	if a.cache == nil {
		return fmt.Errorf("cache not available")
	}

	return a.cache.Get(ctx, key, dest)
}

func (a *PingAMAdapter) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if a.cache == nil {
		return fmt.Errorf("cache not available")
	}

	return a.cache.Set(ctx, key, value, ttl)
}

func (a *PingAMAdapter) clearTokenCache(ctx context.Context, token string) {
	if a.cache == nil {
		return
	}

	// Clear token-related cache entries
	patterns := []string{
		"token:" + token,
		"auth:*",
		"profile:*",
		"roles:*",
		"permissions:*",
	}

	for _, pattern := range patterns {
		if err := a.cache.DeletePattern(ctx, pattern); err != nil {
			a.logger.Warn("Failed to clear cache pattern",
				logging.Error(err),
				logging.String("pattern", pattern))
		}
	}
}

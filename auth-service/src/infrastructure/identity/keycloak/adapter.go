package keycloak

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/infrastructure/identity/models"
	"backend-core/cache"
	"backend-core/logging"
)

// KeycloakAdapter provides a service-specific interface to Keycloak
type KeycloakAdapter struct {
	client *KeycloakClient
	cache  cache.Cache
	logger *logging.Logger
	config KeycloakConfig
}

// NewKeycloakAdapter creates a new Keycloak adapter
func NewKeycloakAdapter(client *KeycloakClient, cache cache.Cache, config KeycloakConfig, logger *logging.Logger) *KeycloakAdapter {
	return &KeycloakAdapter{
		client: client,
		cache:  cache,
		logger: logger,
		config: config,
	}
}

// Authenticate authenticates a user with Keycloak
func (a *KeycloakAdapter) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("auth:%s", credentials.Username)
	var cached models.AuthResult
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("Authentication result found in cache",
			logging.String("username", credentials.Username))
		return &cached, nil
	}

	// Authenticate with Keycloak
	result, err := a.client.Authenticate(ctx, credentials)
	if err != nil {
		a.logger.Error("Keycloak authentication failed",
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

// ValidateToken validates a token with Keycloak
func (a *KeycloakAdapter) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("token:%s", token)
	var cached models.TokenInfo
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("Token validation result found in cache")
		return &cached, nil
	}

	// Validate with Keycloak
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

// GetUserProfile retrieves user profile from Keycloak
func (a *KeycloakAdapter) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("profile:%s", userID)
	var cached models.UserProfile
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User profile found in cache",
			logging.String("user_id", userID))
		return &cached, nil
	}

	// Get from Keycloak
	profile, err := a.client.GetUserProfile(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user profile from Keycloak",
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

// CheckPermission checks user permission with Keycloak
func (a *KeycloakAdapter) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
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

	// Check with Keycloak
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

// GetUserRoles retrieves user roles from Keycloak
func (a *KeycloakAdapter) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("roles:%s", userID)
	var cached []string
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User roles found in cache",
			logging.String("user_id", userID))
		return cached, nil
	}

	// Get from Keycloak
	roles, err := a.client.GetUserRoles(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user roles from Keycloak",
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

// GetUserPermissions retrieves user permissions from Keycloak
func (a *KeycloakAdapter) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("permissions:%s", userID)
	var cached []string
	if err := a.getFromCache(ctx, cacheKey, &cached); err == nil {
		a.logger.Debug("User permissions found in cache",
			logging.String("user_id", userID))
		return cached, nil
	}

	// Get from Keycloak
	permissions, err := a.client.GetUserPermissions(ctx, userID)
	if err != nil {
		a.logger.Error("Failed to get user permissions from Keycloak",
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
func (a *KeycloakAdapter) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
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
func (a *KeycloakAdapter) RevokeToken(ctx context.Context, token string) error {
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

// HealthCheck performs a health check against Keycloak
func (a *KeycloakAdapter) HealthCheck(ctx context.Context) error {
	err := a.client.HealthCheck(ctx)
	if err != nil {
		a.logger.Error("Keycloak health check failed",
			logging.Error(err))
		return err
	}

	a.logger.Debug("Keycloak health check successful")

	return nil
}

// InitiateSSOLogin initiates an SSO login flow
func (a *KeycloakAdapter) InitiateSSOLogin(ctx context.Context, provider string) (*models.AuthURL, error) {
	if !a.config.EnableSSO {
		return nil, fmt.Errorf("SSO is not enabled")
	}

	// Build authorization URL with Keycloak identity provider
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		a.config.OAuth.AuthorizationURL,
		a.config.ClientID,
		a.config.RedirectURI,
		"openid profile email")

	if provider != "" {
		authURL += fmt.Sprintf("&kc_idp_hint=%s", provider)
	}

	a.logger.Info("SSO login initiated",
		logging.String("provider", provider))

	return &models.AuthURL{
		URL:   authURL,
		State: generateState(),
	}, nil
}

// InitiateMFA initiates MFA challenge
func (a *KeycloakAdapter) InitiateMFA(ctx context.Context, userID string) (*models.MFAChallenge, error) {
	if !a.config.EnableMFA {
		return nil, fmt.Errorf("MFA is not enabled")
	}

	// In Keycloak, MFA is typically handled during authentication flow
	// This is a simplified implementation
	challenge := &models.MFAChallenge{
		ChallengeID: generateChallengeID(),
		Methods:     []string{"totp", "sms", "email"},
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	// Store challenge in cache
	cacheKey := fmt.Sprintf("mfa:%s", challenge.ChallengeID)
	if err := a.setCache(ctx, cacheKey, challenge, 5*time.Minute); err != nil {
		a.logger.Warn("Failed to cache MFA challenge", logging.Error(err))
	}

	a.logger.Info("MFA challenge initiated",
		logging.String("user_id", userID),
		logging.String("challenge_id", challenge.ChallengeID))

	return challenge, nil
}

// VerifyMFA verifies MFA challenge response
func (a *KeycloakAdapter) VerifyMFA(ctx context.Context, challengeID, code string) (*models.MFAVerification, error) {
	if !a.config.EnableMFA {
		return nil, fmt.Errorf("MFA is not enabled")
	}

	// Retrieve challenge from cache
	cacheKey := fmt.Sprintf("mfa:%s", challengeID)
	var challenge models.MFAChallenge
	if err := a.getFromCache(ctx, cacheKey, &challenge); err != nil {
		a.logger.Error("MFA challenge not found", logging.Error(err))
		return nil, ErrMFAFailed
	}

	// Verify the code (simplified - in production, verify with Keycloak)
	// For now, accept any 6-digit code
	if len(code) != 6 {
		return nil, ErrMFAFailed
	}

	verification := &models.MFAVerification{
		Success:    true,
		Method:     "totp",
		VerifiedAt: time.Now(),
	}

	// Clear challenge from cache
	if a.cache != nil {
		_ = a.cache.Delete(ctx, cacheKey)
	}

	a.logger.Info("MFA verification successful",
		logging.String("challenge_id", challengeID))

	return verification, nil
}

// Cache helper methods

func (a *KeycloakAdapter) getFromCache(ctx context.Context, key string, dest interface{}) error {
	if a.cache == nil {
		return fmt.Errorf("cache not available")
	}

	return a.cache.Get(ctx, key, dest)
}

func (a *KeycloakAdapter) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if a.cache == nil {
		return fmt.Errorf("cache not available")
	}

	return a.cache.Set(ctx, key, value, ttl)
}

func (a *KeycloakAdapter) clearTokenCache(ctx context.Context, token string) {
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

// Helper functions

func generateState() string {
	// Generate a random state for OAuth2 flow
	// In production, use crypto/rand
	return fmt.Sprintf("state_%d", time.Now().UnixNano())
}

func generateChallengeID() string {
	// Generate a unique challenge ID
	// In production, use UUID or crypto/rand
	return fmt.Sprintf("challenge_%d", time.Now().UnixNano())
}

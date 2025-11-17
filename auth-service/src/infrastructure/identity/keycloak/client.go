package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"auth-service/src/infrastructure/identity/models"
	"backend-core/logging"
)

// KeycloakClient handles HTTP communication with Keycloak
type KeycloakClient struct {
	baseURL      string
	realm        string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	logger       *logging.Logger
	accessToken  string
	tokenExpiry  time.Time
	config       KeycloakConfig
}

// NewKeycloakClient creates a new Keycloak HTTP client
func NewKeycloakClient(config KeycloakConfig, logger *logging.Logger) (*KeycloakClient, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	config.BuildURLs()

	return &KeycloakClient{
		baseURL:      config.BaseURL,
		realm:        config.Realm,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		config: config,
	}, nil
}

// Authenticate authenticates a user with Keycloak using Resource Owner Password Credentials flow
func (c *KeycloakClient) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	tokenURL := c.config.OAuth.TokenURL

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("username", credentials.Username)
	data.Set("password", credentials.Password)
	data.Set("scope", strings.Join(c.config.Scopes, " "))

	resp, err := c.postForm(ctx, tokenURL, data)
	if err != nil {
		c.logger.Error("Keycloak authentication failed",
			logging.Error(err),
			logging.String("username", credentials.Username))
		return nil, fmt.Errorf("%w: %v", ErrAuthenticationFailed, err)
	}

	var tokenResp struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		Scope            string `json:"scope"`
	}

	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse auth result: %w", err)
	}

	// Get user info to populate UserID and Username
	userInfo, err := c.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		c.logger.Warn("Failed to get user info after authentication", logging.Error(err))
	}

	authResult := &models.AuthResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
		Scope:        tokenResp.Scope,
		IssuedAt:     time.Now(),
	}

	if userInfo != nil {
		authResult.UserID = userInfo.Sub
		authResult.Username = userInfo.PreferredUsername
	}

	// Store access token for future requests
	c.accessToken = authResult.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(authResult.ExpiresIn) * time.Second)

	c.logger.Info("User authenticated successfully with Keycloak",
		logging.String("user_id", authResult.UserID),
		logging.String("username", authResult.Username))

	return authResult, nil
}

// ValidateToken validates a token with Keycloak using token introspection
func (c *KeycloakClient) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	introspectURL := c.baseURL + "/realms/" + c.realm + "/protocol/openid-connect/token/introspect"

	data := url.Values{}
	data.Set("token", token)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	resp, err := c.postForm(ctx, introspectURL, data)
	if err != nil {
		c.logger.Error("Token validation failed", logging.Error(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	var introspectResp struct {
		Active   bool   `json:"active"`
		Scope    string `json:"scope"`
		ClientID string `json:"client_id"`
		Username string `json:"username"`
		Sub      string `json:"sub"`
		Exp      int64  `json:"exp"`
		Iat      int64  `json:"iat"`
	}

	if err := json.Unmarshal(resp, &introspectResp); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	tokenInfo := &models.TokenInfo{
		Active:    introspectResp.Active,
		Scope:     introspectResp.Scope,
		ClientID:  introspectResp.ClientID,
		Username:  introspectResp.Username,
		UserID:    introspectResp.Sub,
		ExpiresAt: time.Unix(introspectResp.Exp, 0),
		IssuedAt:  time.Unix(introspectResp.Iat, 0),
	}

	c.logger.Debug("Token validated with Keycloak",
		logging.Bool("active", tokenInfo.Active),
		logging.String("user_id", tokenInfo.UserID))

	return tokenInfo, nil
}

// GetUserProfile retrieves user profile from Keycloak
func (c *KeycloakClient) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	// Get user info using access token
	userInfo, err := c.getUserInfo(ctx, c.accessToken)
	if err != nil {
		c.logger.Error("Failed to get user profile from Keycloak",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	profile := &models.UserProfile{
		ID:         userInfo.Sub,
		Username:   userInfo.PreferredUsername,
		Email:      userInfo.Email,
		FirstName:  userInfo.GivenName,
		LastName:   userInfo.FamilyName,
		Roles:      userInfo.RealmRoles,
		Groups:     userInfo.Groups,
		Attributes: make(map[string]interface{}),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	c.logger.Debug("User profile retrieved successfully",
		logging.String("user_id", profile.ID),
		logging.String("username", profile.Username))

	return profile, nil
}

// CheckPermissions checks user permissions with Keycloak
func (c *KeycloakClient) CheckPermissions(ctx context.Context, userID, resource, action string) (*models.PermissionResult, error) {
	// For Keycloak, we need to use the token to check permissions
	// This is a simplified implementation - in production, you'd use Keycloak's Authorization Services

	c.logger.Debug("Checking permissions with Keycloak",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	// Get user roles
	roles, err := c.GetUserRoles(ctx, userID)
	if err != nil {
		c.logger.Error("Failed to get user roles for permission check",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	// Simple permission check based on roles
	// In production, integrate with Keycloak Authorization Services
	result := &models.PermissionResult{
		Allowed: len(roles) > 0, // Simplified: allow if user has any roles
		Reason:  "Permission check based on roles",
		Roles:   roles,
	}

	c.logger.Debug("Permission checked with Keycloak",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("allowed", result.Allowed))

	return result, nil
}

// GetUserRoles retrieves user roles from Keycloak
func (c *KeycloakClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	userInfo, err := c.getUserInfo(ctx, c.accessToken)
	if err != nil {
		c.logger.Error("Failed to get user roles from Keycloak",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	roles := userInfo.RealmRoles
	if userInfo.ResourceAccess != nil {
		if clientRoles, ok := userInfo.ResourceAccess[c.clientID]; ok {
			if clientRoleList, ok := clientRoles["roles"].([]interface{}); ok {
				for _, role := range clientRoleList {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	c.logger.Debug("User roles retrieved from Keycloak",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions retrieves user permissions from Keycloak
func (c *KeycloakClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	// In Keycloak, permissions are typically derived from roles
	// For fine-grained permissions, use Keycloak Authorization Services
	roles, err := c.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	// For now, return roles as permissions
	// In production, map roles to permissions or use Keycloak's UMA
	permissions := make([]string, len(roles))
	copy(permissions, roles)

	c.logger.Debug("User permissions retrieved from Keycloak",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	return permissions, nil
}

// RefreshToken refreshes an access token
func (c *KeycloakClient) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	tokenURL := c.config.OAuth.TokenURL

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", refreshToken)

	resp, err := c.postForm(ctx, tokenURL, data)
	if err != nil {
		c.logger.Error("Token refresh failed", logging.Error(err))
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	var tokenResp models.TokenResponse
	if err := json.Unmarshal(resp, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update stored access token
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	c.logger.Info("Token refreshed successfully with Keycloak")

	return &tokenResp, nil
}

// RevokeToken revokes an access token
func (c *KeycloakClient) RevokeToken(ctx context.Context, token string) error {
	logoutURL := c.config.OAuth.LogoutURL

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", token)

	_, err := c.postForm(ctx, logoutURL, data)
	if err != nil {
		c.logger.Error("Token revocation failed", logging.Error(err))
		return fmt.Errorf("token revocation failed: %w", err)
	}

	c.logger.Info("Token revoked successfully with Keycloak")

	return nil
}

// HealthCheck performs a health check against Keycloak
func (c *KeycloakClient) HealthCheck(ctx context.Context) error {
	healthURL := c.baseURL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Keycloak health check failed", logging.Error(err))
		return ErrKeycloakUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("keycloak health check returned status %d", resp.StatusCode)
	}

	c.logger.Debug("Keycloak health check successful")

	return nil
}

// Helper methods

// getUserInfo retrieves user information from Keycloak userinfo endpoint
func (c *KeycloakClient) getUserInfo(ctx context.Context, accessToken string) (*KeycloakUserInfo, error) {
	userInfoURL := c.config.OAuth.UserInfoURL

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo KeycloakUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

func (c *KeycloakClient) postForm(ctx context.Context, endpoint string, data url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return body, nil
}

func (c *KeycloakClient) post(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Add authorization header if we have a token
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return body, nil
}

func (c *KeycloakClient) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add authorization header if we have a token
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return body, nil
}

// KeycloakUserInfo represents user information from Keycloak
type KeycloakUserInfo struct {
	Sub               string                            `json:"sub"`
	PreferredUsername string                            `json:"preferred_username"`
	Name              string                            `json:"name"`
	GivenName         string                            `json:"given_name"`
	FamilyName        string                            `json:"family_name"`
	Email             string                            `json:"email"`
	EmailVerified     bool                              `json:"email_verified"`
	RealmRoles        []string                          `json:"realm_access.roles"`
	Groups            []string                          `json:"groups"`
	ResourceAccess    map[string]map[string]interface{} `json:"resource_access"`
}

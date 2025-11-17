package pingam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"auth-service/src/infrastructure/identity/models"
	"backend-core/logging"
)

// PingAMClient handles HTTP communication with PingAM
type PingAMClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	logger       *logging.Logger
	accessToken  string
	tokenExpiry  time.Time
}

// NewPingAMClient creates a new PingAM HTTP client
func NewPingAMClient(config PingAMConfig, logger *logging.Logger) *PingAMClient {
	return &PingAMClient{
		baseURL:      config.BaseURL,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// Authenticate authenticates a user with PingAM
func (c *PingAMClient) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	endpoint := fmt.Sprintf("%s/oauth/token", c.baseURL)

	data := map[string]string{
		"grant_type":    "password",
		"username":      credentials.Username,
		"password":      credentials.Password,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	resp, err := c.post(ctx, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var authResult models.AuthResult
	if err := json.Unmarshal(resp, &authResult); err != nil {
		return nil, fmt.Errorf("failed to parse auth result: %w", err)
	}

	// Store access token for future requests
	c.accessToken = authResult.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(authResult.ExpiresIn) * time.Second)

	c.logger.Info("User authenticated successfully with PingAM",
		logging.String("user_id", authResult.UserID),
		logging.String("username", authResult.Username))

	return &authResult, nil
}

// ValidateToken validates a token with PingAM
func (c *PingAMClient) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	endpoint := fmt.Sprintf("%s/oauth/introspect", c.baseURL)

	data := map[string]string{
		"token":         token,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	resp, err := c.post(ctx, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	var tokenInfo models.TokenInfo
	if err := json.Unmarshal(resp, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	c.logger.Debug("Token validated with PingAM",
		logging.Bool("active", tokenInfo.Active),
		logging.String("user_id", tokenInfo.UserID))

	return &tokenInfo, nil
}

// GetUserProfile retrieves user profile from PingAM
func (c *PingAMClient) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	endpoint := fmt.Sprintf("%s/api/users/%s", c.baseURL, userID)

	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	var profile models.UserProfile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse user profile: %w", err)
	}

	c.logger.Debug("User profile retrieved from PingAM",
		logging.String("user_id", profile.ID),
		logging.String("username", profile.Username))

	return &profile, nil
}

// CheckPermissions checks user permissions with PingAM
func (c *PingAMClient) CheckPermissions(ctx context.Context, userID, resource, action string) (*models.PermissionResult, error) {
	endpoint := fmt.Sprintf("%s/api/authorize", c.baseURL)

	c.logger.Debug("Checking permissions with PingAM",
		logging.String("endpoint", endpoint),
		logging.String("baseURL", c.baseURL),
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	data := map[string]interface{}{
		"user_id":  userID,
		"resource": resource,
		"action":   action,
	}

	resp, err := c.post(ctx, endpoint, data)
	if err != nil {
		c.logger.Error("POST request to PingAM failed",
			logging.Error(err),
			logging.String("endpoint", endpoint))
		return nil, fmt.Errorf("permission check failed: %w", err)
	}

	var result models.PermissionResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse permission result: %w", err)
	}

	c.logger.Debug("Permission checked with PingAM",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("allowed", result.Allowed))

	return &result, nil
}

// GetUserRoles retrieves user roles from PingAM
func (c *PingAMClient) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/api/users/%s/roles", c.baseURL, userID)

	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	var roles []string
	if err := json.Unmarshal(resp, &roles); err != nil {
		return nil, fmt.Errorf("failed to parse user roles: %w", err)
	}

	c.logger.Debug("User roles retrieved from PingAM",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions retrieves user permissions from PingAM
func (c *PingAMClient) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/api/users/%s/permissions", c.baseURL, userID)

	resp, err := c.get(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	var permissions []string
	if err := json.Unmarshal(resp, &permissions); err != nil {
		return nil, fmt.Errorf("failed to parse user permissions: %w", err)
	}

	c.logger.Debug("User permissions retrieved from PingAM",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	return permissions, nil
}

// RefreshToken refreshes an access token
func (c *PingAMClient) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	endpoint := fmt.Sprintf("%s/oauth/token", c.baseURL)

	data := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	resp, err := c.post(ctx, endpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	var tokenResponse models.TokenResponse
	if err := json.Unmarshal(resp, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update stored access token
	c.accessToken = tokenResponse.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	c.logger.Info("Token refreshed successfully with PingAM")

	return &tokenResponse, nil
}

// RevokeToken revokes an access token
func (c *PingAMClient) RevokeToken(ctx context.Context, token string) error {
	endpoint := fmt.Sprintf("%s/oauth/revoke", c.baseURL)

	data := map[string]string{
		"token":         token,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}

	_, err := c.post(ctx, endpoint, data)
	if err != nil {
		return fmt.Errorf("token revocation failed: %w", err)
	}

	c.logger.Info("Token revoked successfully with PingAM")

	return nil
}

// HealthCheck performs a health check against PingAM
func (c *PingAMClient) HealthCheck(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s/health", c.baseURL)

	_, err := c.get(ctx, endpoint)
	if err != nil {
		return fmt.Errorf("PingAM health check failed: %w", err)
	}

	c.logger.Debug("PingAM health check successful")

	return nil
}

// Helper methods

func (c *PingAMClient) get(ctx context.Context, endpoint string) ([]byte, error) {
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *PingAMClient) post(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return io.ReadAll(resp.Body)
}

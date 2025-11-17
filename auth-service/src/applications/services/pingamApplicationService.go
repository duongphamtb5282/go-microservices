package services

import (
	"context"
	"fmt"
	"time"

	"auth-service/src/infrastructure/identity/models"
	"auth-service/src/infrastructure/identity/pingam"
	"backend-core/logging"
)

// PingAMApplicationService handles PingAM-related application logic
type PingAMApplicationService struct {
	pingamAdapter *pingam.PingAMAdapter
	logger        *logging.Logger
}

// NewPingAMApplicationService creates a new PingAM application service
func NewPingAMApplicationService(
	pingamAdapter *pingam.PingAMAdapter,
	logger *logging.Logger,
) *PingAMApplicationService {
	return &PingAMApplicationService{
		pingamAdapter: pingamAdapter,
		logger:        logger,
	}
}

// LoginWithPingAM authenticates user with PingAM
func (s *PingAMApplicationService) LoginWithPingAM(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	s.logger.Info("Starting PingAM authentication",
		logging.String("username", credentials.Username))

	// Authenticate with PingAM
	authResult, err := s.pingamAdapter.Authenticate(ctx, credentials)
	if err != nil {
		s.logger.Error("PingAM authentication failed",
			logging.Error(err),
			logging.String("username", credentials.Username))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	s.logger.Info("User authenticated successfully with PingAM",
		logging.String("user_id", authResult.UserID),
		logging.String("username", authResult.Username))

	return authResult, nil
}

// ValidateSession validates session with PingAM
func (s *PingAMApplicationService) ValidateSession(ctx context.Context, sessionID string) (*models.SessionInfo, error) {
	// For now, we'll use the sessionID as the token and validate it directly
	tokenInfo, err := s.pingamAdapter.ValidateToken(ctx, sessionID)
	if err != nil {
		s.logger.Error("Token validation failed",
			logging.Error(err),
			logging.String("session_id", sessionID))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !tokenInfo.Active {
		s.logger.Warn("Session token is not active",
			logging.String("session_id", sessionID))
		return nil, fmt.Errorf("session token is not active")
	}

	// Create a simple session info
	session := &models.SessionInfo{
		ID:           sessionID,
		UserID:       tokenInfo.UserID,
		SessionToken: sessionID,
		CreatedAt:    tokenInfo.IssuedAt,
		ExpiresAt:    tokenInfo.ExpiresAt,
		LastAccess:   time.Now(),
		IsActive:     tokenInfo.Active,
	}

	s.logger.Debug("Session validated successfully",
		logging.String("session_id", sessionID),
		logging.String("user_id", session.UserID))

	return session, nil
}

// CheckPermission checks user permission with PingAM
func (s *PingAMApplicationService) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	s.logger.Debug("Checking permission with PingAM",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	// Check permission with PingAM
	allowed, err := s.pingamAdapter.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		s.logger.Error("Permission check failed",
			logging.Error(err),
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	s.logger.Debug("Permission check completed",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("allowed", allowed))

	return allowed, nil
}

// GetUserRoles retrieves user roles from PingAM
func (s *PingAMApplicationService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	s.logger.Debug("Getting user roles from PingAM",
		logging.String("user_id", userID))

	roles, err := s.pingamAdapter.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles from PingAM",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	s.logger.Debug("User roles retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions retrieves user permissions from PingAM
func (s *PingAMApplicationService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	s.logger.Debug("Getting user permissions from PingAM",
		logging.String("user_id", userID))

	permissions, err := s.pingamAdapter.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user permissions from PingAM",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	s.logger.Debug("User permissions retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	return permissions, nil
}

// RefreshToken refreshes an access token
func (s *PingAMApplicationService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	s.logger.Info("Refreshing token with PingAM")

	result, err := s.pingamAdapter.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed",
			logging.Error(err))
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	s.logger.Info("Token refreshed successfully")

	return result, nil
}

// RevokeToken revokes an access token
func (s *PingAMApplicationService) RevokeToken(ctx context.Context, token string) error {
	s.logger.Info("Revoking token with PingAM")

	err := s.pingamAdapter.RevokeToken(ctx, token)
	if err != nil {
		s.logger.Error("Token revocation failed",
			logging.Error(err))
		return fmt.Errorf("token revocation failed: %w", err)
	}

	// Note: Token cache clearing would be implemented here if needed

	s.logger.Info("Token revoked successfully")

	return nil
}

// HealthCheck performs a health check against PingAM
func (s *PingAMApplicationService) HealthCheck(ctx context.Context) error {
	s.logger.Debug("Performing PingAM health check")

	err := s.pingamAdapter.HealthCheck(ctx)
	if err != nil {
		s.logger.Error("PingAM health check failed",
			logging.Error(err))
		return fmt.Errorf("PingAM health check failed: %w", err)
	}

	s.logger.Debug("PingAM health check successful")

	return nil
}

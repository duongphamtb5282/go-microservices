package services

import (
	"context"

	"auth-service/src/infrastructure/identity/keycloak"
	"auth-service/src/infrastructure/identity/models"
	"backend-core/logging"
)

// KeycloakApplicationService handles Keycloak application logic
type KeycloakApplicationService struct {
	adapter *keycloak.KeycloakAdapter
	logger  *logging.Logger
}

// NewKeycloakApplicationService creates a new Keycloak application service
func NewKeycloakApplicationService(adapter *keycloak.KeycloakAdapter, logger *logging.Logger) *KeycloakApplicationService {
	return &KeycloakApplicationService{
		adapter: adapter,
		logger:  logger,
	}
}

// Authenticate authenticates a user with Keycloak
func (s *KeycloakApplicationService) Authenticate(ctx context.Context, credentials models.Credentials) (*models.AuthResult, error) {
	s.logger.Info("Authenticating user with Keycloak",
		logging.String("username", credentials.Username))

	result, err := s.adapter.Authenticate(ctx, credentials)
	if err != nil {
		s.logger.Error("Authentication failed",
			logging.Error(err),
			logging.String("username", credentials.Username))
		return nil, err
	}

	s.logger.Info("User authenticated successfully",
		logging.String("user_id", result.UserID),
		logging.String("username", result.Username))

	return result, nil
}

// ValidateToken validates a token with Keycloak
func (s *KeycloakApplicationService) ValidateToken(ctx context.Context, token string) (*models.TokenInfo, error) {
	s.logger.Debug("Validating token with Keycloak")

	tokenInfo, err := s.adapter.ValidateToken(ctx, token)
	if err != nil {
		s.logger.Error("Token validation failed", logging.Error(err))
		return nil, err
	}

	s.logger.Debug("Token validated successfully",
		logging.Bool("active", tokenInfo.Active),
		logging.String("user_id", tokenInfo.UserID))

	return tokenInfo, nil
}

// GetUserProfile retrieves user profile from Keycloak
func (s *KeycloakApplicationService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	s.logger.Debug("Getting user profile from Keycloak",
		logging.String("user_id", userID))

	profile, err := s.adapter.GetUserProfile(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user profile",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	s.logger.Debug("User profile retrieved successfully",
		logging.String("user_id", profile.ID),
		logging.String("username", profile.Username))

	return profile, nil
}

// CheckPermission checks user permission with Keycloak
func (s *KeycloakApplicationService) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	s.logger.Debug("Checking permission with Keycloak",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action))

	allowed, err := s.adapter.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		s.logger.Error("Permission check failed",
			logging.Error(err),
			logging.String("user_id", userID),
			logging.String("resource", resource),
			logging.String("action", action))
		return false, err
	}

	s.logger.Debug("Permission checked successfully",
		logging.String("user_id", userID),
		logging.String("resource", resource),
		logging.String("action", action),
		logging.Bool("allowed", allowed))

	return allowed, nil
}

// GetUserRoles retrieves user roles from Keycloak
func (s *KeycloakApplicationService) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	s.logger.Debug("Getting user roles from Keycloak",
		logging.String("user_id", userID))

	roles, err := s.adapter.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user roles",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	s.logger.Debug("User roles retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("role_count", len(roles)))

	return roles, nil
}

// GetUserPermissions retrieves user permissions from Keycloak
func (s *KeycloakApplicationService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	s.logger.Debug("Getting user permissions from Keycloak",
		logging.String("user_id", userID))

	permissions, err := s.adapter.GetUserPermissions(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user permissions",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	s.logger.Debug("User permissions retrieved successfully",
		logging.String("user_id", userID),
		logging.Int("permission_count", len(permissions)))

	return permissions, nil
}

// RefreshToken refreshes an access token
func (s *KeycloakApplicationService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	s.logger.Debug("Refreshing token with Keycloak")

	result, err := s.adapter.RefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("Token refresh failed", logging.Error(err))
		return nil, err
	}

	s.logger.Info("Token refreshed successfully")

	return result, nil
}

// RevokeToken revokes an access token
func (s *KeycloakApplicationService) RevokeToken(ctx context.Context, token string) error {
	s.logger.Debug("Revoking token with Keycloak")

	err := s.adapter.RevokeToken(ctx, token)
	if err != nil {
		s.logger.Error("Token revocation failed", logging.Error(err))
		return err
	}

	s.logger.Info("Token revoked successfully")

	return nil
}

// InitiateSSOLogin initiates an SSO login flow
func (s *KeycloakApplicationService) InitiateSSOLogin(ctx context.Context, provider string) (*models.AuthURL, error) {
	s.logger.Info("Initiating SSO login",
		logging.String("provider", provider))

	authURL, err := s.adapter.InitiateSSOLogin(ctx, provider)
	if err != nil {
		s.logger.Error("SSO initiation failed",
			logging.Error(err),
			logging.String("provider", provider))
		return nil, err
	}

	s.logger.Info("SSO login initiated",
		logging.String("provider", provider))

	return authURL, nil
}

// InitiateMFA initiates MFA challenge
func (s *KeycloakApplicationService) InitiateMFA(ctx context.Context, userID string) (*models.MFAChallenge, error) {
	s.logger.Info("Initiating MFA challenge",
		logging.String("user_id", userID))

	challenge, err := s.adapter.InitiateMFA(ctx, userID)
	if err != nil {
		s.logger.Error("MFA initiation failed",
			logging.Error(err),
			logging.String("user_id", userID))
		return nil, err
	}

	s.logger.Info("MFA challenge initiated",
		logging.String("user_id", userID),
		logging.String("challenge_id", challenge.ChallengeID))

	return challenge, nil
}

// VerifyMFA verifies MFA challenge response
func (s *KeycloakApplicationService) VerifyMFA(ctx context.Context, challengeID, code string) (*models.MFAVerification, error) {
	s.logger.Info("Verifying MFA challenge",
		logging.String("challenge_id", challengeID))

	verification, err := s.adapter.VerifyMFA(ctx, challengeID, code)
	if err != nil {
		s.logger.Error("MFA verification failed",
			logging.Error(err),
			logging.String("challenge_id", challengeID))
		return nil, err
	}

	s.logger.Info("MFA verification successful",
		logging.String("challenge_id", challengeID))

	return verification, nil
}

// HealthCheck performs a health check against Keycloak
func (s *KeycloakApplicationService) HealthCheck(ctx context.Context) error {
	s.logger.Debug("Performing Keycloak health check")

	err := s.adapter.HealthCheck(ctx)
	if err != nil {
		s.logger.Error("Keycloak health check failed", logging.Error(err))
		return err
	}

	s.logger.Debug("Keycloak health check successful")

	return nil
}

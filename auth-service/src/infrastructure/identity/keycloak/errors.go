package keycloak

import "errors"

var (
	// Configuration errors
	ErrMissingBaseURL      = errors.New("keycloak base URL is required")
	ErrMissingRealm        = errors.New("keycloak realm is required")
	ErrMissingClientID     = errors.New("keycloak client ID is required")
	ErrMissingClientSecret = errors.New("keycloak client secret is required")

	// Authentication errors
	ErrAuthenticationFailed = errors.New("keycloak authentication failed")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrTokenExpired         = errors.New("token has expired")
	ErrInvalidToken         = errors.New("invalid token")

	// Authorization errors
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidResource  = errors.New("invalid resource")

	// User errors
	ErrUserNotFound  = errors.New("user not found")
	ErrInvalidUserID = errors.New("invalid user ID")

	// MFA errors
	ErrMFARequired      = errors.New("MFA verification required")
	ErrMFAFailed        = errors.New("MFA verification failed")
	ErrInvalidMFAMethod = errors.New("invalid MFA method")

	// General errors
	ErrKeycloakUnavailable = errors.New("keycloak service unavailable")
	ErrInvalidResponse     = errors.New("invalid response from keycloak")
)

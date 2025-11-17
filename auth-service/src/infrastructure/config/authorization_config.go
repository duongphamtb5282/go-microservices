package config

// AuthorizationMode defines how authorization is performed
type AuthorizationMode string

const (
	// AuthorizationModeJWT uses JWT claims only (roles, permissions from token)
	AuthorizationModeJWT AuthorizationMode = "jwt"

	// AuthorizationModeJWTWithDB uses JWT with database lookup for roles/permissions
	AuthorizationModeJWTWithDB AuthorizationMode = "jwt_with_db"

	// AuthorizationModeKeycloak uses Keycloak for authorization checks
	AuthorizationModeKeycloak AuthorizationMode = "keycloak"
)

// IdentityProviderMode defines which identity provider to use for authentication
type IdentityProviderMode string

const (
	// IdentityProviderDatabase uses database for user authentication
	IdentityProviderDatabase IdentityProviderMode = "database"

	// IdentityProviderKeycloak uses Keycloak for user authentication
	IdentityProviderKeycloak IdentityProviderMode = "keycloak"

	// IdentityProviderPingAM uses PingAM for user authentication
	IdentityProviderPingAM IdentityProviderMode = "pingam"
)

// AuthorizationConfig holds authorization configuration
type AuthorizationConfig struct {
	// IdentityProvider determines which identity provider to use for authentication
	IdentityProvider IdentityProviderMode `yaml:"identity_provider" mapstructure:"identity_provider"`

	// Mode determines which authorization method to use (jwt, jwt_with_db, or keycloak)
	Mode AuthorizationMode `yaml:"mode" mapstructure:"mode"`

	// Enabled enables/disables authorization checks
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// JWTAuth holds JWT-based authorization settings
	JWTAuth JWTAuthConfig `yaml:"jwt_auth" mapstructure:"jwt_auth"`

	// JWTWithDBAuth holds JWT with database authorization settings
	JWTWithDBAuth JWTWithDBAuthConfig `yaml:"jwt_with_db_auth" mapstructure:"jwt_with_db_auth"`

	// KeycloakAuth holds Keycloak authorization settings
	KeycloakAuth KeycloakAuthConfig `yaml:"keycloak_auth" mapstructure:"keycloak_auth"`
}

// JWTAuthConfig holds JWT authorization settings (token-based only)
type JWTAuthConfig struct {
	// UseRoles enables role-based authorization from JWT
	UseRoles bool `yaml:"use_roles" mapstructure:"use_roles"`

	// UsePermissions enables permission-based authorization from JWT
	UsePermissions bool `yaml:"use_permissions" mapstructure:"use_permissions"`

	// RolesClaimKey is the JWT claim key for roles (default: "roles")
	RolesClaimKey string `yaml:"roles_claim_key" mapstructure:"roles_claim_key"`

	// PermissionsClaimKey is the JWT claim key for permissions (default: "permissions")
	PermissionsClaimKey string `yaml:"permissions_claim_key" mapstructure:"permissions_claim_key"`
}

// JWTWithDBAuthConfig holds JWT with database authorization settings
type JWTWithDBAuthConfig struct {
	// UseRoles enables role-based authorization from database
	UseRoles bool `yaml:"use_roles" mapstructure:"use_roles"`

	// UsePermissions enables permission-based authorization from database
	UsePermissions bool `yaml:"use_permissions" mapstructure:"use_permissions"`

	// CacheTTL defines how long to cache role/permission lookups
	CacheTTL string `yaml:"cache_ttl" mapstructure:"cache_ttl"`

	// RefreshOnAccess refreshes cache TTL on access
	RefreshOnAccess bool `yaml:"refresh_on_access" mapstructure:"refresh_on_access"`
}

// KeycloakAuthConfig holds Keycloak authorization settings
type KeycloakAuthConfig struct {
	// Enabled enables Keycloak authorization
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`

	// CacheTTL defines how long to cache permission check results
	CacheTTL string `yaml:"cache_ttl" mapstructure:"cache_ttl"`

	// FallbackToJWT falls back to JWT auth if Keycloak is unavailable
	FallbackToJWT bool `yaml:"fallback_to_jwt" mapstructure:"fallback_to_jwt"`

	// UseAuthorizationServices enables Keycloak Authorization Services
	UseAuthorizationServices bool `yaml:"use_authorization_services" mapstructure:"use_authorization_services"`
}

// NewAuthorizationConfig returns default authorization configuration with minimal hardcoding
func NewAuthorizationConfig() AuthorizationConfig {
	return AuthorizationConfig{
		IdentityProvider: IdentityProviderDatabase, // Default to database authentication
		Mode:             AuthorizationModeJWTWithDB, // Default to JWT with database
		Enabled:          true,
		JWTAuth: JWTAuthConfig{
			UseRoles:            true,
			UsePermissions:      true,
			RolesClaimKey:       "roles",
			PermissionsClaimKey: "permissions",
		},
		JWTWithDBAuth: JWTWithDBAuthConfig{
			UseRoles:        true,
			UsePermissions:  true,
			CacheTTL:        "15m",
			RefreshOnAccess: true,
		},
		KeycloakAuth: KeycloakAuthConfig{
			Enabled:                  false,
			CacheTTL:                 "5m",
			FallbackToJWT:            true,
			UseAuthorizationServices: false,
		},
	}
}

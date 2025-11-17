package keycloak

import (
	"time"
)

// KeycloakConfig holds configuration for Keycloak integration
type KeycloakConfig struct {
	BaseURL         string        `yaml:"base_url" env:"KEYCLOAK_BASE_URL"`
	Realm           string        `yaml:"realm" env:"KEYCLOAK_REALM"`
	ClientID        string        `yaml:"client_id" env:"KEYCLOAK_CLIENT_ID"`
	ClientSecret    string        `yaml:"client_secret" env:"KEYCLOAK_CLIENT_SECRET"`
	RedirectURI     string        `yaml:"redirect_uri" env:"KEYCLOAK_REDIRECT_URI"`
	Scopes          []string      `yaml:"scopes" env:"KEYCLOAK_SCOPES"`
	Timeout         time.Duration `yaml:"timeout" env:"KEYCLOAK_TIMEOUT"`
	RetryAttempts   int           `yaml:"retry_attempts" env:"KEYCLOAK_RETRY_ATTEMPTS"`
	CacheTTL        time.Duration `yaml:"cache_ttl" env:"KEYCLOAK_CACHE_TTL"`
	EnableSSO       bool          `yaml:"enable_sso" env:"KEYCLOAK_ENABLE_SSO"`
	EnableMFA       bool          `yaml:"enable_mfa" env:"KEYCLOAK_ENABLE_MFA"`
	EnableRiskBased bool          `yaml:"enable_risk_based" env:"KEYCLOAK_ENABLE_RISK_BASED"`

	// SAML Configuration
	SAML SAMLConfig `yaml:"saml"`

	// OAuth Configuration
	OAuth OAuthConfig `yaml:"oauth"`

	// Policy Configuration
	Policy PolicyConfig `yaml:"policy"`

	// Admin API Configuration
	Admin AdminConfig `yaml:"admin"`
}

// SAMLConfig holds SAML-specific configuration
type SAMLConfig struct {
	EntityID    string `yaml:"entity_id" env:"KEYCLOAK_SAML_ENTITY_ID"`
	SSOURL      string `yaml:"sso_url" env:"KEYCLOAK_SAML_SSO_URL"`
	SLOURL      string `yaml:"slo_url" env:"KEYCLOAK_SAML_SLO_URL"`
	Certificate string `yaml:"certificate" env:"KEYCLOAK_SAML_CERT_PATH"`
	PrivateKey  string `yaml:"private_key" env:"KEYCLOAK_SAML_KEY_PATH"`
	Enabled     bool   `yaml:"enabled" env:"KEYCLOAK_SAML_ENABLED"`
}

// OAuthConfig holds OAuth-specific configuration
type OAuthConfig struct {
	AuthorizationURL string   `yaml:"authorization_url" env:"KEYCLOAK_OAUTH_AUTH_URL"`
	TokenURL         string   `yaml:"token_url" env:"KEYCLOAK_OAUTH_TOKEN_URL"`
	UserInfoURL      string   `yaml:"userinfo_url" env:"KEYCLOAK_OAUTH_USERINFO_URL"`
	LogoutURL        string   `yaml:"logout_url" env:"KEYCLOAK_OAUTH_LOGOUT_URL"`
	JWKSURL          string   `yaml:"jwks_url" env:"KEYCLOAK_OAUTH_JWKS_URL"`
	Scopes           []string `yaml:"scopes" env:"KEYCLOAK_OAUTH_SCOPES"`
	Enabled          bool     `yaml:"enabled" env:"KEYCLOAK_OAUTH_ENABLED"`
}

// PolicyConfig holds policy engine configuration
type PolicyConfig struct {
	PolicyURL   string `yaml:"policy_url" env:"KEYCLOAK_POLICY_URL"`
	DecisionURL string `yaml:"decision_url" env:"KEYCLOAK_DECISION_URL"`
	EnableRBAC  bool   `yaml:"enable_rbac" env:"KEYCLOAK_ENABLE_RBAC"`
	Enabled     bool   `yaml:"enabled" env:"KEYCLOAK_POLICY_ENABLED"`
}

// AdminConfig holds Keycloak Admin API configuration
type AdminConfig struct {
	Username string `yaml:"username" env:"KEYCLOAK_ADMIN_USERNAME"`
	Password string `yaml:"password" env:"KEYCLOAK_ADMIN_PASSWORD"`
	Enabled  bool   `yaml:"enabled" env:"KEYCLOAK_ADMIN_ENABLED"`
}

// NewKeycloakConfig creates a new Keycloak configuration from environment-aware defaults
func NewKeycloakConfig() *KeycloakConfig {
	return &KeycloakConfig{
		BaseURL:         "",
		Realm:           "",
		ClientID:        "",
		ClientSecret:    "",
		RedirectURI:     "",
		Scopes:          []string{"openid", "profile", "email", "roles"},
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		CacheTTL:        5 * time.Minute,
		EnableSSO:       true,
		EnableMFA:       true,
		EnableRiskBased: false,
		SAML: SAMLConfig{
			EntityID:    "",
			SSOURL:      "",
			SLOURL:      "",
			Certificate: "",
			PrivateKey:  "",
			Enabled:     false,
		},
		OAuth: OAuthConfig{
			AuthorizationURL: "",
			TokenURL:         "",
			UserInfoURL:      "",
			LogoutURL:        "",
			JWKSURL:          "",
			Scopes:           []string{"openid", "profile", "email", "roles"},
			Enabled:          true,
		},
		Policy: PolicyConfig{
			PolicyURL:   "",
			DecisionURL: "",
			EnableRBAC:  true,
			Enabled:     false,
		},
		Admin: AdminConfig{
			Username: "",
			Password: "",
			Enabled:  false,
		},
	}
}

// BuildURLs builds the OAuth URLs from the base URL and realm
func (c *KeycloakConfig) BuildURLs() {
	if c.BaseURL != "" && c.Realm != "" {
		baseRealmURL := c.BaseURL + "/realms/" + c.Realm
		if c.OAuth.AuthorizationURL == "" {
			c.OAuth.AuthorizationURL = baseRealmURL + "/protocol/openid-connect/auth"
		}
		if c.OAuth.TokenURL == "" {
			c.OAuth.TokenURL = baseRealmURL + "/protocol/openid-connect/token"
		}
		if c.OAuth.UserInfoURL == "" {
			c.OAuth.UserInfoURL = baseRealmURL + "/protocol/openid-connect/userinfo"
		}
		if c.OAuth.LogoutURL == "" {
			c.OAuth.LogoutURL = baseRealmURL + "/protocol/openid-connect/logout"
		}
		if c.OAuth.JWKSURL == "" {
			c.OAuth.JWKSURL = baseRealmURL + "/protocol/openid-connect/certs"
		}
	}
}

// Validate validates the Keycloak configuration
func (c *KeycloakConfig) Validate() error {
	if c.BaseURL == "" {
		return ErrMissingBaseURL
	}
	if c.Realm == "" {
		return ErrMissingRealm
	}
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if c.ClientSecret == "" {
		return ErrMissingClientSecret
	}
	return nil
}

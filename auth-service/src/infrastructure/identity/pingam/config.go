package pingam

import (
	"time"
)

// PingAMConfig holds configuration for PingAM integration
type PingAMConfig struct {
	BaseURL         string        `yaml:"base_url" env:"PINGAM_BASE_URL"`
	ClientID        string        `yaml:"client_id" env:"PINGAM_CLIENT_ID"`
	ClientSecret    string        `yaml:"client_secret" env:"PINGAM_CLIENT_SECRET"`
	RedirectURI     string        `yaml:"redirect_uri" env:"PINGAM_REDIRECT_URI"`
	Scopes          []string      `yaml:"scopes" env:"PINGAM_SCOPES"`
	Timeout         time.Duration `yaml:"timeout" env:"PINGAM_TIMEOUT"`
	RetryAttempts   int           `yaml:"retry_attempts" env:"PINGAM_RETRY_ATTEMPTS"`
	CacheTTL        time.Duration `yaml:"cache_ttl" env:"PINGAM_CACHE_TTL"`
	EnableSSO       bool          `yaml:"enable_sso" env:"PINGAM_ENABLE_SSO"`
	EnableMFA       bool          `yaml:"enable_mfa" env:"PINGAM_ENABLE_MFA"`
	EnableRiskBased bool          `yaml:"enable_risk_based" env:"PINGAM_ENABLE_RISK_BASED"`

	// SAML Configuration
	SAML SAMLConfig `yaml:"saml"`

	// OAuth Configuration
	OAuth OAuthConfig `yaml:"oauth"`

	// Policy Configuration
	Policy PolicyConfig `yaml:"policy"`
}

// SAMLConfig holds SAML-specific configuration
type SAMLConfig struct {
	EntityID    string `yaml:"entity_id" env:"PINGAM_SAML_ENTITY_ID"`
	SSOURL      string `yaml:"sso_url" env:"PINGAM_SAML_SSO_URL"`
	SLOURL      string `yaml:"slo_url" env:"PINGAM_SAML_SLO_URL"`
	Certificate string `yaml:"certificate" env:"PINGAM_SAML_CERT_PATH"`
	PrivateKey  string `yaml:"private_key" env:"PINGAM_SAML_KEY_PATH"`
	Enabled     bool   `yaml:"enabled" env:"PINGAM_SAML_ENABLED"`
}

// OAuthConfig holds OAuth-specific configuration
type OAuthConfig struct {
	AuthorizationURL string   `yaml:"authorization_url" env:"PINGAM_OAUTH_AUTH_URL"`
	TokenURL         string   `yaml:"token_url" env:"PINGAM_OAUTH_TOKEN_URL"`
	UserInfoURL      string   `yaml:"userinfo_url" env:"PINGAM_OAUTH_USERINFO_URL"`
	Scopes           []string `yaml:"scopes" env:"PINGAM_OAUTH_SCOPES"`
	Enabled          bool     `yaml:"enabled" env:"PINGAM_OAUTH_ENABLED"`
}

// PolicyConfig holds policy engine configuration
type PolicyConfig struct {
	PolicyURL   string `yaml:"policy_url" env:"PINGAM_POLICY_URL"`
	DecisionURL string `yaml:"decision_url" env:"PINGAM_DECISION_URL"`
	EnableRBAC  bool   `yaml:"enable_rbac" env:"PINGAM_ENABLE_RBAC"`
	Enabled     bool   `yaml:"enabled" env:"PINGAM_POLICY_ENABLED"`
}

// DefaultPingAMConfig returns a default PingAM configuration
func DefaultPingAMConfig() *PingAMConfig {
	return &PingAMConfig{
		BaseURL:         "https://pingam.company.com",
		ClientID:        "auth-service-client",
		ClientSecret:    "your-client-secret",
		RedirectURI:     "https://auth-service.company.com/callback",
		Scopes:          []string{"openid", "profile", "email", "groups"},
		Timeout:         30 * time.Second,
		RetryAttempts:   3,
		CacheTTL:        5 * time.Minute,
		EnableSSO:       true,
		EnableMFA:       true,
		EnableRiskBased: true,
		SAML: SAMLConfig{
			EntityID:    "auth-service",
			SSOURL:      "https://pingam.company.com/saml/sso",
			SLOURL:      "https://pingam.company.com/saml/slo",
			Certificate: "/etc/ssl/pingam/cert.pem",
			PrivateKey:  "/etc/ssl/pingam/key.pem",
			Enabled:     true,
		},
		OAuth: OAuthConfig{
			AuthorizationURL: "https://pingam.company.com/oauth/authorize",
			TokenURL:         "https://pingam.company.com/oauth/token",
			UserInfoURL:      "https://pingam.company.com/oauth/userinfo",
			Scopes:           []string{"openid", "profile", "email", "groups"},
			Enabled:          true,
		},
		Policy: PolicyConfig{
			PolicyURL:   "https://pingam.company.com/policy",
			DecisionURL: "https://pingam.company.com/authorize",
			EnableRBAC:  true,
			Enabled:     true,
		},
	}
}

package models

import (
	"time"
)

// Credentials represents user authentication credentials
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResult represents the result of authentication
type AuthResult struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	IssuedAt     time.Time `json:"issued_at"`
}

// TokenInfo represents token information from PingAM
type TokenInfo struct {
	Active    bool      `json:"active"`
	Scope     string    `json:"scope"`
	ClientID  string    `json:"client_id"`
	Username  string    `json:"username"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// UserProfile represents user profile information from PingAM
type UserProfile struct {
	ID         string                 `json:"id"`
	Username   string                 `json:"username"`
	Email      string                 `json:"email"`
	FirstName  string                 `json:"first_name"`
	LastName   string                 `json:"last_name"`
	Groups     []string               `json:"groups"`
	Roles      []string               `json:"roles"`
	Attributes map[string]interface{} `json:"attributes"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// PermissionResult represents the result of a permission check
type PermissionResult struct {
	Allowed bool     `json:"allowed"`
	Reason  string   `json:"reason"`
	Roles   []string `json:"roles"`
}

// SessionInfo represents session information
type SessionInfo struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	SessionToken string                 `json:"session_token"`
	Attributes   map[string]interface{} `json:"attributes"`
	CreatedAt    time.Time              `json:"created_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
	LastAccess   time.Time              `json:"last_access"`
	IsActive     bool                   `json:"is_active"`
}

// SSOURL represents a SAML SSO URL
type SSOURL struct {
	URL        string `json:"url"`
	RelayState string `json:"relay_state"`
}

// SAMLResult represents the result of SAML processing
type SAMLResult struct {
	UserID     string            `json:"user_id"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	Attributes map[string]string `json:"attributes"`
	SessionID  string            `json:"session_id"`
}

// AuthURL represents an OAuth authorization URL
type AuthURL struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// UserInfo represents OAuth user information
type UserInfo struct {
	Sub           string   `json:"sub"`
	Name          string   `json:"name"`
	GivenName     string   `json:"given_name"`
	FamilyName    string   `json:"family_name"`
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Groups        []string `json:"groups"`
	Roles         []string `json:"roles"`
}

// PolicyDecision represents a policy evaluation result
type PolicyDecision struct {
	Decision    string                 `json:"decision"`
	Reason      string                 `json:"reason"`
	Obligations map[string]interface{} `json:"obligations"`
}

// RiskAssessment represents risk assessment result
type RiskAssessment struct {
	RiskScore       float64  `json:"risk_score"`
	RiskLevel       string   `json:"risk_level"`
	Factors         []string `json:"factors"`
	Recommendations []string `json:"recommendations"`
}

// MFAChallenge represents MFA challenge information
type MFAChallenge struct {
	ChallengeID string    `json:"challenge_id"`
	Methods     []string  `json:"methods"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// MFAVerification represents MFA verification result
type MFAVerification struct {
	Success    bool      `json:"success"`
	Method     string    `json:"method"`
	VerifiedAt time.Time `json:"verified_at"`
}

// AuditEvent represents an audit event
type AuditEvent struct {
	EventID    string                 `json:"event_id"`
	UserID     string                 `json:"user_id"`
	EventType  string                 `json:"event_type"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Result     string                 `json:"result"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
}

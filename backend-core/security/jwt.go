package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
	audience             string
}

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessTokenDuration time.Duration, issuer, audience string) *JWTManager {
	return &JWTManager{
		secretKey:            secretKey,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: 7 * 24 * time.Hour, // 7 days for refresh tokens
		issuer:               issuer,
		audience:             audience,
	}
}

// GenerateAccessToken generates a new access JWT token
func (j *JWTManager) GenerateAccessToken(userID, username, role string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateRefreshToken generates a new refresh JWT token
func (j *JWTManager) GenerateRefreshToken(userID, username, role string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(userID, username, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessToken(userID, username, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = j.GenerateRefreshToken(userID, username, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateToken generates a new JWT access token (backward compatibility)
func (j *JWTManager) GenerateToken(userID, username, role string) (string, error) {
	return j.GenerateAccessToken(userID, username, role)
}

// ValidateToken validates a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTManager) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return j.GenerateAccessToken(claims.UserID, claims.Username, claims.Role)
}

// RefreshTokenPair generates new access and refresh tokens using a valid refresh token
func (j *JWTManager) RefreshTokenPair(refreshTokenString string) (accessToken, newRefreshToken string, err error) {
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Username, claims.Role)
}

// ValidateRefreshToken validates a refresh token and ensures it's actually a refresh token
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Ensure this is actually a refresh token
	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	return claims, nil
}

// ValidateAccessToken validates an access token and ensures it's actually an access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Ensure this is actually an access token
	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type: expected access token")
	}

	return claims, nil
}

// RefreshToken generates a new access token (backward compatibility)
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	return j.RefreshAccessToken(tokenString)
}

// GetTokenExpiration returns the token expiration time
func (j *JWTManager) GetTokenExpiration(tokenString string) (time.Time, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired
func (j *JWTManager) IsTokenExpired(tokenString string) bool {
	expiration, err := j.GetTokenExpiration(tokenString)
	if err != nil {
		return true
	}

	return time.Now().After(expiration)
}

// JWTConfig holds JWT configuration for validation
type JWTConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

// ValidateJWT validates a JWT token and returns claims
func ValidateJWT(tokenString string, config *JWTConfig) (map[string]interface{}, error) {
	// Create a temporary JWTManager for validation
	manager := NewJWTManager(config.Secret, time.Hour, config.Issuer, config.Audience)

	// Validate the token
	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Convert Claims to map for compatibility
	result := map[string]interface{}{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"iss":      claims.Issuer,
		"aud":      claims.Audience[0],
		"exp":      claims.ExpiresAt.Unix(),
		"iat":      claims.IssuedAt.Unix(),
		"nbf":      claims.NotBefore.Unix(),
	}

	return result, nil
}

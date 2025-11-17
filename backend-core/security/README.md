# Security Package

This package provides a comprehensive security solution with support for authentication, authorization, JWT tokens, password hashing, and security utilities.

## Package Structure

```
security/
├── auth.go              # Authentication utilities
├── jwt.go               # JWT token management
└── security_manager.go  # Security manager
```

## Core Components

### 1. Authentication (`auth.go`)

The authentication module provides user authentication utilities:

```go
type AuthService struct {
    jwtService JWTService
    userService UserService
    logger      Logger
}

// Authenticate user with credentials
func (a *AuthService) Authenticate(ctx context.Context, email, password string) (*AuthResult, error)

// Refresh authentication token
func (a *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)

// Logout user
func (a *AuthService) Logout(ctx context.Context, token string) error

// Validate user session
func (a *AuthService) ValidateSession(ctx context.Context, token string) (*User, error)
```

### 2. JWT Service (`jwt.go`)

The JWT service provides JSON Web Token management:

```go
type JWTService struct {
    secret     string
    expiry     time.Duration
    refreshExpiry time.Duration
    issuer     string
    audience   string
}

// Generate access token
func (j *JWTService) GenerateAccessToken(user *User) (string, error)

// Generate refresh token
func (j *JWTService) GenerateRefreshToken(user *User) (string, error)

// Validate token
func (j *JWTService) ValidateToken(token string) (*Claims, error)

// Extract claims from token
func (j *JWTService) ExtractClaims(token string) (*Claims, error)

// Refresh token
func (j *JWTService) RefreshToken(refreshToken string) (string, error)
```

### 3. Security Manager (`security_manager.go`)

The security manager provides centralized security management:

```go
type SecurityManager struct {
    authService    AuthService
    jwtService     JWTService
    passwordHasher  PasswordHasher
    rateLimiter    RateLimiter
    logger         Logger
}

// Initialize security manager
func NewSecurityManager(config *SecurityConfig) (*SecurityManager, error)

// Authenticate user
func (sm *SecurityManager) Authenticate(ctx context.Context, credentials *Credentials) (*AuthResult, error)

// Authorize user
func (sm *SecurityManager) Authorize(ctx context.Context, user *User, resource string, action string) (bool, error)

// Hash password
func (sm *SecurityManager) HashPassword(password string) (string, error)

// Verify password
func (sm *SecurityManager) VerifyPassword(password, hash string) (bool, error)
```

## Authentication

### User Authentication

```go
// Create authentication service
authService := NewAuthService(&AuthConfig{
    JWTSecret:     "your-secret-key",
    JWTExpiry:     15 * time.Minute,
    RefreshExpiry: 7 * 24 * time.Hour,
    Issuer:        "your-app",
    Audience:      "your-users",
})

// Authenticate user
result, err := authService.Authenticate(ctx, "user@example.com", "password")
if err != nil {
    return err
}

// Use authentication result
accessToken := result.AccessToken
refreshToken := result.RefreshToken
user := result.User
```

### Token Validation

```go
// Validate access token
claims, err := authService.ValidateToken(accessToken)
if err != nil {
    return err
}

// Extract user information
userID := claims.UserID
email := claims.Email
roles := claims.Roles
```

### Refresh Token

```go
// Refresh access token
newResult, err := authService.RefreshToken(ctx, refreshToken)
if err != nil {
    return err
}

// Use new tokens
newAccessToken := newResult.AccessToken
newRefreshToken := newResult.RefreshToken
```

## JWT Token Management

### Token Generation

```go
// Create JWT service
jwtService := NewJWTService(&JWTConfig{
    Secret:        "your-secret-key",
    Expiry:        15 * time.Minute,
    RefreshExpiry: 7 * 24 * time.Hour,
    Issuer:        "your-app",
    Audience:      "your-users",
})

// Generate access token
user := &User{
    ID:    "123",
    Email: "user@example.com",
    Roles: []string{"user", "admin"},
}

accessToken, err := jwtService.GenerateAccessToken(user)
if err != nil {
    return err
}

// Generate refresh token
refreshToken, err := jwtService.GenerateRefreshToken(user)
if err != nil {
    return err
}
```

### Token Validation

```go
// Validate token
claims, err := jwtService.ValidateToken(accessToken)
if err != nil {
    return err
}

// Check token expiration
if claims.ExpiresAt < time.Now().Unix() {
    return errors.New("token expired")
}

// Extract user information
userID := claims.UserID
email := claims.Email
roles := claims.Roles
```

### Custom Claims

```go
// Custom claims structure
type CustomClaims struct {
    jwt.StandardClaims
    UserID    string   `json:"user_id"`
    Email     string   `json:"email"`
    Roles     []string `json:"roles"`
    Permissions []string `json:"permissions"`
}

// Generate token with custom claims
claims := &CustomClaims{
    StandardClaims: jwt.StandardClaims{
        Subject:   user.ID,
        Issuer:    "your-app",
        Audience:  "your-users",
        ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
        IssuedAt:  time.Now().Unix(),
    },
    UserID:      user.ID,
    Email:       user.Email,
    Roles:       user.Roles,
    Permissions: user.Permissions,
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken, err := token.SignedString([]byte(secret))
```

## Password Security

### Password Hashing

```go
// Create password hasher
hasher := NewPasswordHasher(&PasswordConfig{
    Cost: 12, // bcrypt cost
})

// Hash password
hashedPassword, err := hasher.HashPassword("user-password")
if err != nil {
    return err
}

// Store hashed password in database
user.PasswordHash = hashedPassword
```

### Password Verification

```go
// Verify password
isValid, err := hasher.VerifyPassword("user-password", user.PasswordHash)
if err != nil {
    return err
}

if !isValid {
    return errors.New("invalid password")
}
```

### Password Validation

```go
// Validate password strength
validator := NewPasswordValidator(&PasswordValidatorConfig{
    MinLength:     8,
    MaxLength:     128,
    RequireUpper:  true,
    RequireLower:  true,
    RequireNumber: true,
    RequireSpecial: true,
})

err := validator.ValidatePassword("user-password")
if err != nil {
    return err
}
```

## Authorization

### Role-Based Access Control

```go
// Create authorization service
authzService := NewAuthorizationService(&AuthzConfig{
    Roles: map[string][]string{
        "admin": {"read", "write", "delete", "manage"},
        "user":  {"read", "write"},
        "guest": {"read"},
    },
})

// Check user role
hasPermission, err := authzService.HasRole(user, "admin")
if err != nil {
    return err
}

// Check user permission
hasPermission, err := authzService.HasPermission(user, "write")
if err != nil {
    return err
}
```

### Resource-Based Authorization

```go
// Check resource access
hasAccess, err := authzService.CanAccess(user, "user", "read", "123")
if err != nil {
    return err
}

// Check resource ownership
isOwner, err := authzService.IsOwner(user, "user", "123")
if err != nil {
    return err
}
```

### Permission-Based Authorization

```go
// Check specific permission
hasPermission, err := authzService.HasPermission(user, "user:read")
if err != nil {
    return err
}

// Check multiple permissions
hasAllPermissions, err := authzService.HasAllPermissions(user, []string{"user:read", "user:write"})
if err != nil {
    return err
}

// Check any permission
hasAnyPermission, err := authzService.HasAnyPermission(user, []string{"user:read", "user:write"})
if err != nil {
    return err
}
```

## Rate Limiting

### Request Rate Limiting

```go
// Create rate limiter
rateLimiter := NewRateLimiter(&RateLimiterConfig{
    Requests: 100,
    Window:   1 * time.Minute,
    Burst:    10,
})

// Check rate limit
allowed, err := rateLimiter.Allow("user:123")
if err != nil {
    return err
}

if !allowed {
    return errors.New("rate limit exceeded")
}
```

### IP-Based Rate Limiting

```go
// Create IP rate limiter
ipRateLimiter := NewIPRateLimiter(&IPRateLimiterConfig{
    Requests: 1000,
    Window:   1 * time.Minute,
    Burst:    100,
})

// Check IP rate limit
allowed, err := ipRateLimiter.Allow("192.168.1.1")
if err != nil {
    return err
}
```

### User-Based Rate Limiting

```go
// Create user rate limiter
userRateLimiter := NewUserRateLimiter(&UserRateLimiterConfig{
    Requests: 100,
    Window:   1 * time.Minute,
    Burst:    10,
})

// Check user rate limit
allowed, err := userRateLimiter.Allow("user:123")
if err != nil {
    return err
}
```

## Security Headers

### HTTP Security Headers

```go
// Create security headers middleware
securityHeaders := NewSecurityHeaders(&SecurityHeadersConfig{
    XSSProtection:     true,
    ContentTypeNosniff: true,
    FrameOptions:     "DENY",
    HSTS:             true,
    HSTSDuration:     31536000, // 1 year
    CSP:              "default-src 'self'",
})

// Apply security headers
func (sh *SecurityHeaders) Apply(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("X-XSS-Protection", "1; mode=block")
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("X-Frame-Options", "DENY")
    w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
    w.Header().Set("Content-Security-Policy", "default-src 'self'")
}
```

### CORS Configuration

```go
// Create CORS middleware
cors := NewCORS(&CORSConfig{
    AllowedOrigins: []string{"https://example.com"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    ExposedHeaders: []string{"X-Total-Count"},
    MaxAge:         86400, // 24 hours
    Credentials:     true,
})

// Apply CORS
func (c *CORS) Apply(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    if c.isAllowedOrigin(origin) {
        w.Header().Set("Access-Control-Allow-Origin", origin)
        w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.AllowedMethods, ", "))
        w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.AllowedHeaders, ", "))
        w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.MaxAge))
    }
}
```

## Input Validation

### Request Validation

```go
// Create request validator
validator := NewRequestValidator(&ValidatorConfig{
    MaxBodySize: 10 * 1024 * 1024, // 10MB
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedContentTypes: []string{"application/json", "application/x-www-form-urlencoded"},
})

// Validate request
err := validator.ValidateRequest(r)
if err != nil {
    return err
}
```

### Input Sanitization

```go
// Create input sanitizer
sanitizer := NewInputSanitizer(&SanitizerConfig{
    AllowedTags: []string{"b", "i", "em", "strong"},
    AllowedAttributes: []string{"class", "id"},
    MaxLength: 1000,
})

// Sanitize input
sanitized := sanitizer.Sanitize(input)
```

## Security Monitoring

### Security Events

```go
// Create security event logger
securityLogger := NewSecurityLogger(&SecurityLoggerConfig{
    Level: "info",
    Format: "json",
    Output: "file",
    File: "/var/log/security.log",
})

// Log security events
securityLogger.LogLoginAttempt("user@example.com", "192.168.1.1", true)
securityLogger.LogFailedLogin("user@example.com", "192.168.1.1", "invalid_password")
securityLogger.LogSuspiciousActivity("user@example.com", "multiple_failed_logins")
```

### Security Metrics

```go
// Create security metrics
metrics := NewSecurityMetrics(&SecurityMetricsConfig{
    Enabled: true,
    Port: 9090,
})

// Track security metrics
metrics.IncrementLoginAttempts()
metrics.IncrementFailedLogins()
metrics.IncrementSuspiciousActivities()
```

## Configuration

### Security Configuration

```go
type SecurityConfig struct {
    JWT JWTConfig `yaml:"jwt" json:"jwt"`
    Password PasswordConfig `yaml:"password" json:"password"`
    RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
    Headers SecurityHeadersConfig `yaml:"headers" json:"headers"`
    CORS CORSConfig `yaml:"cors" json:"cors"`
}

type JWTConfig struct {
    Secret string `yaml:"secret" json:"secret"`
    Expiry time.Duration `yaml:"expiry" json:"expiry"`
    RefreshExpiry time.Duration `yaml:"refresh_expiry" json:"refresh_expiry"`
    Issuer string `yaml:"issuer" json:"issuer"`
    Audience string `yaml:"audience" json:"audience"`
}

type PasswordConfig struct {
    Cost int `yaml:"cost" json:"cost"`
    MinLength int `yaml:"min_length" json:"min_length"`
    MaxLength int `yaml:"max_length" json:"max_length"`
    RequireUpper bool `yaml:"require_upper" json:"require_upper"`
    RequireLower bool `yaml:"require_lower" json:"require_lower"`
    RequireNumber bool `yaml:"require_number" json:"require_number"`
    RequireSpecial bool `yaml:"require_special" json:"require_special"`
}
```

### Environment Configuration

```go
// Load security configuration from environment
config := &SecurityConfig{
    JWT: JWTConfig{
        Secret: getEnv("JWT_SECRET", "your-secret-key"),
        Expiry: getEnvDuration("JWT_EXPIRY", 15*time.Minute),
        RefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
        Issuer: getEnv("JWT_ISSUER", "your-app"),
        Audience: getEnv("JWT_AUDIENCE", "your-users"),
    },
    Password: PasswordConfig{
        Cost: getEnvInt("PASSWORD_COST", 12),
        MinLength: getEnvInt("PASSWORD_MIN_LENGTH", 8),
        MaxLength: getEnvInt("PASSWORD_MAX_LENGTH", 128),
        RequireUpper: getEnvBool("PASSWORD_REQUIRE_UPPER", true),
        RequireLower: getEnvBool("PASSWORD_REQUIRE_LOWER", true),
        RequireNumber: getEnvBool("PASSWORD_REQUIRE_NUMBER", true),
        RequireSpecial: getEnvBool("PASSWORD_REQUIRE_SPECIAL", true),
    },
}
```

## Best Practices

### 1. Password Security

- Use strong password requirements
- Hash passwords with bcrypt
- Use appropriate cost factor
- Never store plain text passwords

### 2. JWT Security

- Use strong secret keys
- Set appropriate expiration times
- Validate token signatures
- Use HTTPS in production

### 3. Rate Limiting

- Implement rate limiting for all endpoints
- Use different limits for different operations
- Monitor rate limit violations
- Implement progressive delays

### 4. Input Validation

- Validate all input data
- Sanitize user input
- Use whitelist validation
- Implement length limits

### 5. Security Headers

- Set appropriate security headers
- Use HTTPS in production
- Implement CORS properly
- Monitor security events

## Examples

### Complete Security Setup

```go
func main() {
    // Load security configuration
    config, err := LoadSecurityConfig("security.yaml")
    if err != nil {
        log.Fatal("Failed to load security config", err)
    }

    // Create security manager
    securityManager, err := NewSecurityManager(config)
    if err != nil {
        log.Fatal("Failed to create security manager", err)
    }

    // Create HTTP server with security middleware
    server := &http.Server{
        Addr: ":8080",
        Handler: securityMiddleware(securityManager, http.DefaultServeMux),
    }

    // Start server
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start", err)
    }
}

func securityMiddleware(securityManager *SecurityManager, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Apply security headers
        securityManager.ApplySecurityHeaders(w, r)

        // Check rate limit
        if !securityManager.CheckRateLimit(r) {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        // Process request
        next.ServeHTTP(w, r)
    })
}
```

### Authentication Middleware

```go
func authMiddleware(securityManager *SecurityManager, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from header
        token := extractToken(r)
        if token == "" {
            http.Error(w, "Authorization required", http.StatusUnauthorized)
            return
        }

        // Validate token
        claims, err := securityManager.ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add user to context
        ctx := context.WithValue(r.Context(), "user", claims)
        r = r.WithContext(ctx)

        // Process request
        next.ServeHTTP(w, r)
    })
}
```

## Migration Guide

When upgrading security:

1. **Check breaking changes** in the changelog
2. **Update JWT secrets** if needed
3. **Test authentication** in all environments
4. **Update security headers** if using new features
5. **Monitor security events** after upgrade

## Future Enhancements

- **OAuth2 integration** - Support for OAuth2 providers
- **SAML support** - SAML authentication
- **Multi-factor authentication** - MFA support
- **Biometric authentication** - Biometric login
- **Advanced threat detection** - AI-powered threat detection

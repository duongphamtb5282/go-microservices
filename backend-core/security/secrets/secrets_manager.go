package secrets

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecretsManager manages application secrets with encryption and rotation
type SecretsManager struct {
	provider SecretProvider
	logger   *zap.Logger
	secrets  map[string]*Secret
	mutex    sync.RWMutex

	// Rotation settings
	rotationInterval time.Duration
	keyRotationDays  int
}

// Secret represents a managed secret
type Secret struct {
	Name      string     `json:"name"`
	Value     string     `json:"value"`
	Encrypted bool       `json:"encrypted"`
	Version   int        `json:"version"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Tags      []string   `json:"tags"`
}

// SecretProvider interface for different secret storage backends
type SecretProvider interface {
	GetSecret(ctx context.Context, name string) (string, error)
	SetSecret(ctx context.Context, name, value string) error
	DeleteSecret(ctx context.Context, name string) error
	ListSecrets(ctx context.Context) ([]string, error)
	RotateSecret(ctx context.Context, name string) error
}

// EnvironmentSecretProvider implements secrets from environment variables
type EnvironmentSecretProvider struct {
	prefix string
	logger *zap.Logger
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(provider SecretProvider, logger *zap.Logger) *SecretsManager {
	sm := &SecretsManager{
		provider:         provider,
		logger:           logger,
		secrets:          make(map[string]*Secret),
		rotationInterval: 24 * time.Hour, // Default daily rotation check
		keyRotationDays:  90,             // Default 90-day rotation
	}

	// Start rotation checker
	go sm.startRotationChecker()

	return sm
}

// GetSecret retrieves a secret by name
func (sm *SecretsManager) GetSecret(ctx context.Context, name string) (string, error) {
	sm.mutex.RLock()
	secret, exists := sm.secrets[name]
	sm.mutex.RUnlock()

	if exists && secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		// Secret expired, remove from cache
		sm.mutex.Lock()
		delete(sm.secrets, name)
		sm.mutex.Unlock()
		exists = false
	}

	if !exists {
		// Load from provider
		value, err := sm.provider.GetSecret(ctx, name)
		if err != nil {
			return "", fmt.Errorf("failed to get secret %s: %w", name, err)
		}

		secret = &Secret{
			Name:      name,
			Value:     value,
			Encrypted: false, // Environment secrets are not encrypted
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Cache the secret
		sm.mutex.Lock()
		sm.secrets[name] = secret
		sm.mutex.Unlock()
	}

	return secret.Value, nil
}

// SetSecret stores a secret
func (sm *SecretsManager) SetSecret(ctx context.Context, name, value string, tags []string) error {
	if err := sm.provider.SetSecret(ctx, name, value); err != nil {
		return fmt.Errorf("failed to set secret %s: %w", name, err)
	}

	secret := &Secret{
		Name:      name,
		Value:     value,
		Encrypted: false,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Tags:      tags,
	}

	sm.mutex.Lock()
	sm.secrets[name] = secret
	sm.mutex.Unlock()

	sm.logger.Info("Secret stored",
		zap.String("name", name),
		zap.Int("version", secret.Version),
		zap.Strings("tags", tags),
	)

	return nil
}

// RotateSecret rotates a secret
func (sm *SecretsManager) RotateSecret(ctx context.Context, name string) error {
	sm.logger.Info("Rotating secret", zap.String("name", name))

	if err := sm.provider.RotateSecret(ctx, name); err != nil {
		return fmt.Errorf("failed to rotate secret %s: %w", name, err)
	}

	// Update cached secret
	sm.mutex.Lock()
	if secret, exists := sm.secrets[name]; exists {
		secret.Version++
		secret.UpdatedAt = time.Now()
		// Reload value from provider
		if value, err := sm.provider.GetSecret(ctx, name); err == nil {
			secret.Value = value
		}
	}
	sm.mutex.Unlock()

	sm.logger.Info("Secret rotated successfully", zap.String("name", name))
	return nil
}

// DeleteSecret removes a secret
func (sm *SecretsManager) DeleteSecret(ctx context.Context, name string) error {
	if err := sm.provider.DeleteSecret(ctx, name); err != nil {
		return fmt.Errorf("failed to delete secret %s: %w", name, err)
	}

	sm.mutex.Lock()
	delete(sm.secrets, name)
	sm.mutex.Unlock()

	sm.logger.Info("Secret deleted", zap.String("name", name))
	return nil
}

// GenerateSecret generates a cryptographically secure random secret
func (sm *SecretsManager) GenerateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ListSecrets returns all secret names
func (sm *SecretsManager) ListSecrets(ctx context.Context) ([]string, error) {
	return sm.provider.ListSecrets(ctx)
}

// GetSecretMetadata returns metadata about a secret
func (sm *SecretsManager) GetSecretMetadata(name string) (*Secret, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	secret, exists := sm.secrets[name]
	if !exists {
		return nil, fmt.Errorf("secret %s not found", name)
	}

	// Return a copy without the actual value for security
	metadata := *secret
	metadata.Value = "[REDACTED]"
	return &metadata, nil
}

// HealthCheck performs health check on secret provider
func (sm *SecretsManager) HealthCheck(ctx context.Context) error {
	secrets, err := sm.provider.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("secret provider health check failed: %w", err)
	}

	sm.logger.Debug("Secret provider health check passed", zap.Int("secrets_count", len(secrets)))
	return nil
}

// startRotationChecker periodically checks for secrets that need rotation
func (sm *SecretsManager) startRotationChecker() {
	ticker := time.NewTicker(sm.rotationInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.checkAndRotateSecrets(context.Background())
	}
}

// checkAndRotateSecrets checks and rotates secrets that are due
func (sm *SecretsManager) checkAndRotateSecrets(ctx context.Context) {
	sm.mutex.RLock()
	secrets := make(map[string]*Secret)
	for name, secret := range sm.secrets {
		secrets[name] = secret
	}
	sm.mutex.RUnlock()

	for name, secret := range secrets {
		// Check if secret needs rotation (older than keyRotationDays)
		if time.Since(secret.UpdatedAt) > time.Duration(sm.keyRotationDays)*24*time.Hour {
			sm.logger.Info("Secret due for rotation",
				zap.String("name", name),
				zap.Time("updated_at", secret.UpdatedAt),
				zap.Int("days_since_update", int(time.Since(secret.UpdatedAt).Hours()/24)),
			)

			// Attempt rotation (non-blocking)
			go func(secretName string) {
				if err := sm.RotateSecret(ctx, secretName); err != nil {
					sm.logger.Error("Failed to rotate secret",
						zap.String("name", secretName),
						zap.Error(err),
					)
				}
			}(name)
		}
	}
}

// Environment provider implementation
func NewEnvironmentSecretProvider(prefix string, logger *zap.Logger) *EnvironmentSecretProvider {
	return &EnvironmentSecretProvider{
		prefix: prefix,
		logger: logger,
	}
}

func (esp *EnvironmentSecretProvider) GetSecret(ctx context.Context, name string) (string, error) {
	// This is a simplified implementation
	// In a real scenario, you would use a proper secret management system
	// like AWS Secrets Manager, HashiCorp Vault, etc.

	// For now, we'll return a placeholder
	esp.logger.Warn("Using placeholder secret provider - implement proper secret management",
		zap.String("secret_name", name),
	)

	// This should be replaced with actual secret retrieval
	return fmt.Sprintf("placeholder_secret_%s", name), nil
}

func (esp *EnvironmentSecretProvider) SetSecret(ctx context.Context, name, value string) error {
	esp.logger.Warn("SetSecret called on read-only environment provider",
		zap.String("secret_name", name),
	)
	return fmt.Errorf("environment provider is read-only")
}

func (esp *EnvironmentSecretProvider) DeleteSecret(ctx context.Context, name string) error {
	esp.logger.Warn("DeleteSecret called on read-only environment provider",
		zap.String("secret_name", name),
	)
	return fmt.Errorf("environment provider is read-only")
}

func (esp *EnvironmentSecretProvider) ListSecrets(ctx context.Context) ([]string, error) {
	// Return common secret names that should be managed
	return []string{
		"database_password",
		"redis_password",
		"jwt_secret",
		"kafka_password",
		"email_api_key",
		"sms_api_key",
	}, nil
}

func (esp *EnvironmentSecretProvider) RotateSecret(ctx context.Context, name string) error {
	esp.logger.Warn("RotateSecret called on read-only environment provider",
		zap.String("secret_name", name),
	)
	return fmt.Errorf("environment provider does not support rotation")
}

// SecretValidator validates secret strength and format
type SecretValidator struct{}

// ValidateSecret validates a secret meets security requirements
func (sv *SecretValidator) ValidateSecret(name, value string) error {
	if len(value) < 16 {
		return fmt.Errorf("secret %s must be at least 16 characters long", name)
	}

	// Check for complexity (at least 3 of 4 character types)
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range value {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case (char >= '!' && char <= '/') || (char >= ':' && char <= '@') || (char >= '[' && char <= '`') || (char >= '{' && char <= '~'):
			hasSpecial = true
		}
	}

	complexityCount := 0
	if hasUpper {
		complexityCount++
	}
	if hasLower {
		complexityCount++
	}
	if hasDigit {
		complexityCount++
	}
	if hasSpecial {
		complexityCount++
	}

	if complexityCount < 3 {
		return fmt.Errorf("secret %s must contain at least 3 of: uppercase, lowercase, digits, special characters", name)
	}

	return nil
}

// AuditLog represents a secret access audit log
type AuditLog struct {
	SecretName string    `json:"secret_name"`
	Action     string    `json:"action"` // get, set, rotate, delete
	UserID     string    `json:"user_id,omitempty"`
	IPAddress  string    `json:"ip_address,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	Error      string    `json:"error,omitempty"`
}

// LogAuditEvent logs a secret access event
func (sm *SecretsManager) LogAuditEvent(ctx context.Context, event AuditLog) {
	sm.logger.Info("Secret audit event",
		zap.String("secret_name", event.SecretName),
		zap.String("action", event.Action),
		zap.String("user_id", event.UserID),
		zap.String("ip_address", event.IPAddress),
		zap.Bool("success", event.Success),
		zap.String("error", event.Error),
	)

	// In production, you would store this in a secure audit log
	// For now, just log it
}

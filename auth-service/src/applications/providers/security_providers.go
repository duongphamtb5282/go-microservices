package providers

import (
	"time"

	"auth-service/src/infrastructure/config"
	"backend-core/logging"
	"backend-core/security"
)

// JWTManagerProvider creates a JWT manager from backend-core/security
func JWTManagerProvider(cfg *config.Config, logger *logging.Logger) *security.JWTManager {
	// Parse token duration
	tokenDuration, err := time.ParseDuration(cfg.JWT.Expiry)
	if err != nil {
		logger.Warn("Invalid JWT expiry duration, using default 24h",
			logging.Error(err),
			logging.String("configured_expiry", cfg.JWT.Expiry))
		tokenDuration = 24 * time.Hour
	}

	logger.Info("Creating JWT manager",
		logging.String("issuer", cfg.JWT.Issuer),
		logging.String("audience", cfg.JWT.Audience),
		logging.Duration("token_duration", tokenDuration))

	return security.NewJWTManager(
		cfg.JWT.Secret,
		tokenDuration,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)
}

// AuthManagerProvider creates an authentication manager from backend-core/security
func AuthManagerProvider(jwtManager *security.JWTManager, cfg *config.Config, logger *logging.Logger) *security.AuthManager {
	bcryptCost := cfg.JWT.BcryptCost
	if bcryptCost == 0 {
		bcryptCost = 10 // Default bcrypt cost
	}

	logger.Info("Creating Auth manager",
		logging.Int("bcrypt_cost", bcryptCost))

	return security.NewAuthManager(jwtManager, bcryptCost)
}

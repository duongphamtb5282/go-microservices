package providers

import (
	"fmt"

	"auth-service/src/domain/authorization"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/identity/keycloak"
	"auth-service/src/interfaces/rest/middleware"
	"backend-core/cache/decorators"
	backendConfig "backend-core/config"
	"backend-core/logging"
	"backend-core/security"
)

// MiddlewareSetupProvider creates a middleware setup
func MiddlewareSetupProvider(cfg *config.Config, logger *logging.Logger) *middleware.MiddlewareSetup {
	return middleware.NewMiddlewareSetup(cfg, logger)
}

// UnifiedAuthorizationMiddlewareProvider creates a unified authorization middleware
func UnifiedAuthorizationMiddlewareProvider(
	cfg *config.Config,
	jwtManager *security.JWTManager,
	keycloakAdapter *keycloak.KeycloakAdapter,
	roleRepo authorization.RoleRepository,
	permissionRepo authorization.PermissionRepository,
	logger *logging.Logger,
) *middleware.UnifiedAuthorizationMiddleware {
	// Debug: Print actual config values
	fmt.Printf("DEBUG: Authorization config - Mode: '%s', Enabled: %v\n", cfg.Authorization.Mode, cfg.Authorization.Enabled)
	fmt.Printf("DEBUG: Comparison - Is JWT: %v, Is JWT_with_DB: %v, Is Keycloak: %v\n",
		cfg.Authorization.Mode == config.AuthorizationModeJWT,
		cfg.Authorization.Mode == config.AuthorizationModeJWTWithDB,
		cfg.Authorization.Mode == config.AuthorizationModeKeycloak)

	logger.Info("Creating unified authorization middleware",
		logging.String("mode", string(cfg.Authorization.Mode)),
		logging.Bool("enabled", cfg.Authorization.Enabled))

	// If mode is empty, log warning and set default
	if cfg.Authorization.Mode == "" {
		logger.Warn("Authorization mode is empty, using JWT with DB as default")
		cfg.Authorization.Mode = config.AuthorizationModeJWTWithDB
	}

	return middleware.NewUnifiedAuthorizationMiddleware(
		&cfg.Authorization,
		jwtManager,
		keycloakAdapter,
		roleRepo,
		permissionRepo,
		logger,
	)
}

// JWTAuthMiddlewareProvider creates a JWT authentication middleware wrapper
func JWTAuthMiddlewareProvider(
	jwtManager *security.JWTManager,
	logger *logging.Logger,
) *middleware.JWTMiddlewareWrapper {
	logger.Info("Creating JWT authentication middleware")
	return middleware.NewJWTMiddlewareWrapper(jwtManager, logger)
}

// KeycloakAuthorizationMiddlewareProvider creates a Keycloak authorization middleware
func KeycloakAuthorizationMiddlewareProvider(
	keycloakAdapter *keycloak.KeycloakAdapter,
	logger *logging.Logger,
) *middleware.KeycloakAuthorizationMiddleware {
	logger.Info("Creating Keycloak authorization middleware")
	return middleware.NewKeycloakAuthorizationMiddleware(keycloakAdapter, logger)
}

// CacheMiddlewareProvider creates a cache middleware
func CacheMiddlewareProvider(
	cfg *config.Config,
	logger *logging.Logger,
) *middleware.CacheMiddleware {
	// Create Redis config for cache decorator
	redisConfig := &backendConfig.RedisConfig{
		Name:         "auth-service-cache",
		Addr:         "localhost:6379", // TODO: Make configurable via config
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
	}

	// Create cache decorator with proper configuration
	cacheDecorator, err := decorators.NewCacheDecorator(*redisConfig, logger, nil)
	if err != nil {
		logger.Error("Failed to create cache decorator", logging.Error(err))
		// Return nil to indicate cache is disabled
		return nil
	}

	return middleware.NewCacheMiddleware(cacheDecorator, logger)
}

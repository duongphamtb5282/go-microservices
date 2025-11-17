package providers

import (
	"fmt"
	"time"

	"auth-service/src/applications/services"
	"auth-service/src/infrastructure/config"
	"auth-service/src/infrastructure/identity/keycloak"
	"auth-service/src/interfaces/rest/handlers"
	"backend-core/cache"
	"backend-core/logging"
	"backend-core/security"
)

// UserHandlerProvider creates a user handler
func UserHandlerProvider(
	userApplicationService *services.UserApplicationService,
	logger *logging.Logger,
) *handlers.UserHandler {
	return handlers.NewUserHandler(userApplicationService, logger)
}

// AuthHandlerProvider creates an auth handler
func AuthHandlerProvider(
	userApplicationService *services.UserApplicationService,
	jwtManager *security.JWTManager,
	logger *logging.Logger,
	identityProvider handlers.IdentityProvider,
	authConfig *config.AuthorizationConfig,
) *handlers.AuthHandler {
	return handlers.NewAuthHandler(userApplicationService, jwtManager, logger, identityProvider, authConfig)
}

// KeycloakAdapterProvider creates a Keycloak adapter using config
func KeycloakAdapterProvider(
	cfg *config.Config,
	cache cache.Cache,
	logger *logging.Logger,
) (*keycloak.KeycloakAdapter, error) {
	fmt.Printf("DEBUG KEYCLOAK CONFIG: BaseURL='%s', Realm='%s', ClientID='%s', ClientSecret='%s'\n",
		cfg.Keycloak.BaseURL, cfg.Keycloak.Realm, cfg.Keycloak.ClientID, cfg.Keycloak.ClientSecret)
	fmt.Printf("DEBUG: Keycloak config struct: %+v\n", cfg.Keycloak)

	// Parse timeout duration
	timeout, err := time.ParseDuration(cfg.Keycloak.Timeout)
	if err != nil {
		timeout = 30 * time.Second
	}

	// Parse cache TTL
	cacheTTL, err := time.ParseDuration(cfg.Keycloak.CacheTTL)
	if err != nil {
		cacheTTL = 5 * time.Minute
	}

	// Create Keycloak config from application config
	keycloakConfig := keycloak.KeycloakConfig{
		BaseURL:         cfg.Keycloak.BaseURL,
		Realm:           cfg.Keycloak.Realm,
		ClientID:        cfg.Keycloak.ClientID,
		ClientSecret:    cfg.Keycloak.ClientSecret,
		RedirectURI:     cfg.Keycloak.RedirectURI,
		Scopes:          cfg.Keycloak.Scopes,
		Timeout:         timeout,
		RetryAttempts:   cfg.Keycloak.RetryAttempts,
		CacheTTL:        cacheTTL,
		EnableSSO:       cfg.Keycloak.EnableSSO,
		EnableMFA:       cfg.Keycloak.EnableMFA,
		EnableRiskBased: cfg.Keycloak.EnableRiskBased,
	}

	fmt.Printf("DEBUG KEYCLOAK CONFIG: BaseURL='%s', Realm='%s', ClientID='%s'\n",
		keycloakConfig.BaseURL, keycloakConfig.Realm, keycloakConfig.ClientID)

	logger.Info("Creating Keycloak adapter",
		logging.String("base_url", keycloakConfig.BaseURL),
		logging.String("realm", keycloakConfig.Realm),
		logging.String("client_id", keycloakConfig.ClientID))

	// Create Keycloak client
	client, err := keycloak.NewKeycloakClient(keycloakConfig, logger)
	if err != nil {
		logger.Error("Failed to create Keycloak client", logging.Error(err))
		return nil, err
	}

	return keycloak.NewKeycloakAdapter(client, cache, keycloakConfig, logger), nil
}

// KeycloakApplicationServiceProvider creates a Keycloak application service
func KeycloakApplicationServiceProvider(
	keycloakAdapter *keycloak.KeycloakAdapter,
	logger *logging.Logger,
) *services.KeycloakApplicationService {
	return services.NewKeycloakApplicationService(keycloakAdapter, logger)
}

// KeycloakHandlerProvider creates a Keycloak handler
func KeycloakHandlerProvider(
	keycloakService *services.KeycloakApplicationService,
	logger *logging.Logger,
) *handlers.KeycloakHandler {
	return handlers.NewKeycloakHandler(keycloakService, logger)
}

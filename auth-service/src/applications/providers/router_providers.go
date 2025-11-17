package providers

import (
	"fmt"

	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"
	routerPkg "auth-service/src/interfaces/rest/router"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// RouterProvider creates a router with Keycloak authorization
func RouterProvider(
	userHandler *handlers.UserHandler,
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware,
	logger *logging.Logger,
) *routerPkg.Router {
	return routerPkg.NewRouter(userHandler, keycloakAuth, logger)
}

// RouteManagerProvider creates a route manager with full feature support
func RouteManagerProvider(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	cacheMiddleware *middleware.CacheMiddleware,
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware,
	logger *logging.Logger,
	telemetryMiddleware gin.HandlerFunc,
) *routerPkg.RouteManager {
	fmt.Printf("ROUTE_MANAGER_PROVIDER: Called with keycloakAuth=%v\n", keycloakAuth != nil)
	rm := routerPkg.NewRouteManager(authHandler, userHandler, cacheMiddleware, keycloakAuth, logger, telemetryMiddleware)
	fmt.Printf("ROUTE_MANAGER_PROVIDER: RouteManager created\n")
	return rm
}

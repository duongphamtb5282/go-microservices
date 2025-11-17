//go:build wireinject
// +build wireinject

package applications

import (
	"auth-service/src/applications/providers"
	"auth-service/src/infrastructure/config"
	"auth-service/src/interfaces/rest/middleware"
	"auth-service/src/interfaces/rest/router"
	"backend-core/database"
	"backend-core/logging"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// InitializeAuthService creates the complete auth service with proper dependency injection
func InitializeAuthService(
	cfg *config.Config,
	db database.Database,
	brokers []string,
	logger *logging.Logger,
) *gin.Engine {
	wire.Build(
		// Provider sets
		providers.UserRepositoryProvider,
		providers.UserCacheProvider,
		providers.UserDomainServiceProvider,
		providers.EventBusProvider,
		providers.UserApplicationServiceProvider,
		providers.UserHandlerProvider,
		providers.AuthHandlerProvider,
		providers.MiddlewareSetupProvider,
		providers.RouterProvider,

		// Router setup
		wire.Bind(new(*gin.Engine), new(*gin.Engine)),
		setupRouter,
	)
	return &gin.Engine{}
}

// setupRouter configures the router with middleware
func setupRouter(
	router *router.Router,
	middlewareSetup *middleware.MiddlewareSetup,
) *gin.Engine {
	ginRouter := router.SetupRoutes()
	middlewareSetup.SetupMiddleware(ginRouter)
	return ginRouter
}

package handlers

import (
	"auth-service/src/interfaces/rest/api"
	"auth-service/src/interfaces/rest/auth"
	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/role"
	"auth-service/src/interfaces/rest/system"
	"auth-service/src/interfaces/rest/user"
	"backend-core/logging"
)

var RouterGroupApp = new(RouterGroup)

type RouterGroup struct {
	Auth   auth.RouterGroup
	User   user.RouterGroup
	Role   role.RouterGroup
	System system.RouterGroup
}

// InitializeRouterGroups initializes all router groups with API handlers
func InitializeRouterGroups(
	authHandler *handlers.AuthHandler,
	tokenHandler *handlers.TokenHandler,
	userHandler *handlers.UserHandler,
	profileHandler *handlers.ProfileHandler,
	roleHandler *handlers.RoleHandler,
	permissionHandler *handlers.PermissionHandler,
	configHandler *handlers.ConfigHandler,
	logger *logging.Logger,
) *RouterGroup {
	// Initialize API groups
	_ = api.InitializeApiGroups(
		authHandler,
		tokenHandler,
		userHandler,
		profileHandler,
		roleHandler,
		permissionHandler,
		configHandler,
		logger,
	)

	// Initialize router groups
	authRouter := auth.RouterGroup{
		AuthRouter:  auth.AuthRouter{},
		TokenRouter: auth.TokenRouter{},
	}

	userRouter := user.RouterGroup{
		UserRouter:    user.UserRouter{},
		ProfileRouter: user.ProfileRouter{},
	}

	roleRouter := role.RouterGroup{
		RoleRouter:       role.RoleRouter{},
		PermissionRouter: role.PermissionRouter{},
	}

	systemRouter := system.RouterGroup{
		HealthRouter: system.HealthRouter{},
		ConfigRouter: system.ConfigRouter{},
	}

	return &RouterGroup{
		Auth:   authRouter,
		User:   userRouter,
		Role:   roleRouter,
		System: systemRouter,
	}
}

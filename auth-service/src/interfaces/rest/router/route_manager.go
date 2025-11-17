package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"auth-service/src/interfaces/rest/groups"
	"auth-service/src/interfaces/rest/handlers"
	"auth-service/src/interfaces/rest/middleware"

	"backend-core/logging"

	"github.com/gin-gonic/gin"
)

// TokenResponse represents the response from token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// exchangeCodeForTokens exchanges authorization code for tokens
func exchangeCodeForTokens(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:8085/callback")
	data.Set("client_id", "auth-service-client")
	data.Set("client_secret", "your-client-secret-change-in-production")

	req, err := http.NewRequest("POST", "http://localhost:8081/realms/auth-service/protocol/openid-connect/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// RouteManager manages all application routes
type RouteManager struct {
	authHandler         *handlers.AuthHandler
	userHandler         *handlers.UserHandler
	cacheMiddleware     *middleware.CacheMiddleware
	keycloakAuth        *middleware.KeycloakAuthorizationMiddleware
	logger              *logging.Logger
	telemetryMiddleware gin.HandlerFunc
}

// NewRouteManager creates a new route manager
func NewRouteManager(
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	cacheMiddleware *middleware.CacheMiddleware,
	keycloakAuth *middleware.KeycloakAuthorizationMiddleware,
	logger *logging.Logger,
	telemetryMiddleware gin.HandlerFunc,
) *RouteManager {
	return &RouteManager{
		authHandler:         authHandler,
		userHandler:         userHandler,
		cacheMiddleware:     cacheMiddleware,
		keycloakAuth:        keycloakAuth,
		logger:              logger,
		telemetryMiddleware: telemetryMiddleware,
	}
}

// SetupRoutes configures all application routes
func (rm *RouteManager) SetupRoutes() *gin.Engine {
	// Create Gin engine
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	// Add telemetry middleware if available
	if rm.telemetryMiddleware != nil {
		router.Use(rm.telemetryMiddleware)
		fmt.Printf("DEBUG: Telemetry middleware added in route manager\n")
	}

	// Initialize route groups
	authRoutes := groups.NewAuthRoutes(rm.authHandler)
	userRoutes := groups.NewUserRoutes(rm.userHandler, rm.cacheMiddleware)
	systemRoutes := groups.NewSystemRoutes(rm.cacheMiddleware)

	// Register system routes (health, cache, etc.) - includes swagger now
	systemRoutes.RegisterRoutes(router)

	// Register admin routes AFTER system routes with middleware
	fmt.Printf("DEBUG: Registering admin routes AFTER system routes with middleware\n")
	if rm.keycloakAuth != nil {
		router.GET("/admin/test", rm.keycloakAuth.RequireValidKeycloakToken(), func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /test endpoint handler called\n")
			c.JSON(200, gin.H{"message": "Admin test endpoint", "status": "success"})
		})
		router.GET("/admin/users", rm.keycloakAuth.RequireValidKeycloakToken(), func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /users endpoint handler called\n")
			c.JSON(200, gin.H{"message": "Admin users endpoint", "status": "success"})
		})
		fmt.Printf("DEBUG: Admin routes registered with Keycloak middleware\n")
	} else {
		router.GET("/admin/test", func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /test endpoint handler called (no auth)\n")
			c.JSON(200, gin.H{"message": "Admin test endpoint", "status": "success"})
		})
		router.GET("/admin/users", func(c *gin.Context) {
			fmt.Printf("DEBUG: Admin /users endpoint handler called (no auth)\n")
			c.JSON(200, gin.H{"message": "Admin users endpoint", "status": "success"})
		})
		fmt.Printf("DEBUG: Admin routes registered without middleware\n")
	}

	// Register SSO callback endpoint
	router.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		error_param := c.Query("error")

		if error_param != "" {
			c.JSON(400, gin.H{
				"error":             error_param,
				"error_description": c.Query("error_description"),
			})
			return
		}

		if code != "" {
			// Exchange authorization code for tokens
			tokenResponse, err := exchangeCodeForTokens(code)
			if err != nil {
				c.JSON(400, gin.H{
					"error":             "Token Exchange Failed",
					"error_description": err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"message":       "SSO login successful!",
				"access_token":  tokenResponse.AccessToken,
				"refresh_token": tokenResponse.RefreshToken,
				"id_token":      tokenResponse.IDToken,
				"token_type":    tokenResponse.TokenType,
				"expires_in":    tokenResponse.ExpiresIn,
			})
		} else {
			c.JSON(400, gin.H{
				"error":             "No Authorization Code",
				"error_description": "No authorization code received from Keycloak",
			})
		}
	})

	// API routes with versioning
	api := router.Group("/api/v1")
	{
		// Status endpoint with test middleware
		api.GET("/status", func(c *gin.Context) {
			// Test middleware
			fmt.Printf("DEBUG: Test middleware executed for /api/v1/status\n")
			c.Next()
		}, func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "auth-service",
				"version": "1.0.0",
				"status":  "running",
			})
		})

		// Register auth routes
		authRoutes.RegisterRoutes(api)

		// Register user routes
		userRoutes.RegisterRoutes(api)
	}

	rm.logger.Info("All routes registered successfully")
	return router
}

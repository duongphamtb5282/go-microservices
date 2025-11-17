package auth

import (
	"context"
	"strings"

	grpcerrors "backend-core/grpc/errors"
	grpcmiddleware "backend-core/grpc/middleware"
	"backend-core/logging"
	"backend-core/security"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthConfig holds configuration for authentication interceptor
type AuthConfig struct {
	// Enabled determines if authentication is enabled
	Enabled bool
	// JWTSecret is the secret key for JWT validation
	JWTSecret string
	// JWTIssuer is the expected issuer
	JWTIssuer string
	// JWTAudience is the expected audience
	JWTAudience string
	// ExemptMethods are methods that don't require authentication
	ExemptMethods []string
	// APIKeys are valid API keys for service-to-service authentication
	APIKeys map[string]string
}

// UnaryServerInterceptor returns a new unary server interceptor for authentication
func UnaryServerInterceptor(logger *logging.Logger, config *AuthConfig) grpc.UnaryServerInterceptor {
	exemptSet := make(map[string]bool)
	for _, method := range config.ExemptMethods {
		exemptSet[method] = true
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for exempt methods
		if exemptSet[info.FullMethod] {
			return handler(ctx, req)
		}

		// Check if authentication is enabled
		if !config.Enabled {
			return handler(ctx, req)
		}

		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("No metadata in request", logging.String("method", info.FullMethod))
			return nil, grpcerrors.NewUnauthenticatedError("missing metadata")
		}

		// Try API key authentication first
		if apiKeys := md.Get("x-api-key"); len(apiKeys) > 0 {
			apiKey := apiKeys[0]
			if serviceName, valid := validateAPIKey(config, apiKey); valid {
				ctx = grpcmiddleware.WithUserID(ctx, serviceName)
				ctx = grpcmiddleware.WithUsername(ctx, serviceName)
				logger.Debug("API key authentication successful",
					logging.String("service", serviceName),
					logging.String("method", info.FullMethod))
				return handler(ctx, req)
			}
		}

		// Try JWT authentication
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			logger.Warn("Missing authorization header", logging.String("method", info.FullMethod))
			return nil, grpcerrors.NewUnauthenticatedError("missing authorization header")
		}

		authHeader := authHeaders[0]
		token := extractBearerToken(authHeader)
		if token == "" {
			logger.Warn("Invalid authorization header format", logging.String("method", info.FullMethod))
			return nil, grpcerrors.NewUnauthenticatedError("invalid authorization header format")
		}

		// Validate JWT token
		jwtConfig := &security.JWTConfig{
			Secret:   config.JWTSecret,
			Issuer:   config.JWTIssuer,
			Audience: config.JWTAudience,
		}

		claims, err := security.ValidateJWT(token, jwtConfig)
		if err != nil {
			logger.Warn("JWT validation failed",
				logging.String("method", info.FullMethod),
				logging.Error(err))
			return nil, grpcerrors.NewUnauthenticatedError("invalid or expired token")
		}

		// Extract user information from claims
		userID := claims["user_id"].(string)
		username := ""
		if un, ok := claims["username"].(string); ok {
			username = un
		}

		// Extract roles if present
		var roles []string
		if r, ok := claims["roles"].([]interface{}); ok {
			for _, role := range r {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}

		// Add user information to context
		ctx = grpcmiddleware.WithUserID(ctx, userID)
		ctx = grpcmiddleware.WithUsername(ctx, username)
		ctx = grpcmiddleware.WithRoles(ctx, roles)

		logger.Debug("Authentication successful",
			logging.String("user_id", userID),
			logging.String("username", username),
			logging.String("method", info.FullMethod))

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new stream server interceptor for authentication
func StreamServerInterceptor(logger *logging.Logger, config *AuthConfig) grpc.StreamServerInterceptor {
	exemptSet := make(map[string]bool)
	for _, method := range config.ExemptMethods {
		exemptSet[method] = true
	}

	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for exempt methods
		if exemptSet[info.FullMethod] {
			return handler(srv, ss)
		}

		// Check if authentication is enabled
		if !config.Enabled {
			return handler(srv, ss)
		}

		ctx := ss.Context()

		// Extract metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("No metadata in stream", logging.String("method", info.FullMethod))
			return grpcerrors.NewUnauthenticatedError("missing metadata")
		}

		// Try API key authentication first
		if apiKeys := md.Get("x-api-key"); len(apiKeys) > 0 {
			apiKey := apiKeys[0]
			if serviceName, valid := validateAPIKey(config, apiKey); valid {
				logger.Debug("API key authentication successful for stream",
					logging.String("service", serviceName),
					logging.String("method", info.FullMethod))
				return handler(srv, ss)
			}
		}

		// Try JWT authentication
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			logger.Warn("Missing authorization header in stream", logging.String("method", info.FullMethod))
			return grpcerrors.NewUnauthenticatedError("missing authorization header")
		}

		authHeader := authHeaders[0]
		token := extractBearerToken(authHeader)
		if token == "" {
			logger.Warn("Invalid authorization header format in stream", logging.String("method", info.FullMethod))
			return grpcerrors.NewUnauthenticatedError("invalid authorization header format")
		}

		// Validate JWT token
		jwtConfig := &security.JWTConfig{
			Secret:   config.JWTSecret,
			Issuer:   config.JWTIssuer,
			Audience: config.JWTAudience,
		}

		_, err := security.ValidateJWT(token, jwtConfig)
		if err != nil {
			logger.Warn("JWT validation failed for stream",
				logging.String("method", info.FullMethod),
				logging.Error(err))
			return grpcerrors.NewUnauthenticatedError("invalid or expired token")
		}

		return handler(srv, ss)
	}
}

// extractBearerToken extracts the token from "Bearer <token>" format
func extractBearerToken(authHeader string) string {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// validateAPIKey validates an API key
func validateAPIKey(config *AuthConfig, apiKey string) (string, bool) {
	if config.APIKeys == nil {
		return "", false
	}

	for serviceName, validKey := range config.APIKeys {
		if apiKey == validKey {
			return serviceName, true
		}
	}

	return "", false
}

// UnaryClientInterceptor returns a new unary client interceptor that adds authentication
func UnaryClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Add authorization header to metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		// Add correlation ID if present in context
		if correlationID, ok := grpcmiddleware.GetCorrelationID(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-correlation-id", correlationID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a new stream client interceptor that adds authentication
func StreamClientInterceptor(token string) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// Add authorization header to metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)

		// Add correlation ID if present in context
		if correlationID, ok := grpcmiddleware.GetCorrelationID(ctx); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-correlation-id", correlationID)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

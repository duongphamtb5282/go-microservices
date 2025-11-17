package middleware

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Context keys for storing values in context
type contextKey string

const (
	// ContextKeyUserID stores the authenticated user ID
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyUsername stores the authenticated username
	ContextKeyUsername contextKey = "username"
	// ContextKeyRoles stores the user roles
	ContextKeyRoles contextKey = "roles"
	// ContextKeyCorrelationID stores the correlation/trace ID
	ContextKeyCorrelationID contextKey = "correlation_id"
	// ContextKeyClientIP stores the client IP address
	ContextKeyClientIP contextKey = "client_ip"
)

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(string)
	return userID, ok
}

// WithUsername adds username to context
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ContextKeyUsername, username)
}

// GetUsername retrieves username from context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(ContextKeyUsername).(string)
	return username, ok
}

// WithRoles adds roles to context
func WithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, ContextKeyRoles, roles)
}

// GetRoles retrieves roles from context
func GetRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(ContextKeyRoles).([]string)
	return roles, ok
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, ContextKeyCorrelationID, correlationID)
}

// GetCorrelationID retrieves correlation ID from context
func GetCorrelationID(ctx context.Context) (string, bool) {
	correlationID, ok := ctx.Value(ContextKeyCorrelationID).(string)
	return correlationID, ok
}

// WithClientIP adds client IP to context
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ContextKeyClientIP, ip)
}

// GetClientIP retrieves client IP from context
func GetClientIP(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(ContextKeyClientIP).(string)
	return ip, ok
}

// ExtractMetadata extracts metadata from incoming context
func ExtractMetadata(ctx context.Context) map[string]string {
	result := make(map[string]string)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		for k, v := range md {
			if len(v) > 0 {
				result[k] = v[0]
			}
		}
	}

	return result
}

// ExtractPeerInfo extracts peer information from context
func ExtractPeerInfo(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if ok {
		return p.Addr.String()
	}
	return ""
}

// InjectMetadata injects metadata into outgoing context
func InjectMetadata(ctx context.Context, key, value string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	md.Set(key, value)
	return metadata.NewOutgoingContext(ctx, md)
}

package middleware

import (
	"context"
	"strings"

	"google.golang.org/grpc"
)

// Matcher is a function that determines if an interceptor should be applied
type Matcher func(ctx context.Context, fullMethodName string) bool

// MatchFunc creates a selective unary server interceptor
func MatchFunc(matcher Matcher, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if matcher(ctx, info.FullMethod) {
			return interceptor(ctx, req, info, handler)
		}
		return handler(ctx, req)
	}
}

// MatchFuncStream creates a selective stream server interceptor
func MatchFuncStream(matcher Matcher, interceptor grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if matcher(ss.Context(), info.FullMethod) {
			return interceptor(srv, ss, info, handler)
		}
		return handler(srv, ss)
	}
}

// MatchAll returns a matcher that matches all methods
func MatchAll() Matcher {
	return func(ctx context.Context, fullMethodName string) bool {
		return true
	}
}

// MatchNone returns a matcher that matches no methods
func MatchNone() Matcher {
	return func(ctx context.Context, fullMethodName string) bool {
		return false
	}
}

// MatchMethods returns a matcher that matches specific method names
func MatchMethods(methods ...string) Matcher {
	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	return func(ctx context.Context, fullMethodName string) bool {
		return methodSet[fullMethodName]
	}
}

// MatchServices returns a matcher that matches specific service names
func MatchServices(services ...string) Matcher {
	return func(ctx context.Context, fullMethodName string) bool {
		for _, service := range services {
			if strings.HasPrefix(fullMethodName, "/"+service+"/") {
				return true
			}
		}
		return false
	}
}

// MatchPrefix returns a matcher that matches methods with specific prefix
func MatchPrefix(prefix string) Matcher {
	return func(ctx context.Context, fullMethodName string) bool {
		return strings.HasPrefix(fullMethodName, prefix)
	}
}

// ExceptMethods returns a matcher that matches all methods except specified ones
func ExceptMethods(methods ...string) Matcher {
	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	return func(ctx context.Context, fullMethodName string) bool {
		return !methodSet[fullMethodName]
	}
}

// ExceptServices returns a matcher that matches all methods except specified services
func ExceptServices(services ...string) Matcher {
	return func(ctx context.Context, fullMethodName string) bool {
		for _, service := range services {
			if strings.HasPrefix(fullMethodName, "/"+service+"/") {
				return false
			}
		}
		return true
	}
}

// AllButHealthZ is a common matcher that matches all except health check
func AllButHealthZ() Matcher {
	return ExceptMethods(
		"/grpc.health.v1.Health/Check",
		"/grpc.health.v1.Health/Watch",
	)
}

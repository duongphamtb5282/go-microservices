# Backend-Core gRPC Middleware

A comprehensive gRPC middleware package for Go microservices, providing interceptors for authentication, logging, tracing, metrics, rate limiting, validation, and recovery.

## 📦 Features

- **🔐 Authentication**: JWT and API key authentication
- **📝 Logging**: Structured logging with correlation IDs
- **🔍 Tracing**: OpenTelemetry distributed tracing
- **📊 Metrics**: Custom metrics reporting interface
- **⚡ Rate Limiting**: Redis-based rate limiting
- **✅ Validation**: Proto message validation
- **🛡️ Recovery**: Panic recovery with stack traces
- **🔄 Retry**: Client-side retry with exponential backoff
- **⏱️ Timeout**: Request timeout handling
- **🏗️ Builder Pattern**: Fluent API for server and client creation

## 📁 Package Structure

```
backend-core/grpc/
├── middleware/
│   ├── context.go           # Context helpers for storing user info, correlation IDs
│   └── selector.go          # Selective interceptor application
├── interceptors/
│   ├── auth/                # Authentication interceptor (JWT, API keys)
│   ├── logging/             # Logging interceptor with structured logs
│   ├── recovery/            # Panic recovery interceptor
│   ├── validation/          # Request validation interceptor
│   ├── ratelimit/           # Rate limiting interceptor (Redis-backed)
│   ├── metrics/             # Metrics reporting interceptor
│   └── tracing/             # Distributed tracing interceptor
├── server/
│   ├── server_config.go     # Server configuration
│   └── server_builder.go    # Fluent API for building gRPC servers
├── client/
│   ├── client_config.go     # Client configuration
│   └── client_builder.go    # Fluent API for building gRPC clients
├── errors/
│   └── errors.go            # gRPC error mapping utilities
├── health/
│   └── health.go            # Health check implementation
└── README.md                # This file
```

## 🚀 Quick Start

### Server with Middleware

```go
package main

import (
    "backend-core/logging"
    grpcauth "backend-core/grpc/interceptors/auth"
    grpclogging "backend-core/grpc/interceptors/logging"
    grpcserver "backend-core/grpc/server"
    "google.golang.org/grpc"
)

func main() {
    // Initialize logger
    logger, _ := logging.NewLogger(&logging.LoggingConfig{
        Level:  "info",
        Format: "json",
    })

    // Configure server
    serverConfig := &grpcserver.ServerConfig{
        Host:           "0.0.0.0",
        Port:           "50051",
        MaxMessageSize: 10 * 1024 * 1024, // 10MB
    }

    // Configure authentication
    authConfig := &grpcauth.AuthConfig{
        Enabled:       true,
        JWTSecret:     "your-secret-key",
        JWTIssuer:     "microservices",
        JWTAudience:   "clients",
        ExemptMethods: []string{
            "/grpc.health.v1.Health/Check",
        },
    }

    // Configure logging
    loggingConfig := &grpclogging.LoggingConfig{
        LogPayload:        false,
        LogPayloadOnError: true,
    }

    // Build gRPC server with middleware
    grpcServer := grpcserver.NewServerBuilder(serverConfig).
        WithRecovery(logger).                  // 1. Catch panics
        WithLogging(logger, loggingConfig).    // 2. Log requests
        WithTracing().                         // 3. Add tracing
        WithAuth(logger, authConfig).          // 4. Authenticate
        WithValidation().                      // 5. Validate input
        Build()

    // Register your services
    // pb.RegisterYourServiceServer(grpcServer, yourService)

    // Start server
    lis, _ := net.Listen("tcp", ":50051")
    grpcServer.Serve(lis)
}
```

### Client with Middleware

```go
package main

import (
    "backend-core/logging"
    grpcclient "backend-core/grpc/client"
    pb "your-service/proto"
)

func main() {
    // Initialize logger
    logger, _ := logging.NewLogger(&logging.LoggingConfig{
        Level: "info",
    })

    // Configure client
    clientConfig := &grpcclient.ClientConfig{
        Address:        "localhost:50051",
        Timeout:        30 * time.Second,
        MaxMessageSize: 10 * 1024 * 1024,
        Insecure:       true,
    }

    // Build gRPC client with middleware
    conn, err := grpcclient.NewClientBuilder(clientConfig).
        WithRetry(&grpcclient.RetryConfig{
            MaxAttempts:       3,
            BackoffMultiplier: 2.0,
            InitialBackoff:    100 * time.Millisecond,
            MaxBackoff:        5 * time.Second,
        }).
        WithLogging(logger).
        WithTracing().
        Build()

    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Create client
    client := pb.NewYourServiceClient(conn)

    // Make requests
    // resp, err := client.YourMethod(ctx, req)
}
```

## 🔧 Interceptors

### 1. Authentication Interceptor

Validates JWT tokens or API keys on incoming requests.

```go
import grpcauth "backend-core/grpc/interceptors/auth"

authConfig := &grpcauth.AuthConfig{
    Enabled:     true,
    JWTSecret:   "your-secret",
    JWTIssuer:   "microservices",
    JWTAudience: "clients",
    ExemptMethods: []string{
        "/grpc.health.v1.Health/Check",
    },
    APIKeys: map[string]string{
        "auth-service": "api-key-123",
    },
}

builder.WithAuth(logger, authConfig)
```

**Features:**

- JWT token validation
- API key authentication for service-to-service
- Exempt specific methods (health checks, reflection)
- Extract user info (user_id, username, roles) into context

**Context Values:**

```go
import grpcmiddleware "backend-core/grpc/middleware"

// In your handler
userID, ok := grpcmiddleware.GetUserID(ctx)
username, ok := grpcmiddleware.GetUsername(ctx)
roles, ok := grpcmiddleware.GetRoles(ctx)
```

### 2. Logging Interceptor

Structured logging for all gRPC calls.

```go
import grpclogging "backend-core/grpc/interceptors/logging"

loggingConfig := &grpclogging.LoggingConfig{
    LogPayload:        false, // Don't log payloads by default
    LogPayloadOnError: true,  // Log payloads only on errors
}

builder.WithLogging(logger, loggingConfig)
```

**Log Output:**

```json
{
  "level": "info",
  "timestamp": "2025-10-13T10:15:30Z",
  "grpc_method": "/admin.AdminService/RecordUserCreated",
  "grpc_code": "OK",
  "duration": "45ms",
  "correlation_id": "uuid-xxx",
  "user_id": "user-123",
  "peer": "192.168.1.100:45678"
}
```

### 3. Recovery Interceptor

Catches panics and converts them to gRPC errors.

```go
import "backend-core/grpc/interceptors/recovery"

builder.WithRecovery(logger)
```

**Features:**

- Catches panics in handlers
- Logs panic with stack trace
- Returns `INTERNAL` error to client
- Prevents service crash

### 4. Validation Interceptor

Validates proto message fields.

```go
builder.WithValidation()
```

**Usage in Proto Messages:**

```go
// Implement Validator interface
func (r *YourRequest) Validate() error {
    v := validation.NewFieldValidator()
    return v.
        Required("user_id", r.UserId).
        Email("email", r.Email).
        MinLength("username", r.Username, 3).
        Error()
}
```

### 5. Rate Limiting Interceptor

Redis-based distributed rate limiting.

```go
import (
    "backend-core/cache"
    "backend-core/grpc/interceptors/ratelimit"
)

rateLimitConfig := &ratelimit.RateLimitConfig{
    Enabled:           true,
    RequestsPerMinute: 100,
    Burst:             20,
    MethodLimits: map[string]int{
        "/admin.AdminService/RecordUserCreated": 50,
    },
    KeyPrefix: "ratelimit:",
}

builder.WithRateLimit(redisCache, rateLimitConfig, logger)
```

**Features:**

- Rate limit by user ID
- Rate limit by IP address
- Per-method rate limits
- Redis-backed for distributed systems
- Fail-open on Redis errors

### 6. Metrics Interceptor

Custom metrics reporting interface.

```go
import "backend-core/grpc/interceptors/metrics"

// Implement MetricsReporter interface
type MyMetricsReporter struct {}

func (r *MyMetricsReporter) RecordRequest(method string, code string, duration time.Duration) {}
func (r *MyMetricsReporter) RecordMessageSent(method string) {}
func (r *MyMetricsReporter) RecordMessageReceived(method string) {}

builder.WithMetrics(&MyMetricsReporter{})
```

### 7. Tracing Interceptor

Simplified tracing with correlation IDs (use `otelgrpc` for full OpenTelemetry).

```go
builder.WithTracing()
```

**Features:**

- Propagates correlation IDs via metadata
- Integrates with OpenTelemetry when available
- Adds trace context to logs

## 🛠️ Context Helpers

Store and retrieve values from context:

```go
import grpcmiddleware "backend-core/grpc/middleware"

// Store values
ctx = grpcmiddleware.WithUserID(ctx, "user-123")
ctx = grpcmiddleware.WithUsername(ctx, "john.doe")
ctx = grpcmiddleware.WithRoles(ctx, []string{"admin", "user"})
ctx = grpcmiddleware.WithCorrelationID(ctx, "uuid-xxx")

// Retrieve values
userID, ok := grpcmiddleware.GetUserID(ctx)
username, ok := grpcmiddleware.GetUsername(ctx)
roles, ok := grpcmiddleware.GetRoles(ctx)
correlationID, ok := grpcmiddleware.GetCorrelationID(ctx)
```

## 🎯 Selective Interceptors

Apply interceptors only to specific methods:

```go
import "backend-core/grpc/middleware"

// Match specific methods
matcher := middleware.MatchMethods(
    "/admin.AdminService/RecordUserCreated",
    "/admin.AdminService/RecordUserUpdated",
)

// Except health checks
matcher := middleware.AllButHealthZ()

// Except specific services
matcher := middleware.ExceptServices("grpc.health.v1.Health")

// Use with interceptor
builder.WithUnaryInterceptor(
    middleware.MatchFunc(matcher, yourInterceptor),
)
```

## 🏥 Health Checks

Implement gRPC health checking protocol:

```go
import (
    "backend-core/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

// Create health server
healthServer := health.NewHealthServer()
grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

// Set serving status
healthServer.SetServingStatus("admin.AdminService",
    grpc_health_v1.HealthCheckResponse_SERVING)

// Shutdown gracefully
healthServer.Shutdown()
```

## 📊 Error Handling

Map errors to gRPC status codes:

```go
import grpcerrors "backend-core/grpc/errors"

// Predefined errors
err := grpcerrors.NewUnauthenticatedError("invalid token")
err := grpcerrors.NewUnauthorizedError("insufficient permissions")
err := grpcerrors.NewInvalidArgumentError("user_id is required")
err := grpcerrors.NewNotFoundError("user not found")
err := grpcerrors.NewRateLimitError("rate limit exceeded")

// Map any error
return grpcerrors.MapError(err)
```

## 🔗 Integration Examples

### Admin Service (Server)

```go
// See: admin-service/cmd/server/main.go

grpcServer := grpcserver.NewServerBuilder(serverConfig).
    WithRecovery(logger).
    WithLogging(logger, loggingConfig).
    WithTracing().
    WithAuth(logger, authConfig).
    WithValidation().
    Build()
```

### Auth Service (Client)

```go
// See: auth-service/src/infrastructure/grpc/admin_client.go

conn, err := grpcclient.NewClientBuilder(clientConfig).
    WithRetry(retryConfig).
    WithLogging(logger).
    WithTracing().
    Build()
```

## ⚙️ Configuration

### Environment Variables

```bash
# gRPC Server
GRPC_PORT=50051
GRPC_HOST=0.0.0.0

# Authentication
GRPC_AUTH_ENABLED=true
JWT_SECRET=your-secret-key
JWT_ISSUER=microservices
GRPC_API_KEY=service-api-key

# Rate Limiting
REDIS_URL=redis://localhost:6379
```

### YAML Configuration

```yaml
grpc:
  host: "0.0.0.0"
  port: "50051"
  auth:
    enabled: true
    jwt_secret: "${JWT_SECRET}"
    jwt_issuer: "microservices"
    api_key: "${GRPC_API_KEY}"
```

## 🧪 Testing

### Unit Tests

```go
import (
    "testing"
    "backend-core/grpc/interceptors/auth"
)

func TestAuthInterceptor(t *testing.T) {
    // Test authentication logic
}
```

### Integration Tests

```go
import (
    "testing"
    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"
)

func TestGRPCServerWithMiddleware(t *testing.T) {
    // Create in-memory gRPC server
    // Test with middleware enabled
}
```

## 📈 Performance Considerations

- **Interceptor Order**: Recovery → Logging → Tracing → Auth → Validation → Rate Limit
- **Overhead**: Each interceptor adds ~0.1-1ms latency
- **Rate Limiting**: Fail-open on Redis errors to prevent cascading failures
- **Connection Pooling**: Reuse client connections

## 🔒 Security Best Practices

1. **Enable TLS in production**: Set `Insecure: false` and provide certificates
2. **Rotate JWT secrets**: Use environment variables for secrets
3. **Use API keys for service-to-service**: Avoid JWT overhead for internal calls
4. **Validate all inputs**: Implement `Validate()` on proto messages
5. **Set rate limits**: Protect against abuse and DDoS

## 📚 References

- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [go-grpc-middleware](https://github.com/grpc-ecosystem/go-grpc-middleware)
- [OpenTelemetry gRPC](https://opentelemetry.io/docs/instrumentation/go/)
- [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)

## 🤝 Contributing

1. Follow the existing package structure
2. Add tests for new interceptors
3. Update documentation
4. Ensure backward compatibility

## 📝 License

Apache 2.0 License - See LICENSE file for details

---

**Built with ❤️ for Go microservices**

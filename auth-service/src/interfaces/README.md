# Interface Layer - DDD Architecture

This directory contains the Interface Layer components following Domain-Driven Design principles.

## 📁 Directory Structure

```
src/interfaces/
├── rest/                    # HTTP REST API
│   ├── handlers/           # HTTP request handlers
│   ├── middleware/         # HTTP middleware (moved from internal/middleware)
│   ├── protocol/           # HTTP protocol definitions (moved from internal/protocol)
│   │   └── http/
│   │       ├── dto/        # Data Transfer Objects
│   │       │   ├── request/
│   │       │   └── response/
│   │       ├── middleware/  # Protocol-specific middleware
│   │       └── validation/  # Request/response validation
│   └── router/             # Route configuration
├── grpc/                   # gRPC API (future)
│   └── proto/             # Protocol buffer definitions
├── kafka/                  # Kafka consumers/producers
└── websocket/             # WebSocket API (future)
    └── handlers/          # WebSocket handlers
```

## 🎯 Responsibilities

### REST Interface (`src/interfaces/rest/`)

- **Handlers**: HTTP request/response handling
- **Middleware**: Cross-cutting concerns (auth, logging, CORS, etc.)
- **Protocol**: HTTP-specific DTOs, validation, API contracts
- **Router**: Route configuration and mapping

### gRPC Interface (`src/interfaces/grpc/`)

- **Proto**: Protocol buffer definitions
- **Handlers**: gRPC service implementations
- **Middleware**: gRPC-specific middleware

### WebSocket Interface (`src/interfaces/websocket/`)

- **Handlers**: WebSocket connection handling
- **Protocol**: WebSocket-specific message formats

## 🔄 Migration from Legacy Structure

### From `internal/middleware/` → `src/interfaces/rest/middleware/`

- `cache_middleware.go` - Cache middleware
- `casbin_handler.go` - Authorization middleware
- `exception_middleware.go` - Error handling middleware
- `jwt_auth.go` - JWT authentication middleware
- `logging_middleware.go` - Request logging middleware
- `operation_record.go` - Audit logging middleware
- `validation_middleware.go` - Request validation middleware

### From `internal/protocol/` → `src/interfaces/rest/protocol/`

- `http/dto/` - HTTP DTOs and data contracts
- `http/api/` - API endpoint definitions
- `http/handlers/` - HTTP handlers
- `http/router/` - Route configuration
- `http/middleware_setup.go` - Middleware configuration

## 🏗️ DDD Principles Applied

### 1. Interface Layer Responsibility

- **External Communication**: Handle HTTP, gRPC, WebSocket protocols
- **Request/Response Processing**: Protocol-specific data transformation
- **Cross-cutting Concerns**: Middleware for common functionality

### 2. Clean Boundaries

- **No Business Logic**: Interface layer only handles communication
- **Dependency Direction**: Interfaces depend on Application layer
- **Protocol Agnostic**: Easy to add new communication protocols

### 3. Separation of Concerns

- **Middleware**: Infrastructure concerns (auth, logging, CORS)
- **Protocol**: Communication contracts and data transfer
- **Handlers**: Request orchestration (delegate to Application layer)
- **Router**: Route configuration and mapping

## 🚀 Benefits

1. **Clear Separation**: Interface concerns separated from business logic
2. **Protocol Agnostic**: Easy to add gRPC, WebSocket, etc.
3. **Testable**: Interface layer can be tested independently
4. **Maintainable**: Changes to protocols don't affect business logic
5. **Scalable**: Easy to add new communication protocols

## 📝 Usage Examples

### Adding New Middleware

```go
// src/interfaces/rest/middleware/rate_limit_middleware.go
package middleware

func RateLimitMiddleware() gin.HandlerFunc {
    // Rate limiting logic
}
```

### Adding New Protocol

```go
// src/interfaces/grpc/handlers/user_grpc_handler.go
package handlers

func (h *UserGRPCHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    // gRPC handler logic
}
```

### Adding New DTO

```go
// src/interfaces/rest/protocol/http/dto/request/create_user_request.go
package request

type CreateUserRequest struct {
    Username string `json:"username" validate:"required"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

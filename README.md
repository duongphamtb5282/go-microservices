# Golang Keycloak Microservices Platform

A comprehensive, production-ready microservices platform built with Go, featuring Keycloak authentication, Kafka messaging, and full observability stack.

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │    │  Admin Service   │    │ GraphQL Service │
│   (Gin + JWT)   │◄──►│   (gRPC)        │◄──►│   (GraphQL)     │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                      │                       │
         │                      │                       │
         ▼                      ▼                       ▼
┌─────────────────────────────────────────────────────────────┐
│              Backend Core (Shared Library)                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │Database │  │  Cache   │  │ Messaging│  │Security  │  │
│  │ (GORM)  │  │ (Redis)  │  │ (Kafka)  │  │(Keycloak)│  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                      │                       │
         ▼                      ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  PostgreSQL     │    │     Redis       │    │     Kafka       │
│   Database      │    │     Cache       │    │   Messaging     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🚀 Features

### Core Services
- **Auth Service**: JWT-based authentication with Keycloak integration
- **Admin Service**: Administrative operations via gRPC
- **GraphQL Service**: Flexible GraphQL API with MongoDB
- **Notification Service**: Event-driven notification system

### Infrastructure
- **Backend Core**: Shared library with common functionality
  - Database abstraction (GORM with PostgreSQL)
  - Caching layer (Redis with multiple strategies)
  - Messaging (Kafka with Confluent client)
  - Security (JWT, Keycloak integration)
  - Logging (Structured logging with masking)
  - Telemetry (OpenTelemetry integration)

### Observability
- **Metrics**: Prometheus metrics collection
- **Tracing**: Distributed tracing with Jaeger
- **Logging**: Centralized logging with Loki
- **Monitoring**: Grafana dashboards
- **Alerting**: AlertManager integration

### Messaging
- **Kafka**: Event-driven architecture with Confluent Kafka Go client
- **DLQ**: Dead Letter Queue for failed messages
- **Retry**: Exponential backoff retry mechanism
- **Consumer Groups**: Scalable consumer group support

## 📋 Prerequisites

- **Go**: 1.24+ (with CGO enabled for Kafka)
- **Docker**: 20.10+ and Docker Compose 2.0+
- **PostgreSQL**: 14+ (or use Docker)
- **Redis**: 6.0+ (or use Docker)
- **Kafka**: 2.8+ (or use Docker)
- **Keycloak**: 20+ (or use Docker)
- **librdkafka**: Required for Kafka client (installed via Docker or system package)

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.24
- **Web Framework**: Gin 1.11
- **gRPC**: Google gRPC
- **GraphQL**: gqlgen
- **ORM**: GORM 1.25
- **Database**: PostgreSQL 14+
- **Cache**: Redis 6+
- **Messaging**: Confluent Kafka Go v2
- **Authentication**: Keycloak, JWT
- **Configuration**: Viper
- **Logging**: Zap (Uber)
- **Telemetry**: OpenTelemetry

### Infrastructure
- **Containerization**: Docker, Docker Compose
- **Orchestration**: Kubernetes (optional)
- **Monitoring**: Prometheus, Grafana
- **Tracing**: Jaeger
- **Logging**: Loki, Promtail
- **Alerting**: AlertManager

## 📁 Project Structure

```
golang_keycloak/
├── auth-service/          # Authentication service
│   ├── cmd/               # Application entry points
│   ├── src/               # Source code
│   │   ├── applications/  # Application layer
│   │   ├── domain/        # Domain layer
│   │   ├── infrastructure/# Infrastructure layer
│   │   └── interfaces/    # Interface layer
│   ├── config/            # Configuration files
│   ├── migrations/        # Database migrations
│   └── scripts/           # Utility scripts
│
├── admin-service/         # Admin service (gRPC)
│   ├── cmd/
│   ├── src/
│   └── migrations/
│
├── graphql-service/        # GraphQL service
│   ├── cmd/
│   ├── internal/
│   └── migrations/
│
├── notification-service/   # Notification service
│   ├── cmd/
│   └── internal/
│
├── backend-core/           # Shared core library
│   ├── cache/             # Caching abstractions
│   ├── database/          # Database abstractions
│   ├── messaging/         # Messaging (Kafka)
│   ├── security/          # Security utilities
│   ├── logging/           # Logging utilities
│   ├── telemetry/         # Observability
│   ├── middleware/        # HTTP middleware
│   └── config/            # Configuration
│
├── backend-shared/         # Shared models and events
│   ├── models/            # Shared data models
│   ├── events/            # Event definitions
│   └── proto/             # Protocol buffers
│
├── infrastructure/         # Infrastructure as code
│   ├── prometheus/        # Prometheus config
│   ├── grafana/           # Grafana dashboards
│   ├── otel-collector-config.yml
│   └── docker-compose.yml
│
└── guides/                 # Documentation
    ├── auth-service/      # Auth service guides
    ├── backend-core/      # Backend core guides
    └── ...                # Other guides
```

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd golang_keycloak
```

### 2. Start Infrastructure Services

```bash
# Start all infrastructure services (PostgreSQL, Redis, Kafka, Keycloak, etc.)
docker-compose -f infrastructure/docker-compose.yml up -d

# Or start individual services
docker-compose -f auth-service/docker-compose.dev.yml up -d
```

### 3. Configure Environment Variables

```bash
# Copy example environment files
cp auth-service/env.example auth-service/.env
cp admin-service/env.example admin-service/.env

# Edit .env files with your configuration
```

### 4. Run Database Migrations

```bash
# Auth service migrations
cd auth-service
go run cmd/migrate/main.go up

# Admin service migrations
cd ../admin-service
go run cmd/migrate/main.go up
```

### 5. Start Services

```bash
# Start auth service
cd auth-service
go run cmd/http/main.go

# Start admin service (in another terminal)
cd admin-service
go run cmd/server/main.go

# Start GraphQL service (in another terminal)
cd graphql-service
go run cmd/server/main.go
```

### 6. Access Services

- **Auth Service**: http://localhost:8085
- **Admin Service**: http://localhost:8086
- **GraphQL Service**: http://localhost:8087
- **Keycloak**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000
- **Jaeger**: http://localhost:16686

## 🔧 Development

### Building Services

```bash
# Build all services
make build

# Build specific service
cd auth-service && go build ./cmd/http
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./backend-core/messaging/kafka/...
```

### Code Generation

```bash
# Generate gRPC code
cd backend-shared/proto
./generate.sh

# Generate GraphQL code
cd graphql-service
go generate ./...
```

### Hot Reload (Development)

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
cd auth-service
air
```

## 📊 Monitoring & Observability

### Metrics

Services expose Prometheus metrics at `/metrics` endpoint:

```bash
# View metrics
curl http://localhost:8085/metrics
```

### Tracing

Distributed tracing is configured via OpenTelemetry. Traces are sent to Jaeger:

- **Jaeger UI**: http://localhost:16686

### Logging

Structured logging with automatic PII masking:

```go
logger.Info("User logged in",
    logging.String("user_id", userID),
    logging.String("email", email), // Automatically masked
)
```

### Dashboards

Pre-configured Grafana dashboards are available:
- Service metrics
- Kafka consumer lag
- Database performance
- Error rates

## 🔐 Security

### Authentication Flow

1. User authenticates via Keycloak
2. Auth service validates and issues JWT tokens
3. Services validate JWT tokens via middleware
4. RBAC permissions checked via Keycloak

### Environment Variables

Never commit sensitive data. Use environment variables:

```bash
# Required for production
KEYCLOAK_URL=https://keycloak.example.com
KEYCLOAK_CLIENT_ID=your-client-id
KEYCLOAK_CLIENT_SECRET=your-secret

DATABASE_URL=postgres://user:pass@host:5432/db
REDIS_URL=redis://host:6379
KAFKA_BROKERS=broker1:9092,broker2:9092
```

## 📦 Deployment

### Docker

```bash
# Build images
docker build -t auth-service:latest ./auth-service
docker build -t admin-service:latest ./admin-service

# Run with docker-compose
docker-compose up -d
```

### Kubernetes

```bash
# Apply Kubernetes manifests
kubectl apply -f auth-service/k8s/
kubectl apply -f infrastructure/helm/
```

### Production Considerations

1. **Environment Variables**: Use secrets management (Vault, AWS Secrets Manager)
2. **Database**: Use managed PostgreSQL (RDS, Cloud SQL)
3. **Cache**: Use managed Redis (ElastiCache, Memorystore)
4. **Kafka**: Use managed Kafka (Confluent Cloud, MSK)
5. **Monitoring**: Configure alerting rules
6. **Scaling**: Use horizontal pod autoscaling
7. **SSL/TLS**: Configure certificates for all services

## 🧪 Testing

### Unit Tests

```bash
go test -v ./...
```

### Integration Tests

```bash
# Start test infrastructure
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### Load Testing

```bash
# Using k6
cd auth-service/k6
k6 run full-load-test.js
```

## 📚 Documentation

- **API Documentation**: Swagger UI available at `/swagger/index.html`
- **Architecture Guides**: See `guides/` directory
- **Service-Specific Docs**: Each service has its own README.md

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions: `gofmt`, `golint`
- Write tests for new features
- Update documentation
- Ensure all tests pass

## 📝 License

[Specify your license here]

## 🙏 Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [Confluent Kafka Go](https://github.com/confluentinc/confluent-kafka-go)
- [OpenTelemetry](https://opentelemetry.io/)
- [Keycloak](https://www.keycloak.org/)

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Check the guides in `guides/` directory
- Review service-specific README files

## 🔄 Recent Updates

### Kafka Migration (Latest)
- Migrated from Sarama to Confluent Kafka Go v2
- Improved performance and feature support
- Better Confluent Cloud integration

### Observability
- Full OpenTelemetry integration
- Prometheus metrics collection
- Distributed tracing with Jaeger
- Centralized logging with Loki

---

**Built with ❤️ using Go**


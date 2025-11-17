# Communication Testing Scripts

This directory contains comprehensive testing scripts for Kafka and gRPC communication in the auth service ecosystem.

## Overview

The auth service communicates with other microservices through:
- **Kafka**: Event-driven communication for user events, auth events, and audit logs
- **gRPC**: Direct service-to-service communication with the admin service

## Test Scripts

### 1. `test-kafka-integration.sh`
Tests Kafka communication between auth-service and other components.

**Features:**
- Kafka broker connectivity checks
- Topic management verification
- Producer and consumer testing
- Auth service integration testing (user creation triggers Kafka events)
- Performance testing
- Error handling validation

**Usage:**
```bash
# Run all Kafka tests
./test-kafka-integration.sh

# With custom configuration
KAFKA_BROKER=kafka.example.com:9092 AUTH_SERVICE_URL=http://localhost:8085 ./test-kafka-integration.sh
```

### 2. `test-grpc-integration.sh`
Tests gRPC communication between auth-service and admin-service.

**Features:**
- gRPC service discovery
- Health checks
- User CRUD operations via gRPC
- Auth service integration testing
- Performance testing
- Error handling validation

**Prerequisites:**
- Install `grpcurl` for advanced testing:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

**Usage:**
```bash
# Run all gRPC tests
./test-grpc-integration.sh

# With custom configuration
ADMIN_SERVICE_GRPC=admin.example.com:50051 ./test-grpc-integration.sh
```

### 3. `test-all-communications.sh`
Comprehensive test runner that manages the entire testing lifecycle.

**Features:**
- Automated infrastructure startup
- Service health monitoring
- Sequential test execution
- Cleanup and teardown

**Commands:**
```bash
# Show help
./test-all-communications.sh help

# Start all services (infrastructure + applications)
./test-all-communications.sh start

# Check service status
./test-all-communications.sh status

# Run all communication tests
./test-all-communications.sh test

# Run only Kafka tests
./test-all-communications.sh kafka

# Run only gRPC tests
./test-all-communications.sh grpc

# Run end-to-end tests
./test-all-communications.sh e2e

# Stop all services and cleanup
./test-all-communications.sh stop
```

## Quick Start

### Automated Testing (Recommended)

```bash
# Navigate to scripts directory
cd auth-service/scripts

# Start all services and run tests
./test-all-communications.sh start
./test-all-communications.sh test
```

### Manual Testing

If you prefer to manage services manually:

```bash
# Start infrastructure
docker-compose -f ../../infrastructure/docker-compose.yml up -d

# Start admin service
cd ../../admin-service
docker-compose up -d

# Start auth service
cd ../auth-service
docker-compose -f docker-compose.app.yml up -d

# Run individual tests
./scripts/test-kafka-integration.sh
./scripts/test-grpc-integration.sh
```

## Environment Variables

### Kafka Testing
- `KAFKA_BROKER`: Kafka broker address (default: `localhost:9092`)
- `AUTH_SERVICE_URL`: Auth service URL (default: `http://localhost:8085`)
- `KAFKA_TOPIC_USER_EVENTS`: User events topic (default: `user.events`)
- `KAFKA_TOPIC_AUTH_EVENTS`: Auth events topic (default: `auth.events`)

### gRPC Testing
- `ADMIN_SERVICE_GRPC`: Admin service gRPC address (default: `localhost:50051`)
- `ADMIN_SERVICE_HTTP`: Admin service HTTP address (default: `http://localhost:8086`)
- `AUTH_SERVICE_URL`: Auth service URL (default: `http://localhost:8085`)

## Test Coverage

### Kafka Tests
- ✅ Broker connectivity
- ✅ Topic existence and management
- ✅ Message production
- ✅ Message consumption
- ✅ Auth service integration (user creation → Kafka event)
- ✅ Authentication events (login/logout)
- ✅ Performance testing
- ✅ Error handling

### gRPC Tests
- ✅ Service discovery
- ✅ Health checks
- ✅ User creation via gRPC
- ✅ User updates via gRPC
- ✅ User deletion via gRPC
- ✅ Auth service integration
- ✅ Performance testing
- ✅ Error handling

## Monitoring and Debugging

### Kafka Monitoring
```bash
# List topics
kafka-topics.sh --bootstrap-server localhost:9092 --list

# Consume messages
kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic user.events --from-beginning

# Check consumer groups
kafka-consumer-groups.sh --bootstrap-server localhost:9092 --list

# Kafka UI (if running)
open http://localhost:8080
```

### gRPC Monitoring
```bash
# List services
grpcurl -plaintext localhost:50051 list

# Health check
grpcurl -plaintext -d '{}' localhost:50051 grpc.health.v1.Health/Check

# Call a method
grpcurl -plaintext -d '{"userId":"test"}' localhost:50051 backend_shared.AdminService.RecordUserCreated

# gRPC UI (install grpcui first)
grpcui -plaintext localhost:50051
```

### Service Logs
```bash
# Auth service logs
docker logs auth-service

# Admin service logs
docker logs admin-service

# Kafka logs
docker logs auth-service-kafka

# Infrastructure logs
docker-compose -f infrastructure/docker-compose.yml logs
```

## Troubleshooting

### Common Issues

1. **Services not starting**
   - Ensure Docker is running
   - Check port conflicts
   - Verify docker-compose files exist

2. **Kafka connection failures**
   - Wait for Kafka to be fully started (30+ seconds)
   - Check network connectivity
   - Verify broker address

3. **gRPC connection failures**
   - Ensure admin service is running
   - Check gRPC port (50051)
   - Verify service registration

4. **Test failures**
   - Check service logs for detailed errors
   - Ensure all dependencies are running
   - Verify configuration values

### Network Issues

If running in different networks:

```bash
# Check Docker networks
docker network ls

# Connect services to the same network
docker network connect auth-network auth-service
docker network connect auth-network admin-service
```

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Auth Service  │───▶│      Kafka      │───▶│  Notification   │
│                 │    │   (Events)      │    │    Service      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              │
         │ gRPC                                         │
         ▼                                              ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Admin Service  │◀──▶│   Database      │    │  WebSocket      │
│   (User Mgmt)   │    │                 │    │   Chat Hub       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Contributing

When adding new tests:

1. Follow the existing script structure
2. Include proper error handling
3. Add descriptive logging
4. Update this README
5. Test with both automated and manual execution

## Dependencies

- Docker and Docker Compose
- Bash shell
- curl (for HTTP tests)
- nc (netcat) for basic connectivity checks
- grpcurl (optional, for advanced gRPC testing)

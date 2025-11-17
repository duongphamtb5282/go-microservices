# GraphQL Service Makefile
.PHONY: help build run test clean migrate status

# Default target
help:
	@echo "GraphQL Service - Available Commands:"
	@echo "====================================="
	@echo "build         - Build the GraphQL service"
	@echo "run           - Run the GraphQL service"
	@echo "test          - Run all tests"
	@echo "test-resolvers- Test GraphQL resolvers and mutations"
	@echo "test-docker   - Test with Docker Compose"
	@echo "migrate       - Run database migrations"
	@echo "status        - Check migration status"
	@echo "clean         - Clean build artifacts"
	@echo "docker        - Build Docker image"
	@echo "compose       - Start with Docker Compose"

# Build the service
build:
	@echo "Building GraphQL service..."
	go build -o graphql-service ./cmd/server
	go build -o migrate-tool ./cmd/migrate
	@echo "✅ Build completed"

# Run the service
run: build
	@echo "Starting GraphQL service..."
	./graphql-service

# Run tests
test:
	@echo "Running GraphQL service tests..."
	./scripts/test_mongodb_migration.sh

# Test GraphQL resolvers and mutations
test-resolvers: build
	@echo "Testing GraphQL resolvers and mutations..."
	./scripts/test_graphql_resolvers.sh
	@echo "✅ Resolver tests completed"

# Test with Docker Compose
test-docker: build
	@echo "Testing with Docker Compose..."
	./scripts/test_with_docker.sh
	@echo "✅ Docker tests completed"

# Run migrations
migrate: build
	@echo "Running database migrations..."
	./migrate-tool -action=migrate

# Check migration status
status: build
	@echo "Checking migration status..."
	./migrate-tool -action=status

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f graphql-service migrate-tool
	@echo "✅ Cleanup completed"

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t graphql-service .

# Start with Docker Compose
compose:
	@echo "Starting with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose
compose-down:
	@echo "Stopping Docker Compose..."
	docker-compose down

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod tidy
	go mod download
	@echo "✅ Development setup completed"

# Full test suite
test-full: clean build migrate test
	@echo "✅ Full test suite completed"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o graphql-service ./cmd/server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate-tool ./cmd/migrate
	@echo "✅ Production build completed"

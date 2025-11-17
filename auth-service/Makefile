# Auth Service Makefile

# Variables
BINARY_NAME=auth-service
BINARY_PATH=./cmd/http
BUILD_DIR=./bin
DOCKER_COMPOSE_FILE=docker-compose.yml

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@go run $(BINARY_PATH)

# Run with hot reload (Air)
.PHONY: dev-hot
dev-hot:
	@echo "Starting hot reload development server..."
	@air

# Run with hot reload and debug logs
.PHONY: dev-debug
dev-debug:
	@echo "Starting hot reload with debug logging..."
	@LOG_LEVEL=debug air

# Run with hot reload and verbose output
.PHONY: dev-verbose
dev-verbose:
	@echo "Starting hot reload with verbose output..."
	@air -c .air.toml

# Debug build (with debug symbols)
.PHONY: build-debug
build-debug:
	@echo "Building with debug symbols..."
	@mkdir -p $(BUILD_DIR)
	@go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME)-debug $(BINARY_PATH)
	@echo "Debug build completed: $(BUILD_DIR)/$(BINARY_NAME)-debug"

# Install development tools
.PHONY: dev-tools
dev-tools:
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@go install github.com/ramya-rao-a/go-outline@latest
	@echo "Development tools installed"

# Setup development environment
.PHONY: dev-setup
dev-setup: dev-tools
	@echo "Setting up development environment..."
	@echo "Creating .env file from template..."
	@cp env.example .env 2>/dev/null || echo "env.example not found, skipping .env creation"
	@echo "Development environment setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "1. Update .env file with your local configuration"
	@echo "2. Run 'make generate-secret' to generate secure secrets"
	@echo "3. Run 'make dev-hot' to start with hot reload"
	@echo "4. Use F5 in VSCode to start debugging"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@golangci-lint run

# Generate Swagger docs
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/http/main.go -o ./docs

# Migration commands
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	@go run ./cmd/migrate -command=up

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back database migrations..."
	@go run ./cmd/migrate -command=down -target=$(TARGET)

.PHONY: migrate-status
migrate-status:
	@echo "Checking migration status..."
	@go run ./cmd/migrate -command=status

.PHONY: migrate-create
migrate-create:
	@echo "Creating new migration..."
	@go run ./cmd/migrate -command=create -name=$(NAME)

.PHONY: migrate-validate
migrate-validate:
	@echo "Validating migrations..."
	@go run ./cmd/migrate -command=validate

# Start Docker services
.PHONY: docker-up
docker-up:
	@echo "Starting Docker services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

# Stop Docker services
.PHONY: docker-down
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

# Restart Docker services
.PHONY: docker-restart
docker-restart: docker-down docker-up

# Show Docker logs
.PHONY: docker-logs
docker-logs:
	@echo "Showing Docker logs..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

# Run all tests including integration tests
.PHONY: test-all
test-all: docker-up
	@echo "Waiting for services to start..."
	@sleep 10
	@echo "Running all tests..."
	@cd scripts && ./test_correlation_id.sh
	@cd scripts && ./test_cache.sh
	@cd scripts && ./test_event_driven.sh
	@echo "All tests completed!"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Generate secure Keycloak client secret
.PHONY: generate-secret
generate-secret:
	@echo "Generating secure Keycloak client secret..."
	@./scripts/generate-keycloak-secret.sh $(ARGS)

# Generate secret for specific environment (dev/prod)
.PHONY: generate-secret-dev
generate-secret-dev:
	@echo "Generating development Keycloak client secret..."
	@./scripts/generate-keycloak-secret.sh -e KEYCLOAK_CLIENT_SECRET_DEV -f .env.dev

.PHONY: generate-secret-prod
generate-secret-prod:
	@echo "Generating production Keycloak client secret..."
	@./scripts/generate-keycloak-secret.sh -e KEYCLOAK_CLIENT_SECRET_PROD -f .env.prod

# Validate secrets
.PHONY: validate-secrets
validate-secrets:
	@echo "Validating secrets..."
	@./scripts/validate-secrets.sh

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build                - Build the application"
	@echo "  run                  - Build and run the application"
	@echo "  dev                  - Run in development mode"
	@echo "  dev-hot              - Run with hot reload (Air)"
	@echo "  dev-debug            - Run with hot reload and debug logs"
	@echo "  dev-verbose          - Run with hot reload and verbose output"
	@echo "  build-debug          - Build with debug symbols"
	@echo "  dev-tools            - Install development tools (Air, Delve, etc.)"
	@echo "  dev-setup            - Setup complete development environment"
	@echo "  clean                - Clean build artifacts"
	@echo "  test                 - Run tests"
	@echo "  test-coverage        - Run tests with coverage"
	@echo "  fmt                  - Format code"
	@echo "  lint                 - Lint code"
	@echo "  swagger              - Generate Swagger docs"
	@echo "  migrate-up           - Run database migrations"
	@echo "  migrate-down         - Rollback migrations (use TARGET=version)"
	@echo "  migrate-status       - Check migration status"
	@echo "  migrate-create       - Create new migration (use NAME=name)"
	@echo "  migrate-validate     - Validate migrations"
	@echo "  docker-up            - Start Docker services"
	@echo "  docker-down          - Stop Docker services"
	@echo "  docker-restart       - Restart Docker services"
	@echo "  docker-logs          - Show Docker logs"
	@echo "  test-all             - Run all tests including integration"
	@echo "  deps                 - Install dependencies"
	@echo "  deps-update          - Update dependencies"
	@echo "  generate-secret      - Generate secure Keycloak client secret (use ARGS for options)"
	@echo "  generate-secret-dev  - Generate dev environment secret"
	@echo "  generate-secret-prod - Generate prod environment secret"
	@echo "  validate-secrets     - Validate existing secrets"
	@echo "  help                 - Show this help"
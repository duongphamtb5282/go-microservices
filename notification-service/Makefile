# Notification Service Makefile
.PHONY: help build run test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building notification-service..."
	@go build -o notification-service ./cmd/server
	@go build -o notification-service-root .
	@echo "✅ Build completed"

build-server: ## Build only the server version
	@echo "Building notification-service server..."
	@go build -o notification-service ./cmd/server
	@echo "✅ Server build completed"

build-root: ## Build only the root version
	@echo "Building notification-service root..."
	@go build -o notification-service-root .
	@echo "✅ Root build completed"

run: ## Run the application
	@echo "Running notification-service..."
	@./notification-service

run-server: ## Run the server version
	@echo "Running notification-service server..."
	@./notification-service

run-root: ## Run the root version
	@echo "Running notification-service root..."
	@./notification-service-root

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f notification-service
	@rm -f notification-service-root
	@rm -f coverage.out coverage.html

deps: ## Download and tidy dependencies
	@echo "Updating dependencies..."
	@go mod tidy
	@go mod download

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t notification-service:latest .

.DEFAULT_GOAL := help

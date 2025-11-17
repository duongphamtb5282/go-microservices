.PHONY: help build run test clean migrate-up migrate-down proto

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building admin-service..."
	@go build -o admin-service cmd/server/main.go

run: ## Run the application
	@echo "Running admin-service..."
	@go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f admin-service
	@rm -f coverage.out coverage.html

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	@psql -h $(DATABASE_HOST) -U $(DATABASE_USERNAME) -d $(DATABASE_NAME) -f migrations/001_create_user_events_table.sql

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	@echo "Not implemented yet"

proto: ## Generate proto files (run from backend-shared)
	@echo "Generating proto files..."
	@cd ../backend-shared && ./proto/admin/generate.sh

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t admin-service:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8086:8086 -p 50051:50051 --env-file .env admin-service:latest

.DEFAULT_GOAL := help


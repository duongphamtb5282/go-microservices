#!/bin/bash

# Comprehensive Communication Testing Script
# Tests Kafka and gRPC communication for the auth service ecosystem

set -e

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
INFRA_COMPOSE_FILE="$PROJECT_ROOT/infrastructure/docker-compose.yml"
AUTH_SERVICE_COMPOSE_FILE="$PROJECT_ROOT/auth-service/docker-compose.app.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_header() {
    echo -e "${PURPLE}[TEST]${NC} $1"
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# Function to check if Docker is running
check_docker() {
    log_info "Checking if Docker is running..."

    if ! docker info >/dev/null 2>&1; then
        log_error "Docker is not running"
        log_error "Please start Docker and try again"
        exit 1
    fi

    log_success "Docker is running"
}

# Function to check if docker-compose is available
check_docker_compose() {
    log_info "Checking if docker-compose is available..."

    if command -v docker-compose >/dev/null 2>&1; then
        COMPOSE_CMD="docker-compose"
        log_success "docker-compose found"
    elif docker compose version >/dev/null 2>&1; then
        COMPOSE_CMD="docker compose"
        log_success "docker compose (plugin) found"
    else
        log_error "Neither docker-compose nor docker compose found"
        exit 1
    fi
}

# Function to start infrastructure
start_infrastructure() {
    log_header "Starting Infrastructure Services"

    log_step "Starting infrastructure (Redis, Kafka, Zookeeper, Monitoring)..."
    cd "$PROJECT_ROOT"

    if [ -f "$INFRA_COMPOSE_FILE" ]; then
        $COMPOSE_CMD -f "$INFRA_COMPOSE_FILE" up -d
        log_success "Infrastructure services started"

        # Wait for services to be healthy
        log_info "Waiting for services to be ready..."
        sleep 30

        # Check service health
        check_service_health "kafka" "9092"
        check_service_health "redis" "6379"
    else
        log_warn "Infrastructure compose file not found: $INFRA_COMPOSE_FILE"
    fi
}

# Function to start application services
start_application_services() {
    log_header "Starting Application Services"

    log_step "Starting admin service..."
    cd "$PROJECT_ROOT/admin-service"

    # Try to start admin service
    if [ -f "docker-compose.yml" ]; then
        $COMPOSE_CMD up -d
        log_success "Admin service started"
    elif [ -f "Dockerfile" ]; then
        log_info "Building and starting admin service..."
        docker build -t admin-service .
        docker run -d --name admin-service \
            --network auth-network \
            -p 50051:50051 \
            -p 8086:8086 \
            -e DATABASE_HOST=postgres \
            -e DATABASE_PORT=5432 \
            -e DATABASE_USERNAME=admin_user \
            -e DATABASE_PASSWORD=admin_password \
            -e DATABASE_NAME=admin_service \
            admin-service
        log_success "Admin service started (built from Dockerfile)"
    else
        log_warn "No way to start admin service found"
    fi

    log_step "Starting auth service..."
    cd "$PROJECT_ROOT/auth-service"

    if [ -f "docker-compose.app.yml" ]; then
        $COMPOSE_CMD -f docker-compose.app.yml up -d
        log_success "Auth service started"
    elif [ -f "Dockerfile" ]; then
        log_info "Building and starting auth service..."
        docker build -t auth-service .
        docker run -d --name auth-service \
            --network auth-network \
            -p 8085:8085 \
            -e KAFKA_BROKERS=kafka:9092 \
            -e DATABASE_HOST=postgres \
            -e DATABASE_PORT=5432 \
            auth-service
        log_success "Auth service started (built from Dockerfile)"
    else
        log_warn "No way to start auth service found"
    fi

    # Wait for services to be ready
    log_info "Waiting for application services to be ready..."
    sleep 20

    check_service_health "auth-service" "8085"
    check_service_health "admin-service" "8086"
}

# Function to check service health
check_service_health() {
    local service_name=$1
    local port=$2

    log_info "Checking $service_name health..."

    if timeout 30 bash -c "</dev/tcp/localhost/$port" 2>/dev/null; then
        log_success "$service_name is accessible on port $port"
    else
        log_error "$service_name is not accessible on port $port"
        return 1
    fi
}

# Function to run Kafka tests
run_kafka_tests() {
    log_header "Running Kafka Communication Tests"

    local test_script="$PROJECT_ROOT/auth-service/scripts/test-kafka-integration.sh"

    if [ -f "$test_script" ]; then
        log_step "Running Kafka integration tests..."
        bash "$test_script"
    else
        log_error "Kafka test script not found: $test_script"
        return 1
    fi
}

# Function to run gRPC tests
run_grpc_tests() {
    log_header "Running gRPC Communication Tests"

    local test_script="$PROJECT_ROOT/auth-service/scripts/test-grpc-integration.sh"

    if [ -f "$test_script" ]; then
        log_step "Running gRPC integration tests..."
        bash "$test_script"
    else
        log_error "gRPC test script not found: $test_script"
        return 1
    fi
}

# Function to run end-to-end tests
run_end_to_end_tests() {
    log_header "Running End-to-End Communication Tests"

    log_step "Testing complete auth flow with Kafka and gRPC..."

    # This would test the complete flow: auth service -> Kafka -> admin service -> gRPC
    # For now, we'll just run both individual tests
    run_kafka_tests
    run_grpc_tests
}

# Function to show service status
show_service_status() {
    log_header "Service Status"

    echo ""
    echo "Infrastructure Services:"
    echo "-----------------------"

    # Check infrastructure services
    check_service_status "kafka" "9092"
    check_service_status "zookeeper" "2181"
    check_service_status "redis" "6379"
    check_service_status "prometheus" "9090"
    check_service_status "grafana" "3000"

    echo ""
    echo "Application Services:"
    echo "--------------------"

    # Check application services
    check_service_status "postgres" "5432"
    check_service_status "auth-service" "8085"
    check_service_status "admin-service-grpc" "50051"
    check_service_status "admin-service-http" "8086"

    echo ""
    echo "Monitoring URLs:"
    echo "----------------"
    echo "Kafka UI:        http://localhost:8080"
    echo "Prometheus:      http://localhost:9090"
    echo "Grafana:         http://localhost:3000"
    echo "Jaeger:          http://localhost:16686"
    echo ""
}

# Function to check individual service status
check_service_status() {
    local service_name=$1
    local port=$2

    if timeout 5 bash -c "</dev/tcp/localhost/$port" 2>/dev/null; then
        echo -e "  ✅ $service_name: Running (port $port)"
    else
        echo -e "  ❌ $service_name: Not accessible (port $port)"
    fi
}

# Function to cleanup services
cleanup_services() {
    log_header "Cleaning Up Services"

    log_warn "This will stop all running services"

    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log_step "Stopping application services..."
        cd "$PROJECT_ROOT/auth-service"
        $COMPOSE_CMD -f docker-compose.app.yml down 2>/dev/null || true

        cd "$PROJECT_ROOT/admin-service"
        $COMPOSE_CMD down 2>/dev/null || true

        log_step "Stopping infrastructure services..."
        cd "$PROJECT_ROOT"
        $COMPOSE_CMD -f "$INFRA_COMPOSE_FILE" down 2>/dev/null || true

        # Remove test containers
        docker rm -f auth-service admin-service 2>/dev/null || true

        log_success "Cleanup completed"
    else
        log_info "Cleanup cancelled"
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start          Start all infrastructure and application services"
    echo "  stop           Stop all services and cleanup"
    echo "  status         Show status of all services"
    echo "  kafka          Run only Kafka communication tests"
    echo "  grpc           Run only gRPC communication tests"
    echo "  test           Run all communication tests"
    echo "  e2e            Run end-to-end tests"
    echo "  help           Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  KAFKA_BROKER          Kafka broker address (default: localhost:9092)"
    echo "  AUTH_SERVICE_URL      Auth service URL (default: http://localhost:8085)"
    echo "  ADMIN_SERVICE_GRPC    Admin service gRPC address (default: localhost:50051)"
    echo ""
    echo "Examples:"
    echo "  $0 start           # Start all services"
    echo "  $0 test            # Run all tests"
    echo "  $0 kafka           # Run only Kafka tests"
    echo "  KAFKA_BROKER=kafka.example.com:9092 $0 test"
}

# Main execution function
main() {
    local command=${1:-"help"}

    # Check prerequisites
    check_docker
    check_docker_compose

    case $command in
        "start")
            start_infrastructure
            start_application_services
            show_service_status
            ;;
        "stop")
            cleanup_services
            ;;
        "status")
            show_service_status
            ;;
        "kafka")
            run_kafka_tests
            ;;
        "grpc")
            run_grpc_tests
            ;;
        "test")
            run_kafka_tests
            run_grpc_tests
            ;;
        "e2e")
            run_end_to_end_tests
            ;;
        "help"|*)
            show_usage
            ;;
    esac
}

# Run main function
main "$@"

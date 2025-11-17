#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_header() {
    echo -e "${BLUE}$1${NC}"
}

# Set default environment variables for database
export DATABASE_USERNAME="${DATABASE_USERNAME:-auth_user}"
export DATABASE_PASSWORD="${DATABASE_PASSWORD:-auth_password}"
export DATABASE_HOST="${DATABASE_HOST:-localhost}"
export DATABASE_NAME="${DATABASE_NAME:-auth_service}"
export DATABASE_PORT="${DATABASE_PORT:-5432}"

print_header "========================================"
print_header "  Auth Service Startup Script"
print_header "========================================"
echo ""

print_info "Database Configuration:"
echo "  Host: $DATABASE_HOST"
echo "  Port: $DATABASE_PORT"
echo "  Database: $DATABASE_NAME"
echo "  Username: $DATABASE_USERNAME"
echo ""

# Check if PostgreSQL is running
print_info "Checking PostgreSQL connection..."
if docker-compose exec -T postgres psql -U "$DATABASE_USERNAME" -d "$DATABASE_NAME" -c "SELECT 1" > /dev/null 2>&1; then
    print_success "PostgreSQL is running and accessible"
else
    print_error "PostgreSQL is not accessible. Starting with docker-compose..."
    docker-compose up -d postgres
    sleep 3
    
    # Check again
    if docker-compose exec -T postgres psql -U "$DATABASE_USERNAME" -d "$DATABASE_NAME" -c "SELECT 1" > /dev/null 2>&1; then
        print_success "PostgreSQL started successfully"
    else
        print_error "Failed to start PostgreSQL"
        exit 1
    fi
fi

# Check if auth-service binary exists
if [ ! -f "./auth-service" ]; then
    print_info "Building auth-service..."
    if go build -o auth-service cmd/http/main.go; then
        print_success "Build completed"
    else
        print_error "Build failed"
        exit 1
    fi
fi

# Kill any existing instance
print_info "Checking for existing instances..."
if lsof -ti:8085 > /dev/null 2>&1; then
    print_info "Stopping existing instance on port 8085..."
    lsof -ti:8085 | xargs kill -9 2>/dev/null
    sleep 2
    print_success "Existing instance stopped"
fi

# Start the service
print_info "Starting auth-service..."
LOG_FILE="auth-service.log"
./auth-service > "$LOG_FILE" 2>&1 &
AUTH_PID=$!

# Wait for startup
sleep 3

# Check if service started successfully
if ps -p $AUTH_PID > /dev/null; then
    print_success "Auth-service started successfully (PID: $AUTH_PID)"
    print_info "Log file: $LOG_FILE"
    
    # Check logs for PostgreSQL repository confirmation
    if grep -q "✅ Using PostgreSQL UserRepository" "$LOG_FILE"; then
        print_success "PostgreSQL repository is being used"
    else
        print_error "Warning: Service may be using in-memory repository"
        print_info "Check logs: tail -f $LOG_FILE"
    fi
    
    # Test the health endpoint
    sleep 2
    print_info "Testing service health..."
    if curl -s http://localhost:8085/health > /dev/null 2>&1; then
        print_success "Service is responding to HTTP requests"
    else
        print_error "Service is not responding (may still be starting up)"
    fi
    
    echo ""
    print_header "========================================"
    print_success "Auth Service is running!"
    print_header "========================================"
    echo ""
    echo "Service URL: http://localhost:8085"
    echo "API Documentation: http://localhost:8085/swagger/index.html"
    echo ""
    echo "To view logs: tail -f $LOG_FILE"
    echo "To stop: kill $AUTH_PID"
    echo ""
    
else
    print_error "Failed to start auth-service"
    print_info "Check logs for details: tail -20 $LOG_FILE"
    exit 1
fi


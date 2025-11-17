#!/bin/bash

# Quick Setup Script for Auth Service
# This script automates the entire environment setup

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}→ $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

print_header "Auth Service - Quick Setup"

# Step 1: Check prerequisites
print_info "Checking prerequisites..."

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.21 or higher"
    exit 1
fi
print_success "Go $(go version | awk '{print $3}') found"

if ! command_exists docker; then
    print_error "Docker is not installed. Please install Docker"
    exit 1
fi
print_success "Docker $(docker --version | awk '{print $3}' | tr -d ',') found"

if ! command_exists docker-compose; then
    print_error "Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi
print_success "Docker Compose found"

# Step 2: Install Go dependencies
print_header "Installing Go Dependencies"
print_info "Running go mod download..."
go mod download
go mod tidy
print_success "Dependencies installed"

# Step 3: Create .env file
print_header "Creating Environment Configuration"

if [ -f .env ]; then
    print_info ".env file already exists. Backing up..."
    cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    print_success "Backup created"
fi

cat > .env << 'EOF'
# Application
APP_ENV=development
PORT=8085
LOG_LEVEL=debug

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=auth_service
DATABASE_USERNAME=auth_user
DATABASE_PASSWORD=auth_password
DATABASE_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-min-32-chars-long-change-in-production
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h
JWT_ISSUER=auth-service
JWT_AUDIENCE=auth-service-api
JWT_BCRYPT_COST=10

# PingAM Configuration
PINGAM_BASE_URL=http://localhost:1080
PINGAM_CLIENT_ID=auth-service-client
PINGAM_CLIENT_SECRET=your-client-secret
PINGAM_REDIRECT_URI=http://localhost:8085/callback
PINGAM_TIMEOUT=30s
PINGAM_CACHE_TTL=5m

# OpenTelemetry (Optional)
OTEL_ENABLED=false
OTEL_SERVICE_NAME=auth-service
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=development
EOF

chmod 600 .env
print_success ".env file created"

# Step 4: Start infrastructure services
print_header "Starting Infrastructure Services"

print_info "Starting PostgreSQL, Redis, Kafka, MongoDB..."
docker-compose up -d postgres redis kafka mongodb

print_info "Waiting for services to be healthy (30 seconds)..."
sleep 30

# Verify services
SERVICES_OK=true

if docker-compose ps | grep -q "postgres.*Up"; then
    print_success "PostgreSQL is running"
else
    print_error "PostgreSQL failed to start"
    SERVICES_OK=false
fi

if docker-compose ps | grep -q "redis.*Up"; then
    print_success "Redis is running"
else
    print_error "Redis failed to start"
    SERVICES_OK=false
fi

if docker-compose ps | grep -q "kafka.*Up"; then
    print_success "Kafka is running"
else
    print_error "Kafka failed to start"
    SERVICES_OK=false
fi

if [ "$SERVICES_OK" = false ]; then
    print_error "Some services failed to start. Check docker-compose logs"
    exit 1
fi

# Step 5: Initialize database
print_header "Initializing Database"

print_info "Waiting for PostgreSQL to accept connections..."
sleep 10

# Check if migrations directory exists
if [ -d "migrations" ] && [ "$(ls -A migrations)" ]; then
    print_info "Running database migrations..."
    
    # Run migrations
    for file in migrations/*.sql; do
        if [ -f "$file" ]; then
            print_info "Running $(basename $file)..."
            docker-compose exec -T postgres psql -U auth_user -d auth_service < "$file" || true
        fi
    done
    
    print_success "Migrations completed"
else
    print_info "No migrations found, skipping..."
fi

# Verify database
if docker-compose exec -T postgres psql -U auth_user -d auth_service -c "SELECT 1" > /dev/null 2>&1; then
    print_success "Database connection verified"
else
    print_error "Database connection failed"
    exit 1
fi

# Step 6: Start PingAM Mock Server
print_header "Starting PingAM Mock Server"

if [ -f "docker-compose.pingam.yml" ]; then
    print_info "Starting PingAM mock and Swagger UI..."
    docker-compose -f docker-compose.pingam.yml up -d pingam-mock swagger-ui
    
    sleep 10
    
    if curl -s http://localhost:1080/health > /dev/null 2>&1; then
        print_success "PingAM Mock is running at http://localhost:1080"
    else
        print_error "PingAM Mock failed to start"
    fi
    
    if curl -s -I http://localhost:8081 > /dev/null 2>&1; then
        print_success "Swagger UI is running at http://localhost:8081"
    else
        print_error "Swagger UI failed to start"
    fi
else
    print_info "docker-compose.pingam.yml not found, skipping PingAM setup"
fi

# Step 7: Build the application
print_header "Building Auth Service"

print_info "Compiling application..."
go build -o auth-service cmd/http/main.go
print_success "Build successful"

# Step 8: Create logs directory
mkdir -p logs
print_success "Logs directory created"

# Step 9: Make test scripts executable
print_header "Setting Up Test Scripts"

for script in test-*.sh run-all-tests.sh verify-setup.sh; do
    if [ -f "$script" ]; then
        chmod +x "$script"
        print_success "Made $script executable"
    fi
done

# Step 10: Environment verification
print_header "Environment Verification"

cat > verify-setup.sh << 'VERIFY_EOF'
#!/bin/bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "=== Environment Verification ==="
echo ""

# Check PostgreSQL
echo -n "PostgreSQL: "
if docker-compose exec -T postgres psql -U auth_user -d auth_service -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

# Check Redis
echo -n "Redis: "
if docker-compose exec redis redis-cli ping 2>/dev/null | grep -q PONG; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

# Check Kafka
echo -n "Kafka: "
if docker-compose ps | grep -q "kafka.*Up"; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

# Check PingAM Mock
echo -n "PingAM Mock: "
if curl -s http://localhost:1080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running at http://localhost:1080${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

# Check Swagger UI
echo -n "Swagger UI: "
if curl -s -I http://localhost:8081 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running at http://localhost:8081${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
fi

echo ""
echo "=== Environment Setup Complete ==="
VERIFY_EOF

chmod +x verify-setup.sh
./verify-setup.sh

# Final summary
print_header "Setup Complete! 🎉"

echo -e "${GREEN}Your auth service environment is ready!${NC}"
echo ""
echo "Next steps:"
echo ""
echo "1. Start the auth service:"
echo "   ${YELLOW}./auth-service${NC}"
echo ""
echo "2. Run tests:"
echo "   ${YELLOW}./run-all-tests.sh${NC}"
echo ""
echo "3. Access Swagger UI:"
echo "   ${YELLOW}http://localhost:8081${NC}"
echo ""
echo "4. Check health:"
echo "   ${YELLOW}curl http://localhost:8085/api/v1/health${NC}"
echo ""
echo "5. View logs:"
echo "   ${YELLOW}tail -f logs/auth-service.log${NC}"
echo ""
echo "For detailed testing guide, see:"
echo "   ${YELLOW}SETUP_AND_TESTING_GUIDE.md${NC}"
echo ""



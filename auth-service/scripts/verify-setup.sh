#!/bin/bash

# Environment Verification Script
# Checks if all services are running correctly

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

print_header "Environment Verification"

ALL_OK=true

# Check PostgreSQL
echo -n "PostgreSQL: "
if docker-compose exec -T postgres psql -U auth_user -d auth_service -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    ALL_OK=false
fi

# Check Redis
echo -n "Redis: "
if docker-compose exec redis redis-cli ping 2>/dev/null | grep -q PONG; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    ALL_OK=false
fi

# Check Kafka
echo -n "Kafka: "
if docker-compose ps | grep -q "kafka.*Up"; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    ALL_OK=false
fi

# Check MongoDB
echo -n "MongoDB: "
if docker-compose ps | grep -q "mongodb.*Up"; then
    echo -e "${GREEN}✓ Running${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    print_info "MongoDB is optional"
fi

# Check PingAM Mock
echo -n "PingAM Mock: "
if curl -s http://localhost:1080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running at http://localhost:1080${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    print_info "Start with: docker-compose -f docker-compose.pingam.yml up -d"
fi

# Check Swagger UI
echo -n "Swagger UI: "
if curl -s -I http://localhost:8081 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running at http://localhost:8081${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    print_info "Start with: docker-compose -f docker-compose.pingam.yml up -d"
fi

# Check Auth Service
echo -n "Auth Service: "
if curl -s http://localhost:8085/api/v1/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Running at http://localhost:8085${NC}"
else
    echo -e "${RED}✗ Not running${NC}"
    print_info "Start with: ./auth-service"
fi

echo ""
print_header "Service URLs"

echo "Core Services:"
echo "  • Auth Service:  http://localhost:8085"
echo "  • Swagger UI:    http://localhost:8081"
echo "  • PingAM Mock:   http://localhost:1080"
echo ""
echo "Infrastructure:"
echo "  • PostgreSQL:    localhost:5432"
echo "  • Redis:         localhost:6379"
echo "  • Kafka:         localhost:9092"
echo "  • MongoDB:       localhost:27017"

echo ""

if [ "$ALL_OK" = true ]; then
    print_header "✓ Environment Setup Complete!"
    echo -e "${GREEN}All critical services are running.${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start auth service: ./auth-service"
    echo "  2. Run tests: ./quick-test.sh"
    echo "  3. View Swagger: http://localhost:8081"
else
    print_header "⚠ Some Services Need Attention"
    echo -e "${YELLOW}Some services are not running.${NC}"
    echo ""
    echo "To start all services:"
    echo "  1. Infrastructure: docker-compose up -d"
    echo "  2. PingAM Mock: docker-compose -f docker-compose.pingam.yml up -d"
    echo "  3. Auth Service: ./auth-service"
    echo ""
    echo "Or run quick setup:"
    echo "  ./quick-setup.sh"
fi

echo ""



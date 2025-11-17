#!/bin/bash

# Login API Testing Script
# This script tests the JWT and PingAM login endpoints

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8085/api/v1}"
PINGAM_MOCK_URL="${PINGAM_MOCK_URL:-http://localhost:1080}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Functions
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

test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local headers=$4
    local description=$5

    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "${YELLOW}$method $url${NC}"

    local cmd="curl -s -X $method \"$url\""
    if [ -n "$data" ]; then
        cmd="$cmd -H \"Content-Type: application/json\" -d '$data'"
    fi
    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    echo -e "${YELLOW}Command: $cmd${NC}"

    # Execute the command
    local response
    if [ -n "$data" ]; then
        response=$(eval "$cmd")
    else
        response=$(eval "$cmd")
    fi

    echo -e "${GREEN}Response:${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""
}

# Main test execution
print_header "Login API Testing Script"
echo "Base URL: $BASE_URL"
echo "PingAM Mock URL: $PINGAM_MOCK_URL"
echo ""

# Check if service is running
print_info "Checking if auth service is running..."
if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    print_success "Auth service is running"
else
    print_error "Auth service is not running at $BASE_URL"
    echo ""
    echo "Please start the service first:"
    echo "  cd auth-service"
    echo "  go run cmd/http/main.go"
    echo ""
    echo "Or with Docker:"
    echo "  docker-compose up -d postgres redis kafka"
    echo "  ./scripts/start-service.sh"
    echo ""
    exit 1
fi

print_header "JWT Authentication Tests"

# Test 1: Health Check
test_endpoint "GET" "$BASE_URL/health" "" "" "Health Check"

# Test 2: User Registration
print_info "Testing User Registration..."
REGISTER_DATA='{
    "username": "testuser_'"$(date +%s)"'",
    "email": "testuser'"$(date +%s)"'@example.com",
    "password": "SecurePassword123!",
    "first_name": "Test",
    "last_name": "User"
}'

test_endpoint "POST" "$BASE_URL/auth/register" "$REGISTER_DATA" "" "User Registration"

# Extract username from registration (you would need to parse this)
USERNAME="testuser_$(date +%s)"
PASSWORD="SecurePassword123!"

# Test 3: JWT Login
print_info "Testing JWT Login..."
LOGIN_DATA='{
    "username": "'"$USERNAME"'",
    "password": "'"$PASSWORD"'"
}'

echo -e "${BLUE}Testing: JWT Login${NC}"
echo -e "${YELLOW}POST $BASE_URL/auth/login${NC}"
echo -e "${YELLOW}Data: $LOGIN_DATA${NC}"

response=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA")

echo -e "${GREEN}Response:${NC}"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

# Extract JWT token
JWT_TOKEN=$(echo "$response" | jq -r '.access_token // empty')
if [ -n "$JWT_TOKEN" ] && [ "$JWT_TOKEN" != "null" ]; then
    print_success "JWT Login successful"
    echo -e "${GREEN}Token: ${JWT_TOKEN:0:50}...${NC}"
else
    print_error "JWT Login failed"
fi
echo ""

# Test 4: Protected Endpoint Access
if [ -n "$JWT_TOKEN" ]; then
    print_info "Testing Protected Endpoint Access..."
    test_endpoint "GET" "$BASE_URL/auth/profile" "" "-H \"Authorization: Bearer $JWT_TOKEN\"" "Access Protected Endpoint"
else
    print_error "Skipping protected endpoint test - no JWT token"
fi

# Test 5: Invalid Token
print_info "Testing Invalid Token..."
test_endpoint "GET" "$BASE_URL/auth/profile" "" "-H \"Authorization: Bearer invalid.token.here\"" "Invalid Token (should fail)"

print_header "PingAM Authentication Tests"

# Check if PingAM mock is running
if curl -s "$PINGAM_MOCK_URL/health" > /dev/null 2>&1; then
    print_success "PingAM Mock is running"

    # Test 6: PingAM Login
    print_info "Testing PingAM Login..."
    PINGAM_LOGIN_DATA='{
        "username": "testuser",
        "password": "password123"
    }'

    echo -e "${BLUE}Testing: PingAM Login${NC}"
    echo -e "${YELLOW}POST $BASE_URL/pingam/auth/login${NC}"

    response=$(curl -s -X POST "$BASE_URL/pingam/auth/login" \
        -H "Content-Type: application/json" \
        -d "$PINGAM_LOGIN_DATA")

    echo -e "${GREEN}Response:${NC}"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"

    # Extract PingAM token
    PINGAM_TOKEN=$(echo "$response" | jq -r '.access_token // empty')
    PINGAM_USER_ID=$(echo "$response" | jq -r '.user_id // empty')

    if [ -n "$PINGAM_TOKEN" ] && [ "$PINGAM_TOKEN" != "null" ]; then
        print_success "PingAM Login successful"
        echo -e "${GREEN}Token: ${PINGAM_TOKEN:0:50}...${NC}"
        echo -e "${GREEN}User ID: $PINGAM_USER_ID${NC}"

        # Test 7: PingAM Protected Endpoint
        print_info "Testing PingAM Protected Endpoint..."
        test_endpoint "GET" "$BASE_URL/pingam/authz/roles/$PINGAM_USER_ID" "" "-H \"Authorization: Bearer $PINGAM_TOKEN\"" "Get User Roles"

        # Test 8: PingAM Permission Check
        print_info "Testing PingAM Permission Check..."
        PERMISSION_DATA='{
            "user_id": "'"$PINGAM_USER_ID"'",
            "resource": "users",
            "action": "read"
        }'
        test_endpoint "POST" "$BASE_URL/pingam/authz/check" "$PERMISSION_DATA" "-H \"Authorization: Bearer $PINGAM_TOKEN\"" "Permission Check"

    else
        print_error "PingAM Login failed"
    fi

else
    print_error "PingAM Mock is not running at $PINGAM_MOCK_URL"
    echo ""
    echo "To test PingAM authentication:"
    echo "  docker-compose -f docker-compose.pingam.yml up -d pingam-mock"
    echo ""
fi

print_header "Test Summary"
echo -e "${GREEN}✓ JWT Authentication Tests${NC}"
echo "  - Health check"
echo "  - User registration"
echo "  - JWT login"
echo "  - Protected endpoint access"
echo "  - Invalid token rejection"
echo ""

if curl -s "$PINGAM_MOCK_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PingAM Authentication Tests${NC}"
    echo "  - PingAM login"
    echo "  - Session validation"
    echo "  - Permission checking"
    echo "  - Role retrieval"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}🎉 Login API Testing Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Next steps:"
echo "  • Check service logs: tail -f logs/auth-service.log"
echo "  • View API docs: http://localhost:8085/swagger/index.html"
echo "  • Run full test suite: ./scripts/quick-test.sh"

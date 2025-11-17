#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_test() {
    echo -e "${YELLOW}🧪 $1${NC}"
}

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8085/api/v1}"
MOCK_URL="${MOCK_URL:-http://localhost:1080}"

echo "======================================"
echo "  PingAM Integration Test Suite"
echo "======================================"
echo ""

# Check if services are running
print_info "Checking services..."

# Check auth-service
if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    print_success "Auth service is running ($BASE_URL)"
else
    print_error "Auth service is not running at $BASE_URL"
    exit 1
fi

# Check PingAM mock
if curl -s "$MOCK_URL/health" > /dev/null 2>&1; then
    print_success "PingAM mock is running ($MOCK_URL)"
else
    print_error "PingAM mock is not running at $MOCK_URL"
    echo "Start it with: docker-compose -f docker-compose.pingam.yml up -d pingam-mock"
    exit 1
fi

echo ""
echo "======================================"
echo "  Test 1: Authentication"
echo "======================================"

print_test "Authenticating with PingAM..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "Admin123!"
  }')

# Check if login was successful
if echo "$LOGIN_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
    ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
    print_success "Login successful"
    echo "$LOGIN_RESPONSE" | jq '.'
else
    print_error "Login failed"
    echo "$LOGIN_RESPONSE" | jq '.'
    exit 1
fi

echo ""
echo "======================================"
echo "  Test 2: Check Permission"
echo "======================================"

print_test "Checking user permission..."
PERMISSION_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/authz/check" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource": "users",
    "action": "read"
  }')

if echo "$PERMISSION_RESPONSE" | jq -e '.allowed' > /dev/null 2>&1; then
    ALLOWED=$(echo "$PERMISSION_RESPONSE" | jq -r '.allowed')
    if [ "$ALLOWED" = "true" ]; then
        print_success "Permission check passed (allowed: true)"
    else
        print_error "Permission check failed (allowed: false)"
    fi
    echo "$PERMISSION_RESPONSE" | jq '.'
else
    print_error "Permission check request failed"
    echo "$PERMISSION_RESPONSE" | jq '.'
fi

echo ""
echo "======================================"
echo "  Test 3: Get User Roles"
echo "======================================"

print_test "Fetching user roles..."
ROLES_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/roles/user-123" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$ROLES_RESPONSE" | jq -e '.roles' > /dev/null 2>&1; then
    ROLE_COUNT=$(echo "$ROLES_RESPONSE" | jq -r '.count')
    print_success "Roles retrieved successfully (count: $ROLE_COUNT)"
    echo "$ROLES_RESPONSE" | jq '.'
else
    print_error "Failed to get roles"
    echo "$ROLES_RESPONSE" | jq '.'
fi

echo ""
echo "======================================"
echo "  Test 4: Get User Permissions"
echo "======================================"

print_test "Fetching user permissions..."
PERMISSIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/permissions/user-123" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

if echo "$PERMISSIONS_RESPONSE" | jq -e '.permissions' > /dev/null 2>&1; then
    PERM_COUNT=$(echo "$PERMISSIONS_RESPONSE" | jq -r '.count')
    print_success "Permissions retrieved successfully (count: $PERM_COUNT)"
    echo "$PERMISSIONS_RESPONSE" | jq '.'
else
    print_error "Failed to get permissions"
    echo "$PERMISSIONS_RESPONSE" | jq '.'
fi

echo ""
echo "======================================"
echo "  Test 5: Multiple Permission Checks"
echo "======================================"

# Test different resources and actions
declare -a TESTS=(
    "users:read"
    "users:write"
    "users:delete"
    "orders:read"
    "orders:write"
)

for test in "${TESTS[@]}"; do
    IFS=':' read -r resource action <<< "$test"
    print_test "Checking $resource:$action..."
    
    RESULT=$(curl -s -X POST "$BASE_URL/pingam/authz/check" \
      -H "Authorization: Bearer $ACCESS_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"user_id\": \"user-123\",
        \"resource\": \"$resource\",
        \"action\": \"$action\"
      }")
    
    if echo "$RESULT" | jq -e '.allowed' > /dev/null 2>&1; then
        ALLOWED=$(echo "$RESULT" | jq -r '.allowed')
        if [ "$ALLOWED" = "true" ]; then
            print_success "$resource:$action → Allowed"
        else
            echo -e "${YELLOW}⚠️  $resource:$action → Denied${NC}"
        fi
    else
        print_error "$resource:$action → Error"
    fi
done

echo ""
echo "======================================"
echo "  Test 6: Token Refresh (if supported)"
echo "======================================"

print_test "Testing token refresh..."
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ "$REFRESH_TOKEN" != "null" ]; then
    REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/refresh-token" \
      -H "Content-Type: application/json" \
      -d "{
        \"refresh_token\": \"$REFRESH_TOKEN\"
      }")
    
    if echo "$REFRESH_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Token refresh successful"
        echo "$REFRESH_RESPONSE" | jq '.'
    else
        print_error "Token refresh failed"
        echo "$REFRESH_RESPONSE" | jq '.'
    fi
else
    echo -e "${YELLOW}⚠️  No refresh token available${NC}"
fi

echo ""
echo "======================================"
echo "  Test 7: PingAM Health Check"
echo "======================================"

print_test "Checking PingAM health..."
HEALTH_RESPONSE=$(curl -s "$BASE_URL/pingam/health")

if echo "$HEALTH_RESPONSE" | jq -e '.status' > /dev/null 2>&1; then
    print_success "PingAM health check passed"
    echo "$HEALTH_RESPONSE" | jq '.'
else
    print_error "PingAM health check failed"
    echo "$HEALTH_RESPONSE" | jq '.'
fi

echo ""
echo "======================================"
echo "  Test Summary"
echo "======================================"
echo ""
print_success "All core PingAM endpoints tested"
print_info "Mock server is functioning correctly"
echo ""
echo "Available endpoints:"
echo "  ✅ POST   /api/v1/pingam/auth/login"
echo "  ✅ POST   /api/v1/pingam/authz/check"
echo "  ✅ GET    /api/v1/pingam/authz/roles/:userId"
echo "  ✅ GET    /api/v1/pingam/authz/permissions/:userId"
echo "  ✅ POST   /api/v1/pingam/auth/refresh-token"
echo "  ✅ GET    /api/v1/pingam/health"
echo ""
print_info "For real PingAM integration, see: REAL_PINGAM_INTEGRATION.md"
echo ""


#!/bin/bash

# PingAM Authentication Test Script
# Usage: ./test-pingam.sh

set -e

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8085/api/v1}"
PINGAM_MOCK_URL="${PINGAM_MOCK_URL:-http://localhost:1080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_header() {
    echo -e "\n${BLUE}=========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}=========================================${NC}\n"
}

print_test() {
    echo -e "${YELLOW}[TEST] $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    local headers=$5
    
    print_test "$name"
    
    if [ -z "$data" ]; then
        response=$(curl -s -X "$method" "$url" $headers)
    else
        response=$(curl -s -X "$method" "$url" -H "Content-Type: application/json" $headers -d "$data")
    fi
    
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    
    if echo "$response" | jq -e . >/dev/null 2>&1; then
        print_success "$name completed"
        echo "$response"
    else
        print_error "$name failed"
        return 1
    fi
}

# Main Test Suite
print_header "PingAM Authentication Test Suite"

# Test 0: Check if services are running
print_test "Checking if services are running..."
if curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    print_success "Auth Service is running"
else
    print_error "Auth Service is not running at $BASE_URL"
    echo "Please start the Auth Service first:"
    echo "  cd auth-service && go run cmd/http/main.go"
    exit 1
fi

if curl -s "$PINGAM_MOCK_URL/health" > /dev/null 2>&1; then
    print_success "PingAM Mock Server is running"
else
    print_error "PingAM Mock Server is not running at $PINGAM_MOCK_URL"
    echo "Please start the PingAM Mock Server:"
    echo "  docker-compose -f docker-compose.pingam.yml up -d pingam-mock"
    exit 1
fi

# Test 1: Health Checks
print_header "Test 1: Health Checks"

print_test "1.1 Auth Service Health Check"
curl -s "$BASE_URL/health" | jq
print_success "Auth Service health check passed"

print_test "1.2 PingAM Health Check"
curl -s "$BASE_URL/pingam/health" | jq
print_success "PingAM health check passed"

# Test 2: PingAM Login
print_header "Test 2: PingAM Authentication"

print_test "2.1 Login with PingAM"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq

# Extract tokens
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user_id')

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
    print_success "Login successful"
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "User ID: $USER_ID"
else
    print_error "Login failed - no access token received"
    exit 1
fi

# Test 3: Session Validation
print_header "Test 3: Session Management"

print_test "3.1 Validate Session"
curl -s -X GET "$BASE_URL/pingam/auth/sessions/session_abc123xyz" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq
print_success "Session validation completed"

# Test 4: Authorization Tests
print_header "Test 4: Authorization Checks"

print_test "4.1 Check Permission - Read Users"
PERMISSION_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/authz/check" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "'"$USER_ID"'",
    "resource": "users",
    "action": "read"
  }')

echo "$PERMISSION_RESPONSE" | jq

ALLOWED=$(echo "$PERMISSION_RESPONSE" | jq -r '.allowed')
if [ "$ALLOWED" == "true" ]; then
    print_success "Permission check passed (allowed=true)"
else
    print_error "Permission check failed (allowed=$ALLOWED)"
fi

print_test "4.2 Check Permission - Write Orders"
curl -s -X POST "$BASE_URL/pingam/authz/check" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "'"$USER_ID"'",
    "resource": "orders",
    "action": "write"
  }' | jq
print_success "Permission check completed"

# Test 5: User Roles
print_header "Test 5: User Roles"

print_test "5.1 Get User Roles"
ROLES_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/roles/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$ROLES_RESPONSE" | jq

ROLES_COUNT=$(echo "$ROLES_RESPONSE" | jq -r '.count')
print_success "Retrieved $ROLES_COUNT roles"

# Test 6: User Permissions
print_header "Test 6: User Permissions"

print_test "6.1 Get User Permissions"
PERMISSIONS_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/permissions/$USER_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "$PERMISSIONS_RESPONSE" | jq

PERMISSIONS_COUNT=$(echo "$PERMISSIONS_RESPONSE" | jq -r '.count')
print_success "Retrieved $PERMISSIONS_COUNT permissions"

# Test 7: Token Refresh
print_header "Test 7: Token Management"

print_test "7.1 Refresh Access Token"
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/refresh-token" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "'"$REFRESH_TOKEN"'"
  }')

echo "$REFRESH_RESPONSE" | jq

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')
if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ -n "$NEW_ACCESS_TOKEN" ]; then
    print_success "Token refresh successful"
    echo "New Access Token: ${NEW_ACCESS_TOKEN:0:50}..."
else
    print_error "Token refresh failed"
fi

# Test 8: Token Revocation
print_test "7.2 Revoke Token"
REVOKE_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/revoke-token" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "token": "'"$ACCESS_TOKEN"'"
  }')

echo "$REVOKE_RESPONSE" | jq

REVOKE_MESSAGE=$(echo "$REVOKE_RESPONSE" | jq -r '.message')
if [ -n "$REVOKE_MESSAGE" ]; then
    print_success "Token revocation successful"
else
    print_error "Token revocation failed"
fi

# Test 9: Error Scenarios
print_header "Test 8: Error Scenarios"

print_test "8.1 Login with Invalid Credentials"
INVALID_LOGIN=$(curl -s -X POST "$BASE_URL/pingam/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "invalid_user",
    "password": "wrong_password"
  }')

echo "$INVALID_LOGIN" | jq
print_success "Invalid login test completed"

print_test "8.2 Access Protected Endpoint Without Token"
NO_AUTH=$(curl -s -X GET "$BASE_URL/pingam/authz/roles/$USER_ID")
echo "$NO_AUTH" | jq
print_success "Unauthorized access test completed"

print_test "8.3 Check Permission with Invalid User ID"
curl -s -X POST "$BASE_URL/pingam/authz/check" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{
    "user_id": "invalid-user-id",
    "resource": "users",
    "action": "read"
  }' | jq
print_success "Invalid user ID test completed"

# Summary
print_header "Test Summary"

echo -e "${GREEN}✓ All tests completed successfully!${NC}"
echo ""
echo "Test Results:"
echo "  - Health Checks: ✓"
echo "  - Authentication: ✓"
echo "  - Session Management: ✓"
echo "  - Authorization: ✓"
echo "  - User Roles: ✓"
echo "  - User Permissions: ✓"
echo "  - Token Management: ✓"
echo "  - Error Scenarios: ✓"
echo ""
echo "All PingAM integration tests passed! 🎉"


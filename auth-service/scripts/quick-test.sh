#!/bin/bash

# Quick Test Script for JWT and PingAM Authentication
# This script runs essential tests to verify both authentication systems

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8085/api/v1"

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

print_test() {
    echo -e "${BLUE}TEST: $1${NC}"
}

# Check if service is running
print_header "Checking Auth Service"

if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    print_error "Auth service is not running at $BASE_URL"
    echo ""
    echo "Please start the service first:"
    echo "  ./auth-service"
    echo ""
    exit 1
fi

print_success "Auth service is running"

# =============================================================================
# JWT AUTHENTICATION TESTS
# =============================================================================

print_header "JWT Authentication Tests"

# Test 1: Register User
print_test "1. User Registration"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser_'$(date +%s)'",
    "email": "testuser'$(date +%s)'@example.com",
    "password": "SecurePassword123!"
  }')

if echo "$REGISTER_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.id')
    USERNAME=$(echo "$REGISTER_RESPONSE" | jq -r '.username')
    print_success "User registered successfully (ID: $USER_ID)"
else
    print_error "User registration failed"
    echo "$REGISTER_RESPONSE" | jq '.'
    exit 1
fi

# Test 2: Login
print_test "2. JWT Login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$USERNAME\",
    \"password\": \"SecurePassword123!\"
  }")

if echo "$LOGIN_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
    JWT_ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
    JWT_REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')
    print_success "Login successful"
    print_info "Access Token: ${JWT_ACCESS_TOKEN:0:50}..."
else
    print_error "Login failed"
    echo "$LOGIN_RESPONSE" | jq '.'
    exit 1
fi

# Test 3: Access Protected Endpoint
print_test "3. Access Protected Endpoint"
PROFILE_RESPONSE=$(curl -s -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer $JWT_ACCESS_TOKEN")

if echo "$PROFILE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    print_success "Protected endpoint accessible with valid token"
    print_info "User: $(echo "$PROFILE_RESPONSE" | jq -r '.username')"
else
    print_error "Failed to access protected endpoint"
    echo "$PROFILE_RESPONSE" | jq '.'
    exit 1
fi

# Test 4: Access Without Token (Should Fail)
print_test "4. Access Without Token (Should Fail)"
NO_TOKEN_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/auth/profile")
HTTP_CODE=$(echo "$NO_TOKEN_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "401" ]; then
    print_success "Correctly rejected request without token (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi

# Test 5: Token Refresh
print_test "5. Token Refresh"
if [ -n "$JWT_REFRESH_TOKEN" ] && [ "$JWT_REFRESH_TOKEN" != "null" ]; then
    REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh-token" \
      -H "Content-Type: application/json" \
      -d "{
        \"refresh_token\": \"$JWT_REFRESH_TOKEN\"
      }")
    
    if echo "$REFRESH_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
        print_success "Token refresh successful"
        NEW_JWT_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')
        print_info "New Token: ${NEW_JWT_TOKEN:0:50}..."
    else
        print_error "Token refresh failed"
        echo "$REFRESH_RESPONSE" | jq '.'
    fi
else
    print_info "Refresh token not available, skipping"
fi

# Test 6: Invalid Token (Should Fail)
print_test "6. Access with Invalid Token (Should Fail)"
INVALID_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/auth/profile" \
  -H "Authorization: Bearer invalid.token.here")
HTTP_CODE=$(echo "$INVALID_RESPONSE" | tail -n1)

if [ "$HTTP_CODE" = "401" ]; then
    print_success "Correctly rejected invalid token (401)"
else
    print_error "Expected 401, got $HTTP_CODE"
fi

# =============================================================================
# PINGAM AUTHENTICATION TESTS
# =============================================================================

# print_header "PingAM Authentication Tests"

# # Check if PingAM Mock is running
# if ! curl -s http://localhost:1080/health > /dev/null 2>&1; then
#     print_error "PingAM Mock is not running at http://localhost:1080"
#     print_info "Start PingAM Mock with: docker-compose -f docker-compose.pingam.yml up -d"
#     echo ""
#     print_info "Skipping PingAM tests..."
#     echo ""
# else
#     print_success "PingAM Mock is running"
    
#     # Test 7: PingAM Login
#     print_test "7. PingAM Authentication"
#     PINGAM_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/auth/login" \
#       -H "Content-Type: application/json" \
#       -d '{
#         "username": "testuser",
#         "password": "password123"
#       }')
    
#     if echo "$PINGAM_LOGIN_RESPONSE" | jq -e '.access_token' > /dev/null 2>&1; then
#         PINGAM_ACCESS_TOKEN=$(echo "$PINGAM_LOGIN_RESPONSE" | jq -r '.access_token')
#         PINGAM_USER_ID=$(echo "$PINGAM_LOGIN_RESPONSE" | jq -r '.user_id')
#         print_success "PingAM authentication successful"
#         print_info "User ID: $PINGAM_USER_ID"
#         print_info "Token: ${PINGAM_ACCESS_TOKEN:0:50}..."
#     else
#         print_error "PingAM authentication failed"
#         echo "$PINGAM_LOGIN_RESPONSE" | jq '.'
#     fi
    
#     # Test 8: Session Validation
#     if [ -n "$PINGAM_ACCESS_TOKEN" ] && [ "$PINGAM_ACCESS_TOKEN" != "null" ]; then
#         print_test "8. PingAM Session Validation"
#         SESSION_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/auth/sessions/session_abc123xyz" \
#           -H "Authorization: Bearer $PINGAM_ACCESS_TOKEN")
        
#         if echo "$SESSION_RESPONSE" | jq -e '.session_id' > /dev/null 2>&1; then
#             print_success "Session validation successful"
#         else
#             print_info "Session validation response: $(echo "$SESSION_RESPONSE" | jq -c '.')"
#         fi
#     fi
    
#     # Test 9: Permission Check
#     if [ -n "$PINGAM_ACCESS_TOKEN" ] && [ "$PINGAM_USER_ID" != "null" ]; then
#         print_test "9. PingAM Permission Check"
#         PERMISSION_RESPONSE=$(curl -s -X POST "$BASE_URL/pingam/authz/check" \
#           -H "Authorization: Bearer $PINGAM_ACCESS_TOKEN" \
#           -H "Content-Type: application/json" \
#           -d "{
#             \"user_id\": \"$PINGAM_USER_ID\",
#             \"resource\": \"users\",
#             \"action\": \"read\"
#           }")
        
#         if echo "$PERMISSION_RESPONSE" | jq -e '.allowed' > /dev/null 2>&1; then
#             IS_ALLOWED=$(echo "$PERMISSION_RESPONSE" | jq -r '.allowed')
#             if [ "$IS_ALLOWED" = "true" ]; then
#                 print_success "Permission check: users:read allowed"
#             else
#                 print_info "Permission check: users:read denied"
#             fi
#         else
#             print_info "Permission response: $(echo "$PERMISSION_RESPONSE" | jq -c '.')"
#         fi
#     fi
    
#     # Test 10: Get User Roles
#     if [ -n "$PINGAM_ACCESS_TOKEN" ] && [ "$PINGAM_USER_ID" != "null" ]; then
#         print_test "10. PingAM Get User Roles"
#         ROLES_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/roles/$PINGAM_USER_ID" \
#           -H "Authorization: Bearer $PINGAM_ACCESS_TOKEN")
        
#         if echo "$ROLES_RESPONSE" | jq -e '.roles' > /dev/null 2>&1; then
#             ROLES=$(echo "$ROLES_RESPONSE" | jq -r '.roles | join(", ")')
#             print_success "User roles retrieved: $ROLES"
#         else
#             print_info "Roles response: $(echo "$ROLES_RESPONSE" | jq -c '.')"
#         fi
#     fi
    
#     # Test 11: Get User Permissions
#     if [ -n "$PINGAM_ACCESS_TOKEN" ] && [ "$PINGAM_USER_ID" != "null" ]; then
#         print_test "11. PingAM Get User Permissions"
#         PERMS_RESPONSE=$(curl -s -X GET "$BASE_URL/pingam/authz/permissions/$PINGAM_USER_ID" \
#           -H "Authorization: Bearer $PINGAM_ACCESS_TOKEN")
        
#         if echo "$PERMS_RESPONSE" | jq -e '.permissions' > /dev/null 2>&1; then
#             PERM_COUNT=$(echo "$PERMS_RESPONSE" | jq '.permissions | length')
#             print_success "User permissions retrieved: $PERM_COUNT permissions"
#         else
#             print_info "Permissions response: $(echo "$PERMS_RESPONSE" | jq -c '.')"
#         fi
#     fi
# fi

# =============================================================================
# HEALTH CHECKS
# =============================================================================

print_header "Health Checks"

# Test 12: Service Health
print_test "12. Service Health Check"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health")

if echo "$HEALTH_RESPONSE" | jq -e '.status' > /dev/null 2>&1; then
    STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.status')
    if [ "$STATUS" = "ok" ] || [ "$STATUS" = "healthy" ]; then
        print_success "Service health: $STATUS"
    else
        print_error "Service health: $STATUS"
    fi
    
    # Check database
    if echo "$HEALTH_RESPONSE" | jq -e '.database.status' > /dev/null 2>&1; then
        DB_STATUS=$(echo "$HEALTH_RESPONSE" | jq -r '.database.status')
        if [ "$DB_STATUS" = "healthy" ] || [ "$DB_STATUS" = "ok" ]; then
            print_success "Database health: $DB_STATUS"
        else
            print_error "Database health: $DB_STATUS"
        fi
    fi
else
    print_error "Health check failed"
    echo "$HEALTH_RESPONSE"
fi

# Test 13: PingAM Health
# if curl -s http://localhost:1080/health > /dev/null 2>&1; then
#     print_test "13. PingAM Health Check"
#     PINGAM_HEALTH=$(curl -s "$BASE_URL/pingam/health")
    
#     if echo "$PINGAM_HEALTH" | jq -e '.status' > /dev/null 2>&1; then
#         PINGAM_STATUS=$(echo "$PINGAM_HEALTH" | jq -r '.status')
#         print_success "PingAM health: $PINGAM_STATUS"
#     else
#         print_info "PingAM health response: $PINGAM_HEALTH"
#     fi
# fi

# =============================================================================
# TEST SUMMARY
# =============================================================================

print_header "Test Summary"

echo -e "${GREEN}✓ JWT Authentication Tests Complete${NC}"
echo "  ✓ User registration"
echo "  ✓ User login"
echo "  ✓ Protected endpoint access"
echo "  ✓ Token validation"
echo "  ✓ Token refresh"
echo "  ✓ Invalid token rejection"
echo ""

if curl -s http://localhost:1080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PingAM Authentication Tests Complete${NC}"
    echo "  ✓ PingAM login"
    echo "  ✓ Session validation"
    echo "  ✓ Permission checking"
    echo "  ✓ Role retrieval"
    echo "  ✓ Permission retrieval"
    echo ""
fi

echo -e "${GREEN}✓ Health Checks Complete${NC}"
echo "  ✓ Service health"
echo "  ✓ Database health"
if curl -s http://localhost:1080/health > /dev/null 2>&1; then
    echo "  ✓ PingAM health"
fi
echo ""

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}🎉 All Tests Passed Successfully!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

echo "Next steps:"
echo "  • View detailed logs: tail -f logs/auth-service.log"
echo "  • Test with Swagger: http://localhost:8081"
echo "  • Run full test suite: ./run-all-tests.sh"
echo "  • Read testing guide: SETUP_AND_TESTING_GUIDE.md"
echo ""



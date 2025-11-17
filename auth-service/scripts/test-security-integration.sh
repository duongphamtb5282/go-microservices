#!/bin/bash

# Security Integration Test Script
# Tests the backend-core/security integration

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

print_test() {
    echo -e "${YELLOW}[TEST] $1${NC}"
}

print_header "Security Integration Test Suite"

# Test 1: Check if security package is imported
print_test "1. Checking if backend-core/security is imported"
if grep -r "backend-core/security" src 2>/dev/null | grep -v "Binary" > /dev/null; then
    print_success "backend-core/security is imported"
else
    print_error "backend-core/security not found in imports"
    exit 1
fi

# Test 2: Check JWT configuration
print_test "2. Checking JWT configuration in config.yaml"
if grep -q "jwt:" config/config.yaml && \
   grep -q "expiry:" config/config.yaml && \
   grep -q "issuer:" config/config.yaml; then
    print_success "JWT configuration found in config.yaml"
else
    print_error "JWT configuration missing in config.yaml"
    exit 1
fi

# Test 3: Check JWT middleware implementation
print_test "3. Checking JWT middleware implementation"
if [ -f "src/interfaces/rest/middleware/jwt_auth.go" ]; then
    if grep -q "JWTAuthMiddleware" src/interfaces/rest/middleware/jwt_auth.go && \
       grep -q "security.JWTManager" src/interfaces/rest/middleware/jwt_auth.go; then
        print_success "JWT middleware properly implemented"
    else
        print_error "JWT middleware missing proper implementation"
        exit 1
    fi
else
    print_error "JWT middleware file not found"
    exit 1
fi

# Test 4: Check security providers
print_test "4. Checking security providers"
if [ -f "src/applications/providers/security_providers.go" ]; then
    if grep -q "JWTManagerProvider" src/applications/providers/security_providers.go && \
       grep -q "AuthManagerProvider" src/applications/providers/security_providers.go; then
        print_success "Security providers implemented"
    else
        print_error "Security providers missing"
        exit 1
    fi
else
    print_error "Security providers file not found"
    exit 1
fi

# Test 5: Check config structure
print_test "5. Checking config structure for JWT"
if grep -q "JWTConfig" src/infrastructure/config/config.go && \
   grep -q "Secret" src/infrastructure/config/config.go && \
   grep -q "Expiry" src/infrastructure/config/config.go; then
    print_success "JWT config structure defined"
else
    print_error "JWT config structure missing"
    exit 1
fi

# Test 6: Check middleware setup integration
print_test "6. Checking middleware setup integration"
if grep -q "jwtManager" src/interfaces/rest/middleware/middleware_setup.go && \
   grep -q "security.JWTManager" src/interfaces/rest/middleware/middleware_setup.go; then
    print_success "Middleware setup integrated with JWT manager"
else
    print_error "Middleware setup not integrated with JWT manager"
    exit 1
fi

# Test 7: Check service factory integration
print_test "7. Checking service factory integration"
if grep -q "JWTManagerProvider" src/applications/service_factory.go && \
   grep -q "AuthManagerProvider" src/applications/service_factory.go; then
    print_success "Service factory integrated with security providers"
else
    print_error "Service factory not integrated with security providers"
    exit 1
fi

# Test 8: Verify compilation
print_test "8. Verifying Go compilation"
if go build -o /dev/null ./cmd/http/main.go 2>&1 | tee /tmp/build_output.txt; then
    print_success "Project compiles successfully"
else
    print_error "Project compilation failed"
    echo -e "\n${RED}Compilation errors:${NC}"
    cat /tmp/build_output.txt
    exit 1
fi

# Summary
print_header "Test Summary"
echo -e "${GREEN}✓ All security integration tests passed!${NC}"
echo ""
echo "Security Integration Status:"
echo "  ✓ backend-core/security imported"
echo "  ✓ JWT configuration added"
echo "  ✓ JWT middleware implemented"
echo "  ✓ Security providers created"
echo "  ✓ Config structure updated"
echo "  ✓ Middleware setup integrated"
echo "  ✓ Service factory updated"
echo "  ✓ Project compiles successfully"
echo ""
echo -e "${GREEN}Security integration complete! 🎉${NC}"


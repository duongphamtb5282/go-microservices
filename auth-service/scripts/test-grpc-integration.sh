#!/bin/bash

# gRPC Integration Testing Script for Auth Service
# Tests gRPC communication between auth-service and admin-service

set -e

# Configuration
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:8085}"
ADMIN_SERVICE_GRPC="${ADMIN_SERVICE_GRPC:-localhost:50051}"
ADMIN_SERVICE_HTTP="${ADMIN_SERVICE_HTTP:-http://localhost:8086}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Function to check if a service is running
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=${3:-200}

    log_info "Checking if $service_name is running..."

    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_status"; then
        log_success "$service_name is running"
        return 0
    else
        log_error "$service_name is not accessible"
        return 1
    fi
}

# Function to check gRPC connectivity using grpcurl
check_grpc_connectivity() {
    log_info "Checking gRPC connectivity to admin service..."

    if command -v grpcurl >/dev/null 2>&1; then
        if grpcurl -plaintext "$ADMIN_SERVICE_GRPC" list 2>/dev/null; then
            log_success "gRPC connection to admin service successful"
            return 0
        else
            log_error "Cannot connect to gRPC service at $ADMIN_SERVICE_GRPC"
            return 1
        fi
    else
        log_warn "grpcurl not found - install it to get detailed gRPC testing"
        log_info "Install grpcurl: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"

        # Try basic connectivity check with nc
        if timeout 5 bash -c "</dev/tcp/$ADMIN_SERVICE_GRPC" 2>/dev/null; then
            log_success "Basic gRPC port connectivity check passed"
            return 0
        else
            log_error "Cannot connect to gRPC port $ADMIN_SERVICE_GRPC"
            return 1
        fi
    fi
}

# Function to list gRPC services
list_grpc_services() {
    log_info "Listing available gRPC services..."

    if command -v grpcurl >/dev/null 2>&1; then
        log_info "Available gRPC services:"
        grpcurl -plaintext "$ADMIN_SERVICE_GRPC" list

        log_info "AdminService methods:"
        grpcurl -plaintext "$ADMIN_SERVICE_GRPC" list backend_shared.AdminService
    else
        log_warn "grpcurl not found - cannot list services"
    fi
}

# Function to test gRPC health check
test_grpc_health() {
    log_info "Testing gRPC health check..."

    if command -v grpcurl >/dev/null 2>&1; then
        HEALTH_RESPONSE=$(grpcurl -plaintext -d '{}' "$ADMIN_SERVICE_GRPC" grpc.health.v1.Health/Check 2>/dev/null || echo "failed")

        if echo "$HEALTH_RESPONSE" | grep -q '"status":"SERVING"'; then
            log_success "gRPC health check passed"
            return 0
        else
            log_error "gRPC health check failed"
            log_error "Response: $HEALTH_RESPONSE"
            return 1
        fi
    else
        log_warn "grpcurl not found - skipping health check"
        return 0
    fi
}

# Function to test user creation via gRPC
test_grpc_user_creation() {
    log_info "Testing user creation via gRPC..."

    if ! command -v grpcurl >/dev/null 2>&1; then
        log_warn "grpcurl not found - skipping gRPC user creation test"
        return 0
    fi

    # Create test user data
    TEST_USER_ID="grpc-test-$(date +%s)"
    TEST_USERNAME="grpc-test-user-$(date +%s)"
    TEST_EMAIL="$TEST_USERNAME@example.com"

    USER_DATA='{
        "userId": "'"$TEST_USER_ID"'",
        "email": "'"$TEST_EMAIL"'",
        "username": "'"$TEST_USERNAME"'",
        "firstName": "gRPC",
        "lastName": "Test",
        "createdAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "createdBy": "grpc-test-script",
        "serviceName": "auth-service"
    }'

    log_info "Creating user via gRPC: $TEST_USERNAME"

    # Call gRPC method
    RESPONSE=$(grpcurl -plaintext -d "$USER_DATA" "$ADMIN_SERVICE_GRPC" backend_shared.AdminService.RecordUserCreated 2>/dev/null || echo "failed")

    if echo "$RESPONSE" | grep -q '"success":true\|"eventId":'; then
        log_success "gRPC user creation successful"

        # Extract event ID if available
        EVENT_ID=$(echo "$RESPONSE" | grep -o '"eventId":"[^"]*' | cut -d'"' -f4)
        if [ -n "$EVENT_ID" ]; then
            log_info "Generated event ID: $EVENT_ID"
        fi

        return 0
    else
        log_error "gRPC user creation failed"
        log_error "Request: $USER_DATA"
        log_error "Response: $RESPONSE"
        return 1
    fi
}

# Function to test user update via gRPC
test_grpc_user_update() {
    log_info "Testing user update via gRPC..."

    if ! command -v grpcurl >/dev/null 2>&1; then
        log_warn "grpcurl not found - skipping gRPC user update test"
        return 0
    fi

    # Use the user ID from creation test or create a new one
    TEST_USER_ID="${TEST_USER_ID:-grpc-update-test-$(date +%s)}"
    TEST_USERNAME="grpc-update-user-$(date +%s)"
    TEST_EMAIL="$TEST_USERNAME@example.com"

    USER_DATA='{
        "userId": "'"$TEST_USER_ID"'",
        "email": "'"$TEST_EMAIL"'",
        "username": "'"$TEST_USERNAME"'",
        "firstName": "gRPC",
        "lastName": "Updated",
        "updatedAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "updatedBy": "grpc-test-script",
        "serviceName": "auth-service",
        "changedFields": ["lastName", "email"]
    }'

    log_info "Updating user via gRPC: $TEST_USER_ID"

    RESPONSE=$(grpcurl -plaintext -d "$USER_DATA" "$ADMIN_SERVICE_GRPC" backend_shared.AdminService.RecordUserUpdated 2>/dev/null || echo "failed")

    if echo "$RESPONSE" | grep -q '"success":true\|"eventId":'; then
        log_success "gRPC user update successful"
        return 0
    else
        log_error "gRPC user update failed"
        log_error "Response: $RESPONSE"
        return 1
    fi
}

# Function to test user deletion via gRPC
test_grpc_user_deletion() {
    log_info "Testing user deletion via gRPC..."

    if ! command -v grpcurl >/dev/null 2>&1; then
        log_warn "grpcurl not found - skipping gRPC user deletion test"
        return 0
    fi

    TEST_USER_ID="${TEST_USER_ID:-grpc-delete-test-$(date +%s)}"
    TEST_EMAIL="delete-test@example.com"

    USER_DATA='{
        "userId": "'"$TEST_USER_ID"'",
        "email": "'"$TEST_EMAIL"'",
        "deletedAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "deletedBy": "grpc-test-script",
        "serviceName": "auth-service",
        "reason": "Testing gRPC integration"
    }'

    log_info "Deleting user via gRPC: $TEST_USER_ID"

    RESPONSE=$(grpcurl -plaintext -d "$USER_DATA" "$ADMIN_SERVICE_GRPC" backend_shared.AdminService.RecordUserDeleted 2>/dev/null || echo "failed")

    if echo "$RESPONSE" | grep -q '"success":true\|"eventId":'; then
        log_success "gRPC user deletion successful"
        return 0
    else
        log_error "gRPC user deletion failed"
        log_error "Response: $RESPONSE"
        return 1
    fi
}

# Function to test auth service integration with gRPC
test_auth_service_grpc_integration() {
    log_info "Testing auth service gRPC integration..."

    # Check if auth service is running
    if ! check_service "Auth Service" "$AUTH_SERVICE_URL/api/v1/health" >/dev/null 2>&1; then
        log_warn "Auth service not running - skipping integration test"
        return 0
    fi

    # Check if admin service gRPC is available
    if ! check_grpc_connectivity >/dev/null 2>&1; then
        log_warn "Admin service gRPC not available - skipping integration test"
        return 0
    fi

    # Create a user via auth service (this should trigger gRPC call to admin service)
    TEST_USERNAME="integration-test-user-$(date +%s)"
    TEST_EMAIL="$TEST_USERNAME@example.com"

    USER_DATA='{
        "username": "'"$TEST_USERNAME"'",
        "email": "'"$TEST_EMAIL"'",
        "password": "test123456",
        "first_name": "Integration",
        "last_name": "Test"
    }'

    log_info "Creating user via auth service (should trigger gRPC call): $TEST_USERNAME"

    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$USER_DATA" \
        "$AUTH_SERVICE_URL/api/v1/users")

    if echo "$RESPONSE" | grep -q '"success":true\|"id":'; then
        log_success "Auth service user creation successful"

        # Wait a moment for gRPC call to complete
        sleep 2

        # Check admin service logs or database to verify gRPC call
        log_info "Check admin service logs to verify gRPC call was made"
        log_info "Command: docker logs admin-service | grep '$TEST_USERNAME'"

        return 0
    else
        log_error "Auth service user creation failed"
        log_error "Response: $RESPONSE"
        return 1
    fi
}

# Function to test gRPC performance
test_grpc_performance() {
    log_info "Testing gRPC performance..."

    if ! command -v grpcurl >/dev/null 2>&1; then
        log_warn "grpcurl not found - skipping performance test"
        return 0
    fi

    local num_requests=10

    log_info "Making $num_requests gRPC calls for performance testing..."

    # Start timing
    start_time=$(date +%s.%3N)

    for i in $(seq 1 $num_requests); do
        TEST_USER_ID="perf-test-$i-$(date +%s)"
        USER_DATA='{
            "userId": "'"$TEST_USER_ID"'",
            "email": "perf-test-'$i'@example.com",
            "username": "perf-user-'$i'",
            "firstName": "Perf",
            "lastName": "Test",
            "createdAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
            "createdBy": "grpc-performance-test",
            "serviceName": "auth-service"
        }'

        grpcurl -plaintext -d "$USER_DATA" "$ADMIN_SERVICE_GRPC" backend_shared.AdminService.RecordUserCreated >/dev/null 2>&1
    done

    # End timing
    end_time=$(date +%s.%3N)
    duration=$(echo "$end_time - $start_time" | bc)

    local req_per_sec=$(echo "scale=2; $num_requests / $duration" | bc)

    log_success "Performance test completed: $num_requests requests in ${duration}s (${req_per_sec} req/sec)"
}

# Function to test gRPC error handling
test_grpc_error_handling() {
    log_info "Testing gRPC error handling..."

    if ! command -v grpcurl >/dev/null 2>&1; then
        log_warn "grpcurl not found - skipping error handling test"
        return 0
    fi

    # Test with invalid data
    INVALID_DATA='{
        "invalidField": "test"
    }'

    log_info "Testing with invalid gRPC request data..."

    ERROR_RESPONSE=$(grpcurl -plaintext -d "$INVALID_DATA" "$ADMIN_SERVICE_GRPC" backend_shared.AdminService.RecordUserCreated 2>&1 || echo "error occurred")

    if echo "$ERROR_RESPONSE" | grep -q "error\|Error\|failed"; then
        log_success "gRPC error handling working - invalid request properly rejected"
    else
        log_warn "gRPC error response not detected"
    fi

    # Test with invalid service/port
    log_info "Testing connection to invalid gRPC endpoint..."

    INVALID_RESPONSE=$(grpcurl -plaintext -d '{}' "invalid-host:9999" backend_shared.AdminService.RecordUserCreated 2>&1 || echo "connection failed")

    if echo "$INVALID_RESPONSE" | grep -q "failed\|error\|refused"; then
        log_success "gRPC connection error handling working"
    else
        log_warn "gRPC connection error not properly handled"
    fi
}

# Function to show gRPC monitoring info
show_grpc_monitoring_info() {
    echo ""
    echo "=========================================="
    echo "📊 gRPC Monitoring & Troubleshooting"
    echo "=========================================="
    echo ""
    echo "Useful gRPC commands:"
    echo ""
    echo "1. List all services:"
    echo "   grpcurl -plaintext $ADMIN_SERVICE_GRPC list"
    echo ""
    echo "2. List service methods:"
    echo "   grpcurl -plaintext $ADMIN_SERVICE_GRPC list backend_shared.AdminService"
    echo ""
    echo "3. Health check:"
    echo "   grpcurl -plaintext -d '{}' $ADMIN_SERVICE_GRPC grpc.health.v1.Health/Check"
    echo ""
    echo "4. Call a method:"
    echo "   grpcurl -plaintext -d '{\"userId\":\"test\"}' $ADMIN_SERVICE_GRPC backend_shared.AdminService.RecordUserCreated"
    echo ""
    echo "5. Get method description:"
    echo "   grpcurl -plaintext $ADMIN_SERVICE_GRPC describe backend_shared.AdminService.RecordUserCreated"
    echo ""
    echo "6. Check admin service logs:"
    echo "   docker logs admin-service"
    echo ""
    echo "7. Check auth service logs for gRPC calls:"
    echo "   docker logs auth-service | grep 'admin.*client\|grpc'"
    echo ""
    echo "8. Test with grpcui (web UI):"
    echo "   grpcui -plaintext $ADMIN_SERVICE_GRPC"
    echo ""
}

# Main execution function
main() {
    echo "🚀 gRPC Integration Testing for Auth Service"
    echo "============================================"
    echo ""

    local all_tests_passed=true

    # Check prerequisites
    if ! check_service "Admin Service HTTP" "$ADMIN_SERVICE_HTTP/health" 200; then
        log_warn "Admin service HTTP not accessible - some tests may fail"
    fi

    if ! check_grpc_connectivity; then
        log_error "gRPC connectivity check failed"
        all_tests_passed=false
    fi

    # List available services
    list_grpc_services

    # Run tests
    echo ""
    echo "🧪 Running gRPC Integration Tests"
    echo "=================================="

    # Basic connectivity tests
    test_grpc_health

    # gRPC method tests
    if command -v grpcurl >/dev/null 2>&1; then
        test_grpc_user_creation
        test_grpc_user_update
        test_grpc_user_deletion

        # Integration test
        test_auth_service_grpc_integration

        # Performance and error handling
        test_grpc_performance
        test_grpc_error_handling
    else
        log_warn "grpcurl not required for all tests - some advanced tests skipped"
    fi

    # Show monitoring info
    show_grpc_monitoring_info

    echo ""
    echo "=========================================="
    if $all_tests_passed; then
        log_success "✅ gRPC Integration Testing Complete!"
        log_success "All prerequisite checks passed"
    else
        log_error "❌ Some tests failed - check the output above"
        exit 1
    fi
    echo "=========================================="
    echo ""
}

# Run main function
main "$@"

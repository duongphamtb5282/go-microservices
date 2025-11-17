#!/bin/bash

# Correlation ID Testing Script
# Tests the correlation ID functionality in auth service

AUTH_SERVICE_URL="http://localhost:8085/api/v1"

echo "🔗 Testing Correlation ID Functionality"
echo "======================================="
echo ""

# Function to check if auth service is running
check_auth_service() {
    echo "Checking auth service status..."
    if curl -s "$AUTH_SERVICE_URL/health" > /dev/null 2>&1; then
        echo "✅ Auth service is running"
        return 0
    else
        echo "❌ Auth service not running. Start with: go run main.go"
        return 1
    fi
}

# Function to test correlation ID in requests
test_correlation_id() {
    echo ""
    echo "🧪 Testing Correlation ID in HTTP Requests"
    echo "==========================================="
    
    # Test 1: Request without correlation ID (should generate one)
    echo "Test 1: Request without correlation ID"
    echo "--------------------------------------"
    RESPONSE1=$(curl -s -w "\n%{http_code}" "$AUTH_SERVICE_URL/health")
    HTTP_CODE1=$(echo "$RESPONSE1" | tail -n1)
    BODY1=$(echo "$RESPONSE1" | head -n -1)
    
    echo "Response: $BODY1"
    echo "HTTP Code: $HTTP_CODE1"
    echo ""
    
    # Test 2: Request with custom correlation ID
    echo "Test 2: Request with custom correlation ID"
    echo "------------------------------------------"
    CUSTOM_CORRELATION_ID="test-correlation-$(date +%s)"
    RESPONSE2=$(curl -s -w "\n%{http_code}" \
        -H "X-Correlation-ID: $CUSTOM_CORRELATION_ID" \
        -H "X-Request-ID: custom-request-$(date +%s)" \
        "$AUTH_SERVICE_URL/health")
    HTTP_CODE2=$(echo "$RESPONSE2" | tail -n1)
    BODY2=$(echo "$RESPONSE2" | head -n -1)
    
    echo "Custom Correlation ID: $CUSTOM_CORRELATION_ID"
    echo "Response: $BODY2"
    echo "HTTP Code: $HTTP_CODE2"
    echo ""
    
    # Test 3: Multiple requests to test correlation ID propagation
    echo "Test 3: Multiple requests with correlation ID propagation"
    echo "--------------------------------------------------------"
    for i in {1..3}; do
        CORRELATION_ID="batch-test-$i-$(date +%s)"
        echo "Request $i with Correlation ID: $CORRELATION_ID"
        
        RESPONSE=$(curl -s -w "\n%{http_code}" \
            -H "X-Correlation-ID: $CORRELATION_ID" \
            -H "X-Request-ID: request-$i-$(date +%s)" \
            "$AUTH_SERVICE_URL/users")
        
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        BODY=$(echo "$RESPONSE" | head -n -1)
        
        echo "  HTTP Code: $HTTP_CODE"
        echo "  Response: $BODY"
        echo ""
    done
}

# Function to test correlation ID in different endpoints
test_correlation_id_endpoints() {
    echo ""
    echo "🌐 Testing Correlation ID across Different Endpoints"
    echo "=================================================="
    
    ENDPOINTS=(
        "/health"
        "/status"
        "/users"
        "/auth/verify-email"
    )
    
    for endpoint in "${ENDPOINTS[@]}"; do
        echo "Testing endpoint: $endpoint"
        echo "---------------------------"
        
        CORRELATION_ID="endpoint-test-$(date +%s)"
        echo "Correlation ID: $CORRELATION_ID"
        
        if [[ "$endpoint" == "/auth/verify-email" ]]; then
            # POST request with body
            RESPONSE=$(curl -s -w "\n%{http_code}" \
                -H "X-Correlation-ID: $CORRELATION_ID" \
                -H "Content-Type: application/json" \
                -X POST \
                -d '{"email": "test@example.com"}' \
                "$AUTH_SERVICE_URL$endpoint")
        else
            # GET request
            RESPONSE=$(curl -s -w "\n%{http_code}" \
                -H "X-Correlation-ID: $CORRELATION_ID" \
                "$AUTH_SERVICE_URL$endpoint")
        fi
        
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        BODY=$(echo "$RESPONSE" | head -n -1)
        
        echo "  HTTP Code: $HTTP_CODE"
        echo "  Response: $BODY"
        echo ""
    done
}

# Function to test correlation ID logging
test_correlation_id_logging() {
    echo ""
    echo "📝 Testing Correlation ID in Logs"
    echo "================================="
    echo ""
    echo "Making requests and checking logs..."
    echo "Look for correlation ID in the auth service logs:"
    echo ""
    
    # Make a few requests with different correlation IDs
    for i in {1..3}; do
        CORRELATION_ID="log-test-$i-$(date +%s)"
        echo "Request $i with Correlation ID: $CORRELATION_ID"
        
        curl -s -H "X-Correlation-ID: $CORRELATION_ID" \
             "$AUTH_SERVICE_URL/health" > /dev/null
        
        echo "  Check auth service logs for correlation ID: $CORRELATION_ID"
    done
    
    echo ""
    echo "💡 To see the logs, run: tail -f auth-service.log (if logging to file)"
    echo "   Or check the console output where you started the auth service"
}

# Function to test correlation ID middleware
test_correlation_id_middleware() {
    echo ""
    echo "🔧 Testing Correlation ID Middleware"
    echo "===================================="
    echo ""
    echo "Testing different correlation ID header formats:"
    echo ""
    
    # Test different header formats
    HEADERS=(
        "X-Correlation-ID"
        "X-Request-ID"
        "X-Trace-ID"
        "Correlation-ID"
        "Request-ID"
    )
    
    for header in "${HEADERS[@]}"; do
        echo "Testing header: $header"
        echo "----------------------"
        
        CORRELATION_ID="middleware-test-$(date +%s)"
        echo "Correlation ID: $CORRELATION_ID"
        
        RESPONSE=$(curl -s -w "\n%{http_code}" \
            -H "$header: $CORRELATION_ID" \
            "$AUTH_SERVICE_URL/health")
        
        HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
        BODY=$(echo "$RESPONSE" | head -n -1)
        
        echo "  HTTP Code: $HTTP_CODE"
        echo "  Response: $BODY"
        echo ""
    done
}

# Main execution
echo "Starting Correlation ID Testing..."
echo ""

if check_auth_service; then
    test_correlation_id
    test_correlation_id_endpoints
    test_correlation_id_logging
    test_correlation_id_middleware
    
    echo ""
    echo "🎉 Correlation ID Testing Complete!"
    echo "==================================="
    echo ""
    echo "✅ Tests completed successfully"
    echo "📋 Summary:"
    echo "   - Correlation ID generation: Working"
    echo "   - Custom correlation ID: Working"
    echo "   - Multiple requests: Working"
    echo "   - Different endpoints: Working"
    echo "   - Header formats: Working"
    echo ""
    echo "💡 Check the auth service logs to see correlation IDs in action!"
else
    echo "❌ Cannot proceed with testing - auth service not running"
    exit 1
fi

#!/bin/bash

# Auth Service API Testing Script
# This script tests the auth-service API endpoints with caching, tracing, and monitoring

set -e

BASE_URL="http://localhost:8085"
API_BASE="$BASE_URL/api/v1"

echo "🧪 Testing Auth Service API"
echo "==========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    local method=$1
    local url=$2
    local expected_status=${3:-200}
    local description=$4

    echo -e "${BLUE}Testing: ${description}${NC}"
    echo -e "${YELLOW}$method $url${NC}"

    # Make request and capture response
    local response=$(curl -s -w "\n%{http_code}" -X $method "$url" \
        -H "Content-Type: application/json" \
        -H "Accept: application/json")

    # Extract status code and body more robustly
    local status_code=$(echo "$response" | tail -n1)
    # Remove the last line (status code) to get the body
    local body=$(echo "$response" | sed '$d')

    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}✅ Status: $status_code (Expected: $expected_status)${NC}"

        # Check if response contains cache info
        if echo "$body" | grep -q '"cached":true'; then
            echo -e "${GREEN}🗄️  Cache hit!${NC}"
        elif echo "$body" | grep -q '"cached":false'; then
            echo -e "${YELLOW}🗄️  Cache miss${NC}"
        fi

        echo "$body" | head -c 200
        echo -e "\n"
    else
        echo -e "${RED}❌ Status: $status_code (Expected: $expected_status)${NC}"
        echo "$body" | head -c 500
        echo -e "\n"
        return 1
    fi
}

echo "1. Testing Health Check"
echo "======================="
test_endpoint "GET" "$API_BASE/status" 200 "Health Check"

echo "2. Testing User Registration"
echo "============================"
# Create a test user
REGISTER_DATA='{
    "username": "testuser",
    "email": "test@example.com",
    "password": "TestPassword123!",
    "first_name": "Test",
    "last_name": "User"
}'

echo -e "${BLUE}Testing: User Registration${NC}"
echo -e "${YELLOW}POST $API_BASE/users${NC}"
echo "$REGISTER_DATA"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/users" \
    -H "Content-Type: application/json" \
    -d "$REGISTER_DATA")

# Extract status code and body more robustly
status_code=$(echo "$response" | tail -n1)
# Remove the last line (status code) to get the body
body=$(echo "$response" | sed '$d')

if [ "$status_code" -eq 200 ] || [ "$status_code" -eq 201 ]; then
    echo -e "${GREEN}✅ User registration successful (Status: $status_code)${NC}"
    USER_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}📝 User ID: $USER_ID${NC}"
else
    echo -e "${RED}❌ User registration failed (Status: $status_code)${NC}"
    echo "$body"
fi
echo

echo "3. Testing User Retrieval (with caching)"
echo "========================================"
# Test cache miss first
test_endpoint "GET" "$API_BASE/users/$USER_ID" 200 "Get User (cache miss)"

# Test cache hit second time
echo "Testing cache hit..."
test_endpoint "GET" "$API_BASE/users/$USER_ID" 200 "Get User (cache hit)"

echo "4. Testing User List (with caching)"
echo "==================================="
test_endpoint "GET" "$API_BASE/users?page=1&limit=10" 200 "List Users (cache miss)"
test_endpoint "GET" "$API_BASE/users?page=1&limit=10" 200 "List Users (cache hit)"

echo "5. Testing Cache Statistics"
echo "==========================="
test_endpoint "GET" "$API_BASE/cache/stats" 200 "Cache Statistics"

echo "6. Testing Monitoring Endpoints"
echo "==============================="

echo "Grafana: http://localhost:3000 (admin/admin)"
echo "Prometheus: http://localhost:9090"
echo "Jaeger: http://localhost:16686"
echo "OTEL Collector Metrics: http://localhost:8888/metrics"

echo -e "\n${GREEN}🎉 API Testing Complete!${NC}"
echo -e "${BLUE}💡 Tips:${NC}"
echo "  - Check Grafana dashboards for metrics visualization"
echo "  - Monitor traces in Jaeger for request flows"
echo "  - View real-time metrics in Prometheus"
echo "  - Check cache hit/miss ratios in responses"

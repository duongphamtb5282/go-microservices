#!/bin/bash

# Telemetry Testing Script for Auth-Service
# This script tests OpenTelemetry tracing, metrics, and business metrics

set -e

# Colors
GREEN='\033[0.;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AUTH_SERVICE_URL="http://localhost:8085"
JAEGER_UI_URL="http://localhost:16686"
PROMETHEUS_URL="http://localhost:9090"
METRICS_ENDPOINT="$AUTH_SERVICE_URL/metrics"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Telemetry Testing - Auth Service${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Test 1: Check if auth-service is running
echo -e "${YELLOW}[1] Checking if auth-service is running...${NC}"
if curl -s -f "$AUTH_SERVICE_URL/api/v1/health" > /dev/null; then
    echo -e "${GREEN}✓ Auth-service is running${NC}"
else
    echo -e "${RED}✗ Auth-service is not running${NC}"
    echo -e "${YELLOW}Start the service with: ./auth-service${NC}"
    exit 1
fi
echo ""

# Test 2: Check if Jaeger is running (optional)
echo -e "${YELLOW}[2] Checking if Jaeger is running...${NC}"
if curl -s -f "$JAEGER_UI_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Jaeger is running at $JAEGER_UI_URL${NC}"
    JAEGER_RUNNING=true
else
    echo -e "${YELLOW}⚠ Jaeger is not running (optional for telemetry testing)${NC}"
    echo -e "${YELLOW}To start: docker run -d -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:1.50${NC}"
    JAEGER_RUNNING=false
fi
echo ""

# Test 3: Check if Prometheus metrics endpoint is available
echo -e "${YELLOW}[3] Checking metrics endpoint...${NC}"
if curl -s -f "$METRICS_ENDPOINT" > /dev/null; then
    echo -e "${GREEN}✓ Metrics endpoint is available at $METRICS_ENDPOINT${NC}"
    
    # Count metrics
    METRIC_COUNT=$(curl -s "$METRICS_ENDPOINT" | grep -v "^#" | wc -l | tr -d ' ')
    echo -e "${GREEN}  Found $METRIC_COUNT metrics${NC}"
else
    echo -e "${RED}✗ Metrics endpoint is not available${NC}"
    exit 1
fi
echo ""

# Test 4: Generate telemetry data - User Registration
echo -e "${YELLOW}[4] Testing telemetry with user registration...${NC}"
TIMESTAMP=$(date +%s)
TEST_USER="telemetry_test_$TIMESTAMP"
TEST_EMAIL="telemetry_test_$TIMESTAMP@example.com"

REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_SERVICE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"$TEST_USER\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"TelemetryTest123!\"
  }")

REGISTER_STATUS=$(echo "$REGISTER_RESPONSE" | tail -n1)
REGISTER_BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')

if [ "$REGISTER_STATUS" = "201" ]; then
    echo -e "${GREEN}✓ User registered successfully${NC}"
    echo -e "${GREEN}  Response: $REGISTER_BODY${NC}"
else
    echo -e "${YELLOW}⚠ Registration returned status: $REGISTER_STATUS${NC}"
    echo -e "${YELLOW}  This is okay if user already exists${NC}"
fi
echo ""

# Test 5: Generate telemetry data - Login
echo -e "${YELLOW}[5] Testing telemetry with login...${NC}"
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$AUTH_SERVICE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"TelemetryTest123!\"
  }")

LOGIN_STATUS=$(echo "$LOGIN_RESPONSE" | tail -n1)
LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$LOGIN_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
    TOKEN=$(echo "$LOGIN_BODY" | jq -r '.token' 2>/dev/null || echo "")
    if [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
        echo -e "${GREEN}  Got JWT token: ${TOKEN:0:30}...${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Login returned status: $LOGIN_STATUS${NC}"
fi
echo ""

# Test 6: Check business metrics
echo -e "${YELLOW}[6] Checking business metrics...${NC}"
BUSINESS_METRICS=$(curl -s "$METRICS_ENDPOINT" | grep "^business_" || echo "")

if [ -n "$BUSINESS_METRICS" ]; then
    echo -e "${GREEN}✓ Business metrics found:${NC}"
    echo "$BUSINESS_METRICS" | while read -r line; do
        echo -e "${GREEN}  $line${NC}"
    done
else
    echo -e "${YELLOW}⚠ No business metrics found${NC}"
    echo -e "${YELLOW}  Business metrics might not be enabled${NC}"
fi
echo ""

# Test 7: Check HTTP request metrics
echo -e "${YELLOW}[7] Checking HTTP request metrics...${NC}"
HTTP_METRICS=$(curl -s "$METRICS_ENDPOINT" | grep -E "http_request|request_duration" | head -10 || echo "")

if [ -n "$HTTP_METRICS" ]; then
    echo -e "${GREEN}✓ HTTP metrics found (showing first 10):${NC}"
    echo "$HTTP_METRICS" | while read -r line; do
        echo -e "${GREEN}  ${line:0:100}${NC}"
    done
else
    echo -e "${YELLOW}⚠ No HTTP metrics found${NC}"
fi
echo ""

# Test 8: Generate load for better metrics
echo -e "${YELLOW}[8] Generating traffic for metrics (10 requests)...${NC}"
for i in {1..10}; do
    curl -s "$AUTH_SERVICE_URL/api/v1/health" > /dev/null &
done
wait
echo -e "${GREEN}✓ Generated 10 health check requests${NC}"
echo ""

# Sleep to let metrics update
echo -e "${YELLOW}Waiting 2 seconds for metrics to update...${NC}"
sleep 2
echo ""

# Test 9: Check updated metrics
echo -e "${YELLOW}[9] Checking updated request count...${NC}"
REQUEST_TOTAL=$(curl -s "$METRICS_ENDPOINT" | grep "http_requests_total" | head -1 || echo "none")
echo -e "${GREEN}  $REQUEST_TOTAL${NC}"
echo ""

# Test 10: Check trace in Jaeger (if running)
if [ "$JAEGER_RUNNING" = true ]; then
    echo -e "${YELLOW}[10] Checking for traces in Jaeger...${NC}"
    echo -e "${BLUE}  Open Jaeger UI: $JAEGER_UI_URL${NC}"
    echo -e "${BLUE}  Service: auth-service${NC}"
    echo -e "${BLUE}  Operation: POST /api/v1/auth/register${NC}"
    echo -e "${BLUE}  Look for trace with user: $TEST_USER${NC}"
    
    # Try to query Jaeger API
    JAEGER_API="http://localhost:16686/api/traces?service=auth-service&limit=1"
    TRACES=$(curl -s "$JAEGER_API" 2>/dev/null || echo "")
    if [ -n "$TRACES" ]; then
        TRACE_COUNT=$(echo "$TRACES" | jq '.data | length' 2>/dev/null || echo "0")
        echo -e "${GREEN}  Found $TRACE_COUNT recent traces${NC}"
    fi
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}✓ Service Health Check: Pass${NC}"
echo -e "${GREEN}✓ Metrics Endpoint: Pass${NC}"
echo -e "${GREEN}✓ User Registration: Generated telemetry${NC}"
echo -e "${GREEN}✓ User Login: Generated telemetry${NC}"
echo -e "${GREEN}✓ Business Metrics: $([ -n "$BUSINESS_METRICS" ] && echo "Available" || echo "Not enabled")${NC}"
echo -e "${GREEN}✓ HTTP Metrics: $([ -n "$HTTP_METRICS" ] && echo "Available" || echo "Not enabled")${NC}"

if [ "$JAEGER_RUNNING" = true ]; then
    echo -e "${GREEN}✓ Jaeger Tracing: Available${NC}"
else
    echo -e "${YELLOW}⚠ Jaeger Tracing: Not running${NC}"
fi
echo ""

# Next steps
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Next Steps${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}1. View metrics:${NC}"
echo -e "   curl $METRICS_ENDPOINT | grep business"
echo ""
echo -e "${YELLOW}2. View traces in Jaeger:${NC}"
echo -e "   open $JAEGER_UI_URL"
echo ""
echo -e "${YELLOW}3. Query Prometheus (if running):${NC}"
echo -e "   open $PROMETHEUS_URL"
echo ""
echo -e "${YELLOW}4. Generate more traffic:${NC}"
echo -e "   ab -n 1000 -c 10 $AUTH_SERVICE_URL/api/v1/health"
echo ""
echo -e "${YELLOW}5. Check telemetry documentation:${NC}"
echo -e "   cat guides/auth-service/TELEMETRY_TESTING_GUIDE.md"
echo ""

echo -e "${GREEN}Telemetry testing complete! 🎉${NC}"


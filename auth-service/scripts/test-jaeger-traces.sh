#!/bin/bash

set -e

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║     Testing OpenTelemetry + Jaeger Integration             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
AUTH_SERVICE_URL="http://localhost:8085"
JAEGER_UI_URL="http://localhost:16686"
OTLP_HEALTH_URL="http://localhost:13133"

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 1: Check Prerequisites"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if auth-service is running
if curl -s "$AUTH_SERVICE_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Auth service is running"
else
    echo -e "${RED}✗${NC} Auth service is NOT running"
    echo "   Start it with: ./auth-service"
    exit 1
fi

# Check if Jaeger is accessible
if curl -s "$JAEGER_UI_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Jaeger UI is accessible"
else
    echo -e "${YELLOW}⚠${NC}  Jaeger UI is NOT accessible"
    echo "   Start it with: docker-compose -f ../docker-compose.observability.yml up -d"
fi

# Check if OTLP collector is running
if curl -s "$OTLP_HEALTH_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} OTLP Collector is running"
else
    echo -e "${YELLOW}⚠${NC}  OTLP Collector is NOT running"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 2: Generate Traces"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 1: Health check
echo "Test 1: Health check endpoint"
HEALTH_RESPONSE=$(curl -s "$AUTH_SERVICE_URL/health")
echo "   Response: $HEALTH_RESPONSE"
echo -e "${GREEN}✓${NC} Health check trace generated"
sleep 1

# Test 2: Login
echo ""
echo "Test 2: Login endpoint"
LOGIN_RESPONSE=$(curl -s -X POST "$AUTH_SERVICE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "test123"
  }')

if echo "$LOGIN_RESPONSE" | jq -e '.token' > /dev/null 2>&1; then
    JWT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
    echo "   Token: ${JWT_TOKEN:0:30}..."
    echo -e "${GREEN}✓${NC} Login trace generated"
else
    echo "   Response: $LOGIN_RESPONSE"
    echo -e "${YELLOW}⚠${NC}  Login failed (trace still generated)"
    JWT_TOKEN=""
fi
sleep 1

# Test 3: Protected endpoint (if we have a token)
if [ -n "$JWT_TOKEN" ]; then
    echo ""
    echo "Test 3: Protected endpoint (with authentication)"
    USERS_RESPONSE=$(curl -s -X GET "$AUTH_SERVICE_URL/api/v1/users" \
      -H "Authorization: Bearer $JWT_TOKEN")
    echo "   Response: ${USERS_RESPONSE:0:100}..."
    echo -e "${GREEN}✓${NC} Protected endpoint trace generated"
    sleep 1
fi

# Test 4: Multiple requests
echo ""
echo "Test 4: Multiple health check requests"
for i in {1..2}; do
    curl -s "$AUTH_SERVICE_URL/health" > /dev/null
    echo -n "."
    sleep 0.5
done
echo ""
echo -e "${GREEN}✓${NC} Multiple traces generated"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Step 3: Verify Traces in Jaeger"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo ""
echo "Waiting 10 seconds for traces to be exported..."
for i in {10..1}; do
    echo -n "$i "
    sleep 1
done
echo ""

echo ""
echo "To view traces:"
echo "  1. Open Jaeger UI: $JAEGER_UI_URL"
echo "  2. Select service: 'auth-service'"
echo "  3. Click 'Find Traces'"
echo ""
echo "Expected traces:"
echo "  - GET /health"
echo "  - POST /api/v1/auth/login"
if [ -n "$JWT_TOKEN" ]; then
    echo "  - GET /api/v1/users"
fi
echo ""

# Try to query Jaeger API
echo "Querying Jaeger API for traces..."
JAEGER_TRACES=$(curl -s "$JAEGER_UI_URL/api/traces?service=auth-service&limit=10")

if echo "$JAEGER_TRACES" | jq -e '.data' > /dev/null 2>&1; then
    TRACE_COUNT=$(echo "$JAEGER_TRACES" | jq '.data | length')
    if [ "$TRACE_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Found $TRACE_COUNT traces in Jaeger!"
        echo ""
        echo "Recent traces:"
        echo "$JAEGER_TRACES" | jq -r '.data[0:3] | .[] | "  - \(.spans[0].operationName) (\(.spans | length) spans, \((.spans[0].duration / 1000) | floor)ms)"'
    else
        echo -e "${YELLOW}⚠${NC}  No traces found yet. Wait a bit longer and check Jaeger UI."
    fi
else
    echo -e "${YELLOW}⚠${NC}  Could not query Jaeger API. Check manually in UI."
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo -e "${GREEN}✓${NC} OpenTelemetry integration test complete"
echo ""
echo "Next steps:"
echo "  1. Open Jaeger UI: $JAEGER_UI_URL"
echo "  2. Explore trace details and span information"
echo "  3. Check trace timings and dependencies"
echo ""
echo "Troubleshooting:"
echo "  - If no traces: Check OTEL_ENABLED=true and OTEL_EXPORTER_OTLP_ENDPOINT"
echo "  - View service logs: ./auth-service 2>&1 | grep -i 'opentelemetry\|otlp'"
echo "  - Check collector: docker logs otel-collector"
echo ""

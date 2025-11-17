#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║   🔍 OpenTelemetry + Jaeger Diagnostic Tool                ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

# Check function
check() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
        ((PASSED++))
    else
        echo -e "${RED}✗${NC} $2"
        ((FAILED++))
        if [ -n "$3" ]; then
            echo -e "   ${YELLOW}Fix:${NC} $3"
        fi
    fi
}

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  Checking Docker Services"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if Jaeger is running
docker ps | grep -q jaeger
check $? "Jaeger container is running" "docker-compose -f ../docker-compose.observability.yml up -d"

# Check if otel-collector is running (if used)
if docker ps | grep -q otel-collector; then
    check 0 "OTLP Collector container is running"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  Checking Service Endpoints"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check Jaeger UI
curl -s http://localhost:16686 > /dev/null
check $? "Jaeger UI accessible (http://localhost:16686)"

# Check Jaeger OTLP endpoint
curl -s http://localhost:4318 > /dev/null
check $? "Jaeger OTLP HTTP endpoint (http://localhost:4318)"

# Check auth-service
curl -s http://localhost:8085/health > /dev/null
check $? "Auth-service is running (http://localhost:8085)"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  Checking Environment Variables"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

[ "$OTEL_ENABLED" = "true" ]
check $? "OTEL_ENABLED=true" "export OTEL_ENABLED=true"

[ -n "$OTEL_EXPORTER_OTLP_ENDPOINT" ]
check $? "OTEL_EXPORTER_OTLP_ENDPOINT is set (current: $OTEL_EXPORTER_OTLP_ENDPOINT)" "export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  Checking Service Logs"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f "auth-service.log" ]; then
    if grep -q "OpenTelemetry initialized successfully" auth-service.log; then
        check 0 "OpenTelemetry initialization found in logs"
    else
        check 1 "OpenTelemetry initialization NOT found in logs" "Check service startup logs"
    fi
    
    if grep -q "traces export.*connection refused" auth-service.log; then
        check 1 "Found connection refused errors in logs" "Verify OTLP endpoint and use HTTP (not HTTPS)"
    fi
else
    echo -e "${YELLOW}⚠${NC}  No auth-service.log file found"
    echo "   Run: ./auth-service 2>&1 | tee auth-service.log"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "5️⃣  Testing Trace Generation"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Sending test requests..."
for i in {1..5}; do
    curl -s http://localhost:8085/health > /dev/null
    echo -n "."
    sleep 0.5
done
echo ""
check 0 "Sent 5 test requests to generate traces"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "6️⃣  Querying Jaeger API"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

echo "Waiting 10 seconds for traces to be exported..."
sleep 10

SERVICES=$(curl -s http://localhost:16686/api/services | jq -r '.data[]' 2>/dev/null)
if echo "$SERVICES" | grep -q "auth-service"; then
    check 0 "auth-service found in Jaeger!"
    
    TRACES=$(curl -s "http://localhost:16686/api/traces?service=auth-service&limit=10" | jq '.data | length' 2>/dev/null)
    if [ "$TRACES" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Found $TRACES traces for auth-service"
    else
        echo -e "${YELLOW}⚠${NC}  No traces found yet (this is normal if just started)"
    fi
else
    check 1 "auth-service NOT found in Jaeger" "Check steps 1-4 above"
    echo ""
    echo "Available services in Jaeger:"
    echo "$SERVICES" | sed 's/^/   - /'
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ Everything looks good!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Open Jaeger UI: http://localhost:16686"
    echo "  2. Select 'auth-service' from dropdown"
    echo "  3. Click 'Find Traces'"
    echo "  4. Explore your traces!"
else
    echo -e "${RED}❌ Some checks failed${NC}"
    echo ""
    echo "Quick fix commands:"
    echo ""
    echo "# 1. Start observability stack"
    echo "cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang"
    echo "docker-compose -f docker-compose.observability.yml up -d"
    echo ""
    echo "# 2. Set environment variables"
    echo "export OTEL_ENABLED=true"
    echo "export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318"
    echo ""
    echo "# 3. Restart auth-service"
    echo "cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service"
    echo "./auth-service 2>&1 | tee auth-service.log"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

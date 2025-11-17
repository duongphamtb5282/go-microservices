# Jaeger Troubleshooting Guide

## Issue: auth-service not appearing in Jaeger

### Quick Check Script

Run this to diagnose the issue:

```bash
#!/bin/bash

echo "🔍 Diagnosing OpenTelemetry + Jaeger Connection..."
echo ""

# 1. Check if observability stack is running
echo "1️⃣ Checking observability stack..."
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang
if docker-compose -f docker-compose.observability.yml ps | grep -q "Up"; then
    echo "✅ Observability stack is running"
else
    echo "❌ Observability stack is NOT running"
    echo "   Fix: docker-compose -f docker-compose.observability.yml up -d"
fi

# 2. Check Jaeger UI
echo ""
echo "2️⃣ Checking Jaeger UI..."
if curl -s http://localhost:16686 > /dev/null 2>&1; then
    echo "✅ Jaeger UI is accessible at http://localhost:16686"
else
    echo "❌ Jaeger UI is NOT accessible"
fi

# 3. Check OTLP Collector
echo ""
echo "3️⃣ Checking OTLP Collector..."
if curl -s http://localhost:4318 > /dev/null 2>&1; then
    echo "✅ OTLP Collector HTTP endpoint is accessible"
else
    echo "❌ OTLP Collector is NOT accessible"
fi

# 4. Check environment variables
echo ""
echo "4️⃣ Checking environment variables..."
echo "   OTEL_ENABLED=${OTEL_ENABLED:-not set}"
echo "   OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT:-not set}"

if [ "$OTEL_ENABLED" = "true" ]; then
    echo "✅ OTEL_ENABLED is set to true"
else
    echo "❌ OTEL_ENABLED is not set to true"
    echo "   Fix: export OTEL_ENABLED=true"
fi

if [ -n "$OTEL_EXPORTER_OTLP_ENDPOINT" ]; then
    echo "✅ OTEL_EXPORTER_OTLP_ENDPOINT is set"
else
    echo "❌ OTEL_EXPORTER_OTLP_ENDPOINT is not set"
    echo "   Fix: export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318"
fi

# 5. Check if auth-service is running
echo ""
echo "5️⃣ Checking auth-service..."
if curl -s http://localhost:8085/health > /dev/null 2>&1; then
    echo "✅ Auth service is running"
else
    echo "❌ Auth service is NOT running"
fi

# 6. Test if traces are being sent
echo ""
echo "6️⃣ Generating test trace..."
curl -s http://localhost:8085/health > /dev/null
echo "   Sent health check request"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "�� If all checks pass, wait 30 seconds and check Jaeger UI"
echo "   URL: http://localhost:16686"
echo ""
echo "🔧 Common Fixes:"
echo "   1. Restart auth-service WITH environment variables"
echo "   2. Check service logs for OpenTelemetry errors"
echo "   3. Verify OTLP endpoint is HTTP not HTTPS"

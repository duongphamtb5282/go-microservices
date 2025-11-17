#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║   🚀 Starting Auth-Service with OpenTelemetry              ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Set OpenTelemetry environment variables
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=auth-service
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318

# Set database credentials
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

echo "✅ Environment variables set:"
echo "   OTEL_ENABLED=$OTEL_ENABLED"
echo "   OTEL_SERVICE_NAME=$OTEL_SERVICE_NAME"
echo "   OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT"
echo ""

# Check if Jaeger is running
if curl -s http://localhost:16686 > /dev/null 2>&1; then
    echo "✅ Jaeger UI is accessible at http://localhost:16686"
else
    echo "⚠️  Jaeger UI is NOT accessible"
    echo "   Start it with:"
    echo "   cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang"
    echo "   docker-compose -f docker-compose.observability.yml up -d"
    echo ""
fi

if curl -s http://localhost:4318 > /dev/null 2>&1; then
    echo "✅ OTLP endpoint is accessible at http://localhost:4318"
else
    echo "⚠️  OTLP endpoint is NOT accessible"
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Starting auth-service..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Look for:"
echo '  "message":"OpenTelemetry initialized successfully"'
echo ""

# Start the service
./auth-service 2>&1 | tee auth-service.log

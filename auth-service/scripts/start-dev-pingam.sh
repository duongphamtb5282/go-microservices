#!/bin/bash
set -e

echo "=== Starting Auth Service in Dev Mode with PingAM ==="

# 1. Stop existing services
echo "Stopping existing services..."
pkill -f "./auth-service" 2>/dev/null || true
lsof -ti:8085 | xargs kill -9 2>/dev/null || true
sleep 1

# 2. Start PingAM mock
echo "Starting PingAM mock..."
docker-compose -f docker-compose.pingam.yml up -d pingam-mock postgres
sleep 3

# Verify PingAM mock is accessible
if curl -s http://localhost:1080/mockserver/status > /dev/null 2>&1; then
    echo "✓ PingAM mock is running"
else
    echo "✗ PingAM mock is not accessible, waiting..."
    sleep 5
fi

# 3. Verify config
echo "Checking config..."
MODE=$(grep -A 2 "^authorization:" config/config.yaml | grep "mode:" | awk '{print $2}' | tr -d '"' | tr -d "'")
echo "Authorization mode in config: $MODE"

if [ "$MODE" != "pingam" ]; then
    echo "⚠ WARNING: Mode is '$MODE', not 'pingam'"
    echo "Please update config/config.yaml manually or run:"
    echo "  sed -i '' 's/mode: \"jwt\"/mode: \"pingam\"/' config/config.yaml"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 4. Build
echo "Building..."
go build -o auth-service cmd/http/main.go
if [ $? -ne 0 ]; then
    echo "✗ Build failed"
    exit 1
fi
echo "✓ Build successful"

# 5. Start service
echo "Starting auth-service..."
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password
./auth-service > dev-pingam.log 2>&1 &
SERVICE_PID=$!
echo "Service started with PID: $SERVICE_PID"

# 6. Wait and verify
sleep 4
echo ""
echo "=== Service Status ==="
if ps -p $SERVICE_PID > /dev/null; then
    echo "✓ Service is running"
    if curl -s http://localhost:8085/api/v1/health > /dev/null 2>&1; then
        echo "✓ Health check passed"
        curl -s http://localhost:8085/api/v1/health | jq '.'
    else
        echo "✗ Health check failed"
    fi
else
    echo "✗ Service failed to start"
    echo ""
    echo "=== Last 20 lines of log ==="
    tail -20 dev-pingam.log
    exit 1
fi

echo ""
echo "=== Config Verification ==="
echo "Checking what mode was loaded..."
if grep -q "DEBUG CONFIG" dev-pingam.log; then
    grep "DEBUG CONFIG" dev-pingam.log
else
    echo "⚠ No DEBUG CONFIG found in logs (old binary?)"
fi

if grep -q "DEBUG.*Authorization config" dev-pingam.log; then
    grep "DEBUG.*Authorization config" dev-pingam.log
else
    echo "⚠ No authorization debug found"
fi

echo ""
echo "=== Ready to Test ==="
echo ""
echo "Option 1 - Use test script:"
echo "  ./test-pingam-mode.sh"
echo ""
echo "Option 2 - Manual test:"
echo '  # Login'
echo '  JWT_TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \'
echo '    -H "Content-Type: application/json" \'
echo '    -d'"'"'{"email":"test@example.com","password":"any"}'"'"' | jq -r '"'"'.token'"'"')'
echo ''
echo '  # Test endpoint (should use PingAM)'
echo '  curl -X GET http://localhost:8085/api/v1/users -H "Authorization: Bearer $JWT_TOKEN" | jq '"'"'.'"'"
echo ""
echo "View logs: tail -f dev-pingam.log"
echo "Stop service: pkill -f './auth-service'"
echo ""
echo "========================================="


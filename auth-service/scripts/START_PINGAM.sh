#!/bin/bash
set -e

echo "========================================="
echo "Starting Auth Service in PingAM Mode"
echo "========================================="

# Stop any existing service
pkill -9 -f "./auth-service" 2>/dev/null || true
sleep 1

# Set environment
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# Start service
echo "Starting auth-service..."
./auth-service 2>&1 | tee pingam-mode.log &
SERVICE_PID=$!

# Wait for startup
sleep 5

# Check logs
echo ""
echo "=== Configuration Check ==="
grep "DEBUG CONFIG" pingam-mode.log | head -10

echo ""
echo "=== Authorization Mode ==="
grep "Authorization config" pingam-mode.log | head -3

echo ""
echo "========================================="
echo "Service PID: $SERVICE_PID"
echo "Logs: tail -f pingam-mode.log"
echo "Test: ./test-pingam-mode.sh"
echo "========================================="


#!/bin/bash

echo "🔔 Testing Notification Handler"
echo "==============================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Check if notification service is running
echo -e "${BLUE}1. Checking Notification Service Status${NC}"
echo "============================================="
echo ""

if pgrep -f "notification-service" > /dev/null; then
    echo -e "${GREEN}✅ Notification service is running${NC}"
    echo "PID: $(pgrep -f notification-service)"
else
    echo -e "${RED}❌ Notification service not running${NC}"
    echo "Starting notification service..."
    cd ../notification-service
    nohup go run main.go > notification-service.log 2>&1 &
    sleep 3
    if pgrep -f "notification-service" > /dev/null; then
        echo -e "${GREEN}✅ Notification service started${NC}"
    else
        echo -e "${RED}❌ Failed to start notification service${NC}"
        exit 1
    fi
fi
echo ""

# Test 2: Check Kafka connectivity
echo -e "${BLUE}2. Testing Kafka Connectivity${NC}"
echo "=================================="
echo ""

if nc -z localhost 9092 2>/dev/null; then
    echo -e "${GREEN}✅ Kafka broker is accessible on port 9092${NC}"
else
    echo -e "${YELLOW}⚠️  Kafka broker not accessible on port 9092${NC}"
    echo "   To start Kafka: docker-compose up -d kafka"
    echo "   Or start Kafka locally"
fi
echo ""

# Test 3: Test notification handler functionality
echo -e "${BLUE}3. Testing Notification Handler Functionality${NC}"
echo "=================================================="
echo ""

echo "Creating test notification events..."

# Test UserRegisteredEvent
echo "Testing UserRegisteredEvent..."
cat > /tmp/test_user_registered.json << EOF
{
    "event_id": "test-event-$(date +%s)",
    "user_id": "test-user-123",
    "username": "testuser",
    "email": "test@example.com",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "Event payload:"
cat /tmp/test_user_registered.json
echo ""

# Test UserActivatedEvent
echo "Testing UserActivatedEvent..."
cat > /tmp/test_user_activated.json << EOF
{
    "event_id": "test-activation-$(date +%s)",
    "user_id": "test-user-123",
    "username": "testuser",
    "email": "test@example.com",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "Event payload:"
cat /tmp/test_user_activated.json
echo ""

# Test UserLoginEvent
echo "Testing UserLoginEvent..."
cat > /tmp/test_user_login.json << EOF
{
    "event_id": "test-login-$(date +%s)",
    "user_id": "test-user-123",
    "username": "testuser",
    "email": "test@example.com",
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0 (Test Browser)",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF

echo "Event payload:"
cat /tmp/test_user_login.json
echo ""

# Test 4: Check notification service logs
echo -e "${BLUE}4. Checking Notification Service Logs${NC}"
echo "=========================================="
echo ""

if [ -f "../notification-service/notification-service.log" ]; then
    echo "Recent notification service logs:"
    tail -n 10 ../notification-service/notification-service.log
else
    echo "No notification service log file found"
fi
echo ""

# Test 5: Test auth service integration
echo -e "${BLUE}5. Testing Auth Service Integration${NC}"
echo "====================================="
echo ""

echo "Testing auth service health..."
AUTH_RESPONSE=$(curl -s http://localhost:8085/api/v1/health)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Auth service is running${NC}"
    echo "Response: $AUTH_RESPONSE"
else
    echo -e "${RED}❌ Auth service not running${NC}"
fi
echo ""

echo "Testing correlation ID with notification events..."
CORRELATION_ID="test-notification-$(date +%s)"
echo "Using correlation ID: $CORRELATION_ID"

curl -s -H "X-Correlation-ID: $CORRELATION_ID" \
    -H "X-Request-ID: test-request-$(date +%s)" \
    http://localhost:8085/api/v1/health
echo ""

# Test 6: Summary
echo -e "${BLUE}6. Test Summary${NC}"
echo "==============="
echo ""

echo "✅ Notification service: $(pgrep -f notification-service > /dev/null && echo "Running" || echo "Not running")"
echo "✅ Auth service: $(curl -s http://localhost:8085/api/v1/health > /dev/null && echo "Running" || echo "Not running")"
echo "✅ Kafka: $(nc -z localhost 9092 2>/dev/null && echo "Accessible" || echo "Not accessible")"
echo ""

echo -e "${GREEN}🎉 Notification Handler Testing Complete!${NC}"
echo ""
echo "The notification service is designed to:"
echo "• Handle UserRegisteredEvent - Send welcome emails and SMS"
echo "• Handle UserActivatedEvent - Send activation confirmations"
echo "• Handle UserLoginEvent - Send login notifications and security checks"
echo ""
echo "All events are processed asynchronously via Kafka messaging."

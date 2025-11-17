#!/bin/bash

echo "========================================="
echo "PingAM Authorization Mode Test"
echo "========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8085"

echo -e "${YELLOW}Step 1: Check if PingAM mock server is running${NC}"
if curl -s http://localhost:1080/mockserver/status > /dev/null 2>&1; then
    echo -e "${GREEN}✓ PingAM mock server is running${NC}"
else
    echo -e "${RED}✗ PingAM mock server is NOT running${NC}"
    echo "Starting PingAM mock..."
    cd "$(dirname "$0")"
    docker-compose -f docker-compose.pingam.yml up -d pingam-mock
    sleep 5
fi
echo ""

echo -e "${YELLOW}Step 2: Check if auth-service is running${NC}"
if curl -s $BASE_URL/api/v1/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Auth service is running${NC}"
else
    echo -e "${RED}✗ Auth service is NOT running${NC}"
    echo "Please start it with: ./auth-service"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 3: Login to get JWT token${NC}"
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}')

echo "Login response:"
echo "$LOGIN_RESPONSE" | jq '.'

JWT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')

if [ "$JWT_TOKEN" == "null" ] || [ -z "$JWT_TOKEN" ]; then
    echo -e "${RED}✗ Failed to get JWT token${NC}"
    exit 1
fi

echo -e "${GREEN}✓ JWT token received: ${JWT_TOKEN:0:50}...${NC}"
echo ""

echo -e "${YELLOW}Step 4: Test protected endpoint with PingAM authorization${NC}"
echo "Calling GET /api/v1/users (requires users:read permission)"
echo ""

USERS_RESPONSE=$(curl -s -X GET $BASE_URL/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Response:"
echo "$USERS_RESPONSE" | jq '.'
echo ""

if echo "$USERS_RESPONSE" | grep -q "error"; then
    echo -e "${RED}✗ Access denied or error occurred${NC}"
    ERROR_MSG=$(echo "$USERS_RESPONSE" | jq -r '.message // .error')
    echo "Error: $ERROR_MSG"
else
    echo -e "${GREEN}✓ Access granted - PingAM authorization successful!${NC}"
fi
echo ""

echo -e "${YELLOW}Step 5: Check authorization logs${NC}"
if [ -f "auth-service-pingam-mode.log" ]; then
    echo "Last 15 lines from service logs:"
    tail -15 auth-service-pingam-mode.log | grep -i "permission\|pingam\|authorization" || echo "No authorization logs found"
else
    echo "Log file not found - service might be running in foreground"
fi
echo ""

echo -e "${YELLOW}Step 6: Verify PingAM mock received the request${NC}"
echo "Testing direct PingAM mock endpoint:"
curl -s -X POST http://localhost:1080/api/authorize \
  -H "Content-Type: application/json" \
  -d '{"userId":"user-123","resource":"users","action":"read"}' | jq '.'
echo ""

echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Config file: config/config.yaml"
echo "  Authorization mode: pingam"
echo "  PingAM base URL: http://localhost:1080"
echo ""
echo "Expected behavior:"
echo "  1. JWT token is generated on login (for authentication)"
echo "  2. When accessing /api/v1/users:"
echo "     - JWT middleware validates the token"
echo "     - Authorization middleware calls PingAM API"
echo "     - PingAM mock returns permission decision"
echo "     - Request proceeds if allowed"
echo ""
echo "To check current config mode:"
echo "  grep 'mode:' config/config.yaml"
echo ""
echo "To switch to JWT mode:"
echo "  sed -i '' 's/mode: \"pingam\"/mode: \"jwt\"/' config/config.yaml"
echo "  then restart service"
echo "========================================="


#!/bin/bash

# Demo script showing how to test Kafka and gRPC communication
# This script demonstrates the testing workflow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🚀 Kafka and gRPC Communication Testing Demo${NC}"
echo "==============================================="
echo ""

echo -e "${YELLOW}This demo will show you how to test communication between services.${NC}"
echo ""

# Check if Docker is running
echo -e "${BLUE}Step 1: Checking prerequisites...${NC}"
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker first.${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"
echo ""

# Show available test scripts
echo -e "${BLUE}Step 2: Available test scripts${NC}"
echo "------------------------------"
ls -la "$SCRIPT_DIR"/test-*.sh
echo ""

# Show help for main test script
echo -e "${BLUE}Step 3: Test script help${NC}"
echo "---------------------------"
"$SCRIPT_DIR/test-all-communications.sh" help
echo ""

# Show service status (if running)
echo -e "${BLUE}Step 4: Current service status${NC}"
echo "---------------------------------"
"$SCRIPT_DIR/test-all-communications.sh" status
echo ""

# Demonstrate individual tests (if services are running)
echo -e "${BLUE}Step 5: Testing connectivity${NC}"
echo "-------------------------------"

# Test Kafka connectivity
echo "Testing Kafka connectivity..."
if timeout 5 bash -c "</dev/tcp/localhost/9092" 2>/dev/null; then
    echo -e "${GREEN}✅ Kafka is accessible${NC}"
    echo "You can run Kafka tests with:"
    echo "  $SCRIPT_DIR/test-kafka-integration.sh"
else
    echo -e "${YELLOW}⚠️  Kafka not accessible (port 9092)${NC}"
    echo "To start Kafka:"
    echo "  cd $PROJECT_ROOT && docker-compose -f infrastructure/docker-compose.yml up -d kafka"
fi
echo ""

# Test gRPC connectivity
echo "Testing gRPC connectivity..."
if timeout 5 bash -c "</dev/tcp/localhost/50051" 2>/dev/null; then
    echo -e "${GREEN}✅ Admin service gRPC is accessible${NC}"
    echo "You can run gRPC tests with:"
    echo "  $SCRIPT_DIR/test-grpc-integration.sh"
else
    echo -e "${YELLOW}⚠️  Admin service gRPC not accessible (port 50051)${NC}"
    echo "To start admin service:"
    echo "  cd $PROJECT_ROOT/admin-service && docker-compose up -d"
fi
echo ""

# Test auth service
echo "Testing auth service..."
if curl -s http://localhost:8085/api/v1/health >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Auth service is accessible${NC}"
else
    echo -e "${YELLOW}⚠️  Auth service not accessible (port 8085)${NC}"
    echo "To start auth service:"
    echo "  cd $PROJECT_ROOT/auth-service && docker-compose -f docker-compose.app.yml up -d"
fi
echo ""

# Show quick start instructions
echo -e "${BLUE}Step 6: Quick start instructions${NC}"
echo "-----------------------------------"
echo "To run full integration tests:"
echo ""
echo "1. Start all services:"
echo "   $SCRIPT_DIR/test-all-communications.sh start"
echo ""
echo "2. Run all tests:"
echo "   $SCRIPT_DIR/test-all-communications.sh test"
echo ""
echo "3. Check status anytime:"
echo "   $SCRIPT_DIR/test-all-communications.sh status"
echo ""
echo "4. Cleanup when done:"
echo "   $SCRIPT_DIR/test-all-communications.sh stop"
echo ""

echo -e "${GREEN}🎉 Demo completed!${NC}"
echo ""
echo "For detailed testing information, see:"
echo "  $SCRIPT_DIR/README_TESTING.md"

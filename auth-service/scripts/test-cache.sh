#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_header() {
    echo -e "${BLUE}$1${NC}"
}

print_step() {
    echo -e "${CYAN}➤ $1${NC}"
}

BASE_URL="http://localhost:8085/api/v1"

print_header "========================================"
print_header "  Auth Service Cache Testing Script"
print_header "========================================"
echo ""

# Check if Redis is running
print_step "Checking Redis connection..."
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    print_success "Redis is running"
else
    print_error "Redis is not running. Starting Redis..."
    docker-compose up -d redis
    sleep 2
    if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
        print_success "Redis started successfully"
    else
        print_error "Failed to start Redis"
        exit 1
    fi
fi

# Check if service is running
print_step "Checking if auth-service is running..."
if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
    print_success "Auth-service is running"
else
    print_error "Auth-service is not running. Please start it with: ./start-service.sh"
    exit 1
fi

echo ""
print_header "Test 1: Create User & Populate Cache"
print_header "========================================"

USERNAME="cachetest_$(date +%s)"
EMAIL="${USERNAME}@example.com"

print_step "Creating user: ${USERNAME}"
RESPONSE=$(curl -s -X POST "${BASE_URL}/users" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"email\": \"${EMAIL}\",
    \"password\": \"SecurePass123!\"
  }")

USER_ID=$(echo "$RESPONSE" | jq -r '.user.id')

if [ "$USER_ID" != "null" ] && [ -n "$USER_ID" ]; then
    print_success "User created successfully"
    print_info "User ID: ${USER_ID}"
    print_info "Username: ${USERNAME}"
    print_info "Email: ${EMAIL}"
else
    print_error "Failed to create user"
    echo "$RESPONSE" | jq '.'
    exit 1
fi

echo ""
print_header "Test 2: Verify Cache in Redis"
print_header "========================================"

print_step "Checking if user is cached in Redis..."
sleep 1

# Check user by ID cache
CACHE_KEY="user:${USER_ID}"
if docker-compose exec -T redis redis-cli EXISTS "${CACHE_KEY}" | grep -q "1"; then
    print_success "User found in cache: ${CACHE_KEY}"
    
    print_info "Cache content:"
    docker-compose exec -T redis redis-cli GET "${CACHE_KEY}" | jq '.'
    
    # Check TTL
    TTL=$(docker-compose exec -T redis redis-cli TTL "${CACHE_KEY}" | tr -d '\r')
    print_info "Cache TTL: ${TTL} seconds (~$(($TTL / 60)) minutes)"
else
    print_error "User NOT found in cache"
fi

# Check user by email cache
EMAIL_CACHE_KEY="user:email:${EMAIL}"
if docker-compose exec -T redis redis-cli EXISTS "${EMAIL_CACHE_KEY}" | grep -q "1"; then
    print_success "User email mapping found in cache: ${EMAIL_CACHE_KEY}"
    CACHED_ID=$(docker-compose exec -T redis redis-cli GET "${EMAIL_CACHE_KEY}" | tr -d '\r')
    print_info "Cached User ID: ${CACHED_ID}"
else
    print_error "User email mapping NOT found in cache"
fi

echo ""
print_header "Test 3: Cache Hit Test (Fast Response)"
print_header "========================================"

print_step "Fetching user from cache (should be fast)..."
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -X GET "${BASE_URL}/users/${USER_ID}")
END_TIME=$(date +%s%N)
DURATION=$((($END_TIME - $START_TIME) / 1000000)) # Convert to milliseconds

if echo "$RESPONSE" | jq -e '.user' > /dev/null 2>&1; then
    print_success "User retrieved successfully"
    print_info "Response time: ${DURATION}ms"
    
    if [ $DURATION -lt 50 ]; then
        print_success "Fast response - likely a cache hit!"
    else
        print_info "Slower response - might be a cache miss (${DURATION}ms)"
    fi
else
    print_error "Failed to retrieve user"
    echo "$RESPONSE"
fi

# Check logs for cache hit
print_step "Checking logs for cache hit..."
if tail -10 auth-service.log 2>/dev/null | grep -q "User found in cache"; then
    print_success "Logs confirm: CACHE HIT ✨"
elif tail -10 auth-service.log 2>/dev/null | grep -q "Cache miss"; then
    print_info "Logs show: CACHE MISS (fetched from database)"
else
    print_info "Could not determine cache status from logs"
fi

echo ""
print_header "Test 4: Cache Invalidation Test"
print_header "========================================"

print_step "Clearing cache for user ${USER_ID}..."
docker-compose exec -T redis redis-cli DEL "${CACHE_KEY}" > /dev/null
docker-compose exec -T redis redis-cli DEL "${EMAIL_CACHE_KEY}" > /dev/null
print_success "Cache cleared"

print_step "Verifying cache is empty..."
if docker-compose exec -T redis redis-cli EXISTS "${CACHE_KEY}" | grep -q "0"; then
    print_success "Cache confirmed empty"
else
    print_error "Cache still contains data"
fi

print_step "Fetching user again (should miss cache, hit database)..."
START_TIME=$(date +%s%N)
RESPONSE=$(curl -s -X GET "${BASE_URL}/users/${USER_ID}")
END_TIME=$(date +%s%N)
DURATION=$((($END_TIME - $START_TIME) / 1000000))

if echo "$RESPONSE" | jq -e '.user' > /dev/null 2>&1; then
    print_success "User retrieved from database"
    print_info "Response time: ${DURATION}ms"
    
    if [ $DURATION -gt 20 ]; then
        print_success "Slower response confirms database access"
    fi
else
    print_error "Failed to retrieve user"
fi

print_step "Verifying cache was repopulated..."
sleep 1
if docker-compose exec -T redis redis-cli EXISTS "${CACHE_KEY}" | grep -q "1"; then
    print_success "Cache repopulated after database fetch"
else
    print_error "Cache was NOT repopulated"
fi

echo ""
print_header "Test 5: Cache Performance Comparison"
print_header "========================================"

print_step "Running 3 consecutive requests to measure performance..."

# Request 1: Should hit cache
START_TIME=$(date +%s%N)
curl -s -X GET "${BASE_URL}/users/${USER_ID}" > /dev/null
END_TIME=$(date +%s%N)
TIME1=$((($END_TIME - $START_TIME) / 1000000))

# Request 2: Should hit cache
START_TIME=$(date +%s%N)
curl -s -X GET "${BASE_URL}/users/${USER_ID}" > /dev/null
END_TIME=$(date +%s%N)
TIME2=$((($END_TIME - $START_TIME) / 1000000))

# Clear cache
docker-compose exec -T redis redis-cli DEL "${CACHE_KEY}" > /dev/null

# Request 3: Should miss cache
START_TIME=$(date +%s%N)
curl -s -X GET "${BASE_URL}/users/${USER_ID}" > /dev/null
END_TIME=$(date +%s%N)
TIME3=$((($END_TIME - $START_TIME) / 1000000))

echo ""
print_info "Request 1 (Cache Hit):  ${TIME1}ms"
print_info "Request 2 (Cache Hit):  ${TIME2}ms"
print_info "Request 3 (Cache Miss): ${TIME3}ms"

AVG_CACHE_HIT=$(( ($TIME1 + $TIME2) / 2 ))
SPEEDUP=$(( $TIME3 * 100 / $AVG_CACHE_HIT ))

echo ""
if [ $TIME3 -gt $AVG_CACHE_HIT ]; then
    IMPROVEMENT=$(( $TIME3 - $AVG_CACHE_HIT ))
    print_success "Cache is ${IMPROVEMENT}ms faster (${SPEEDUP}% of database time)"
else
    print_info "Performance similar (cache might not be working optimally)"
fi

echo ""
print_header "Test 6: Redis Cache Statistics"
print_header "========================================"

print_step "Fetching Redis statistics..."
echo ""

# Count all cached users
USER_KEYS=$(docker-compose exec -T redis redis-cli KEYS "user:*" | wc -l | tr -d ' ')
print_info "Total cached user keys: ${USER_KEYS}"

# Memory usage
MEMORY=$(docker-compose exec -T redis redis-cli INFO memory | grep "used_memory_human" | cut -d: -f2 | tr -d '\r')
print_info "Redis memory usage: ${MEMORY}"

# Connected clients
CLIENTS=$(docker-compose exec -T redis redis-cli INFO clients | grep "connected_clients" | cut -d: -f2 | tr -d '\r')
print_info "Connected clients: ${CLIENTS}"

# Total commands processed
COMMANDS=$(docker-compose exec -T redis redis-cli INFO stats | grep "total_commands_processed" | cut -d: -f2 | tr -d '\r')
print_info "Total commands processed: ${COMMANDS}"

echo ""
print_header "========================================"
print_header "  Cache Testing Complete!"
print_header "========================================"
echo ""
print_success "Summary:"
echo "  - Test User ID: ${USER_ID}"
echo "  - Username: ${USERNAME}"
echo "  - Email: ${EMAIL}"
echo "  - Cache Key: ${CACHE_KEY}"
echo "  - Avg Cache Hit Time: ${AVG_CACHE_HIT}ms"
echo "  - Database Fetch Time: ${TIME3}ms"
echo ""
print_info "To manually inspect cache:"
echo "  docker-compose exec redis redis-cli"
echo "  > KEYS user:*"
echo "  > GET ${CACHE_KEY}"
echo ""
print_info "To view logs:"
echo "  tail -f auth-service.log | grep cache"
echo ""


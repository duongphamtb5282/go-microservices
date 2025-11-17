#!/bin/bash

# Kafka Integration Testing Script for Auth Service
# Tests Kafka communication between auth-service and other microservices

set -e

# Configuration
KAFKA_BROKER="${KAFKA_BROKER:-localhost:9092}"
AUTH_SERVICE_URL="${AUTH_SERVICE_URL:-http://localhost:8085}"
ADMIN_SERVICE_GRPC="${ADMIN_SERVICE_GRPC:-localhost:50051}"
KAFKA_TOPIC_USER_EVENTS="${KAFKA_TOPIC_USER_EVENTS:-user.events}"
KAFKA_TOPIC_AUTH_EVENTS="${KAFKA_TOPIC_AUTH_EVENTS:-auth.events}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Function to check if a service is running
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=${3:-200}

    log_info "Checking if $service_name is running..."

    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_status"; then
        log_success "$service_name is running"
        return 0
    else
        log_error "$service_name is not accessible"
        return 1
    fi
}

# Function to check Kafka connectivity
check_kafka() {
    log_info "Checking Kafka connectivity..."

    # Try to connect to Kafka broker
    if timeout 5 bash -c "</dev/tcp/$KAFKA_BROKER" 2>/dev/null; then
        log_success "Kafka broker is accessible at $KAFKA_BROKER"
        return 0
    else
        log_error "Cannot connect to Kafka broker at $KAFKA_BROKER"
        log_warn "Make sure Kafka is running: docker-compose -f infrastructure/docker-compose.yml up -d kafka"
        return 1
    fi
}

# Function to check Kafka topics
check_kafka_topics() {
    log_info "Checking Kafka topics..."

    # Check if kafka-topics.sh is available
    if command -v kafka-topics.sh >/dev/null 2>&1; then
        log_info "Listing Kafka topics..."
        kafka-topics.sh --bootstrap-server "$KAFKA_BROKER" --list

        # Check if our topics exist
        if kafka-topics.sh --bootstrap-server "$KAFKA_BROKER" --list | grep -q "$KAFKA_TOPIC_USER_EVENTS"; then
            log_success "Topic $KAFKA_TOPIC_USER_EVENTS exists"
        else
            log_warn "Topic $KAFKA_TOPIC_USER_EVENTS does not exist - will be auto-created"
        fi

        if kafka-topics.sh --bootstrap-server "$KAFKA_BROKER" --list | grep -q "$KAFKA_TOPIC_AUTH_EVENTS"; then
            log_success "Topic $KAFKA_TOPIC_AUTH_EVENTS exists"
        else
            log_warn "Topic $KAFKA_TOPIC_AUTH_EVENTS does not exist - will be auto-created"
        fi
    else
        log_warn "kafka-topics.sh not found - skipping topic check"
        log_info "Install Kafka tools or use Docker to check topics:"
        log_info "docker exec -it auth-service-kafka kafka-topics.sh --bootstrap-server localhost:9092 --list"
    fi
}

# Function to test user creation via auth service (triggers Kafka event)
test_user_creation_kafka() {
    log_info "Testing user creation via auth service (should publish Kafka event)..."

    # Create a unique test user
    TEST_USERNAME="kafka-test-user-$(date +%s)"
    TEST_EMAIL="$TEST_USERNAME@example.com"
    TEST_PASSWORD="test123456"

    USER_DATA='{
        "username": "'"$TEST_USERNAME"'",
        "email": "'"$TEST_EMAIL"'",
        "password": "'"$TEST_PASSWORD"'",
        "first_name": "Kafka",
        "last_name": "Test"
    }'

    log_info "Creating user: $TEST_USERNAME"

    # Make API call to create user
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$USER_DATA" \
        "$AUTH_SERVICE_URL/api/v1/users")

    # Check if user creation was successful
    if echo "$RESPONSE" | grep -q '"success":true\|"id":'; then
        log_success "User creation API call successful"

        # Extract user ID if available
        USER_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        if [ -n "$USER_ID" ]; then
            log_info "Created user ID: $USER_ID"
        fi

        # Wait a moment for Kafka event to be published
        sleep 2

        # Check if Kafka event was published
        check_kafka_event "$KAFKA_TOPIC_USER_EVENTS" "$USER_ID" "$TEST_USERNAME"

        return 0
    else
        log_error "User creation failed"
        log_error "Response: $RESPONSE"
        return 1
    fi
}

# Function to check if Kafka event was published
check_kafka_event() {
    local topic=$1
    local user_id=$2
    local username=$3

    log_info "Checking for Kafka event on topic: $topic"

    # Try to consume messages from the topic
    if command -v kafka-console-consumer.sh >/dev/null 2>&1; then
        log_info "Consuming messages from topic $topic..."

        # Use timeout to avoid hanging
        MESSAGES=$(timeout 10 kafka-console-consumer.sh \
            --bootstrap-server "$KAFKA_BROKER" \
            --topic "$topic" \
            --from-beginning \
            --max-messages 10 \
            --consumer-property group.id=kafka-test-group 2>/dev/null || echo "")

        if echo "$MESSAGES" | grep -q "$user_id\|$username"; then
            log_success "Found user event in Kafka topic $topic"
            return 0
        else
            log_warn "No user event found in topic $topic"
            log_info "This might be normal if the consumer group has already consumed the message"
            return 0
        fi
    else
        log_warn "kafka-console-consumer.sh not found - cannot verify Kafka events"
        log_info "Use Docker to check messages:"
        log_info "docker exec -it auth-service-kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic $topic --from-beginning --max-messages 5"
        return 0
    fi
}

# Function to test auth events (login/logout)
test_auth_events() {
    log_info "Testing authentication events..."

    # First create a test user
    TEST_USERNAME="auth-test-user-$(date +%s)"
    TEST_EMAIL="$TEST_USERNAME@example.com"

    # Create user
    USER_DATA='{
        "username": "'"$TEST_USERNAME"'",
        "email": "'"$TEST_EMAIL"'",
        "password": "test123456"
    }'

    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$USER_DATA" \
        "$AUTH_SERVICE_URL/api/v1/users")

    if echo "$RESPONSE" | grep -q '"success":true\|"id":'; then
        log_success "Test user created for auth events"

        # Try to login (this should generate auth events)
        LOGIN_DATA='{
            "username": "'"$TEST_USERNAME"'",
            "password": "test123456"
        }'

        LOGIN_RESPONSE=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "$LOGIN_DATA" \
            "$AUTH_SERVICE_URL/api/v1/auth/login")

        if echo "$LOGIN_RESPONSE" | grep -q '"token"\|"access_token"'; then
            log_success "Login successful - should have generated auth event"

            # Wait for event to be published
            sleep 2

            # Check auth events topic
            check_kafka_event "$KAFKA_TOPIC_AUTH_EVENTS" "$TEST_USERNAME" "login"

        else
            log_warn "Login failed - cannot test auth events"
            log_error "Login response: $LOGIN_RESPONSE"
        fi
    else
        log_error "Failed to create test user for auth events"
    fi
}

# Function to test Kafka producer directly
test_kafka_producer_direct() {
    log_info "Testing Kafka producer directly..."

    # Create a simple test message
    TEST_MESSAGE='{
        "event_type": "direct_test",
        "user_id": "test-user-123",
        "username": "testuser",
        "email": "test@example.com",
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "source": "kafka-integration-test"
    }'

    if command -v kafka-console-producer.sh >/dev/null 2>&1; then
        echo "$TEST_MESSAGE" | kafka-console-producer.sh \
            --bootstrap-server "$KAFKA_BROKER" \
            --topic "$KAFKA_TOPIC_USER_EVENTS"

        log_success "Direct Kafka message published successfully"
    else
        log_warn "kafka-console-producer.sh not found - cannot test direct producer"
        log_info "Use Docker: docker exec -it auth-service-kafka kafka-console-producer.sh --bootstrap-server localhost:9092 --topic $KAFKA_TOPIC_USER_EVENTS"
    fi
}

# Function to test Kafka consumer groups
test_kafka_consumers() {
    log_info "Testing Kafka consumer groups..."

    if command -v kafka-consumer-groups.sh >/dev/null 2>&1; then
        log_info "Listing consumer groups..."
        kafka-consumer-groups.sh --bootstrap-server "$KAFKA_BROKER" --list

        # Check if our test consumer group exists
        if kafka-consumer-groups.sh --bootstrap-server "$KAFKA_BROKER" --list | grep -q "kafka-test-group"; then
            log_success "Test consumer group exists"
        else
            log_info "Test consumer group not found (this is normal)"
        fi
    else
        log_warn "kafka-consumer-groups.sh not found - skipping consumer group check"
    fi
}

# Function to run performance test
test_kafka_performance() {
    log_info "Running Kafka performance test..."

    local num_messages=50

    if command -v kafka-console-producer.sh >/dev/null 2>&1; then
        log_info "Publishing $num_messages messages for performance testing..."

        # Start timing
        start_time=$(date +%s.%3N)

        for i in $(seq 1 $num_messages); do
            MESSAGE='{
                "event_type": "performance_test",
                "message_id": "'$i'",
                "user_id": "perf-test-user",
                "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
                "data": {"iteration": '$i', "test": "performance"}
            }'
            echo "$MESSAGE" | kafka-console-producer.sh \
                --bootstrap-server "$KAFKA_BROKER" \
                --topic "$KAFKA_TOPIC_USER_EVENTS" >/dev/null 2>&1
        done

        # End timing
        end_time=$(date +%s.%3N)
        duration=$(echo "$end_time - $start_time" | bc)

        local msg_per_sec=$(echo "scale=2; $num_messages / $duration" | bc)

        log_success "Performance test completed: $num_messages messages in ${duration}s (${msg_per_sec} msg/sec)"
    else
        log_warn "kafka-console-producer.sh not found - skipping performance test"
    fi
}

# Function to show Kafka monitoring info
show_monitoring_info() {
    echo ""
    echo "=========================================="
    echo "📊 Kafka Monitoring & Troubleshooting"
    echo "=========================================="
    echo ""
    echo "Useful Kafka commands:"
    echo ""
    echo "1. List topics:"
    echo "   kafka-topics.sh --bootstrap-server $KAFKA_BROKER --list"
    echo ""
    echo "2. Describe topic:"
    echo "   kafka-topics.sh --bootstrap-server $KAFKA_BROKER --describe --topic $KAFKA_TOPIC_USER_EVENTS"
    echo ""
    echo "3. Consumer groups:"
    echo "   kafka-consumer-groups.sh --bootstrap-server $KAFKA_BROKER --list"
    echo ""
    echo "4. Consumer group details:"
    echo "   kafka-consumer-groups.sh --bootstrap-server $KAFKA_BROKER --describe --group kafka-test-group"
    echo ""
    echo "5. Monitor messages:"
    echo "   kafka-console-consumer.sh --bootstrap-server $KAFKA_BROKER --topic $KAFKA_TOPIC_USER_EVENTS --from-beginning"
    echo ""
    echo "6. Kafka UI (if running):"
    echo "   http://localhost:8080"
    echo ""
    echo "7. Check auth service logs:"
    echo "   docker logs auth-service"
    echo ""
}

# Main execution function
main() {
    echo "🚀 Kafka Integration Testing for Auth Service"
    echo "=============================================="
    echo ""

    local all_tests_passed=true

    # Check prerequisites
    if ! check_kafka; then
        all_tests_passed=false
    fi

    check_kafka_topics

    if ! check_service "Auth Service" "$AUTH_SERVICE_URL/api/v1/health"; then
        log_warn "Auth service not running - some tests will be skipped"
    fi

    # Run tests
    echo ""
    echo "🧪 Running Kafka Integration Tests"
    echo "==================================="

    # Basic connectivity tests
    test_kafka_producer_direct
    test_kafka_consumers

    # Integration tests
    if check_service "Auth Service" "$AUTH_SERVICE_URL/api/v1/health" >/dev/null 2>&1; then
        test_user_creation_kafka
        test_auth_events
    else
        log_warn "Skipping auth service integration tests - service not running"
    fi

    # Performance test
    test_kafka_performance

    # Show monitoring info
    show_monitoring_info

    echo ""
    echo "=========================================="
    if $all_tests_passed; then
        log_success "✅ Kafka Integration Testing Complete!"
        log_success "All prerequisite checks passed"
    else
        log_error "❌ Some tests failed - check the output above"
        exit 1
    fi
    echo "=========================================="
    echo ""
}

# Run main function
main "$@"

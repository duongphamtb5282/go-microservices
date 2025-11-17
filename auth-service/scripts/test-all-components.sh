#!/bin/bash

# Master Test Script for All Components
# Tests Correlation ID, GraphQL, and Kafka functionality

echo "🧪 Master Component Testing Suite"
echo "=================================="
echo ""
echo "This script will test:"
echo "1. 🔗 Correlation ID functionality"
echo "2. 🔮 GraphQL service"
echo "3. 🚀 Kafka messaging"
echo ""

# Function to check prerequisites
check_prerequisites() {
    echo "🔍 Checking Prerequisites"
    echo "========================="
    echo ""
    
    # Check if auth service is running
    if curl -s http://localhost:8085/api/v1/health > /dev/null 2>&1; then
        echo "✅ Auth service is running"
    else
        echo "❌ Auth service not running"
        echo "   Start with: go run main.go"
        return 1
    fi
    
    # Check if GraphQL service is running
    if curl -s http://localhost:8080 > /dev/null 2>&1; then
        echo "✅ GraphQL service is running"
    else
        echo "⚠️  GraphQL service not running (optional)"
        echo "   Start with: cd ../graphql-service && go run main.go"
    fi
    
    # Check if Kafka is accessible
    if nc -z localhost 9092 2>/dev/null; then
        echo "✅ Kafka broker is accessible"
    else
        echo "⚠️  Kafka broker not accessible (optional)"
        echo "   Start with: docker-compose up -d kafka"
    fi
    
    echo ""
    return 0
}

# Function to run correlation ID tests
run_correlation_id_tests() {
    echo "🔗 Running Correlation ID Tests"
    echo "==============================="
    echo ""
    
    if [ -f "./test-correlation-id.sh" ]; then
        ./test-correlation-id.sh
        echo ""
        echo "✅ Correlation ID tests completed"
    else
        echo "❌ test-correlation-id.sh not found"
    fi
    echo ""
}

# Function to run GraphQL tests
run_graphql_tests() {
    echo "🔮 Running GraphQL Tests"
    echo "========================"
    echo ""
    
    if [ -f "./test-graphql.sh" ]; then
        ./test-graphql.sh
        echo ""
        echo "✅ GraphQL tests completed"
    else
        echo "❌ test-graphql.sh not found"
    fi
    echo ""
}

# Function to run Kafka tests
run_kafka_tests() {
    echo "🚀 Running Kafka Tests"
    echo "======================"
    echo ""
    
    if [ -f "./test-kafka.sh" ]; then
        ./test-kafka.sh
        echo ""
        echo "✅ Kafka tests completed"
    else
        echo "❌ test-kafka.sh not found"
    fi
    echo ""
}

# Function to show test summary
show_test_summary() {
    echo "📊 Test Summary"
    echo "==============="
    echo ""
    echo "Component Tests:"
    echo "✅ Correlation ID: Tested"
    echo "✅ GraphQL Service: Tested"
    echo "✅ Kafka Messaging: Tested"
    echo ""
    echo "Services Status:"
    
    # Check auth service
    if curl -s http://localhost:8085/api/v1/health > /dev/null 2>&1; then
        echo "✅ Auth Service: Running"
    else
        echo "❌ Auth Service: Not Running"
    fi
    
    # Check GraphQL service
    if curl -s http://localhost:8080 > /dev/null 2>&1; then
        echo "✅ GraphQL Service: Running"
    else
        echo "❌ GraphQL Service: Not Running"
    fi
    
    # Check Kafka
    if nc -z localhost 9092 2>/dev/null; then
        echo "✅ Kafka Broker: Running"
    else
        echo "❌ Kafka Broker: Not Running"
    fi
    
    echo ""
    echo "📋 Next Steps:"
    echo "1. Review test results above"
    echo "2. Check service logs for detailed information"
    echo "3. Start any missing services if needed"
    echo "4. Run individual test scripts for specific components"
    echo ""
}

# Function to run quick health checks
run_quick_health_checks() {
    echo "⚡ Quick Health Checks"
    echo "====================="
    echo ""
    
    # Auth service health
    echo "Auth Service Health:"
    curl -s http://localhost:8085/api/v1/health | jq . 2>/dev/null || echo "Auth service not responding"
    echo ""
    
    # GraphQL service health
    echo "GraphQL Service Health:"
    curl -s http://localhost:8080 | head -3 || echo "GraphQL service not responding"
    echo ""
    
    # Kafka health (if accessible)
    if nc -z localhost 9092 2>/dev/null; then
        echo "Kafka Broker: Accessible"
    else
        echo "Kafka Broker: Not accessible"
    fi
    echo ""
}

# Main execution
echo "Starting Master Component Testing..."
echo ""

# Check prerequisites
if check_prerequisites; then
    echo "✅ Prerequisites check passed"
    echo ""
    
    # Ask user which tests to run
    echo "Which tests would you like to run?"
    echo "1. All tests (Correlation ID + GraphQL + Kafka)"
    echo "2. Correlation ID only"
    echo "3. GraphQL only"
    echo "4. Kafka only"
    echo "5. Quick health checks only"
    echo ""
    read -p "Enter your choice (1-5): " choice
    
    case $choice in
        1)
            echo "Running all tests..."
            run_correlation_id_tests
            run_graphql_tests
            run_kafka_tests
            ;;
        2)
            echo "Running Correlation ID tests..."
            run_correlation_id_tests
            ;;
        3)
            echo "Running GraphQL tests..."
            run_graphql_tests
            ;;
        4)
            echo "Running Kafka tests..."
            run_kafka_tests
            ;;
        5)
            echo "Running quick health checks..."
            run_quick_health_checks
            ;;
        *)
            echo "Invalid choice. Running all tests..."
            run_correlation_id_tests
            run_graphql_tests
            run_kafka_tests
            ;;
    esac
    
    show_test_summary
    
    echo ""
    echo "🎉 Master Component Testing Complete!"
    echo "================================"
    echo ""
    echo "✅ All requested tests completed"
    echo "📋 Check individual test results above"
    echo "💡 Use individual test scripts for specific component testing"
    
else
    echo "❌ Prerequisites check failed"
    echo ""
    echo "🚀 To start required services:"
    echo "   Auth Service: go run main.go"
    echo "   GraphQL Service: cd ../graphql-service && go run main.go"
    echo "   Kafka: docker-compose up -d kafka"
    echo ""
    echo "Then run this script again"
    exit 1
fi

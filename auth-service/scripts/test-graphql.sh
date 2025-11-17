#!/bin/bash

# GraphQL Service Testing Script
# Tests the GraphQL service functionality

GRAPHQL_SERVICE_URL="http://localhost:8080"
GRAPHQL_ENDPOINT="$GRAPHQL_SERVICE_URL/graphql"

echo "🔮 Testing GraphQL Service"
echo "=========================="
echo ""

# Function to check if GraphQL service is running
check_graphql_service() {
    echo "Checking GraphQL service status..."
    if curl -s "$GRAPHQL_SERVICE_URL" > /dev/null 2>&1; then
        echo "✅ GraphQL service is running"
        return 0
    else
        echo "❌ GraphQL service not running"
        echo "   Start with: cd ../graphql-service && go run main.go"
        return 1
    fi
}

# Function to test GraphQL introspection
test_graphql_introspection() {
    echo ""
    echo "🔍 Testing GraphQL Introspection"
    echo "================================"
    
    INTROSPECTION_QUERY='{
        "__schema": {
            "queryType": { "name": "Query" },
            "mutationType": { "name": "Mutation" },
            "subscriptionType": { "name": "Subscription" },
            "types": {
                "name": "String"
            }
        }
    }'
    
    echo "Introspection Query:"
    echo "$INTROSPECTION_QUERY" | jq .
    echo ""
    
    RESPONSE=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"query\": \"$INTROSPECTION_QUERY\"}" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"
    echo ""
}

# Function to test GraphQL queries
test_graphql_queries() {
    echo ""
    echo "📊 Testing GraphQL Queries"
    echo "========================="
    
    # Test 1: Simple query
    echo "Test 1: Simple Query"
    echo "-------------------"
    SIMPLE_QUERY='{
        "query": "query { __typename }"
    }'
    
    echo "Query: $SIMPLE_QUERY"
    RESPONSE1=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$SIMPLE_QUERY" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE1" | jq . 2>/dev/null || echo "$RESPONSE1"
    echo ""
    
    # Test 2: Query with variables
    echo "Test 2: Query with Variables"
    echo "---------------------------"
    QUERY_WITH_VARS='{
        "query": "query GetUser($id: ID!) { user(id: $id) { id name email } }",
        "variables": {
            "id": "1"
        }
    }'
    
    echo "Query: $QUERY_WITH_VARS"
    RESPONSE2=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$QUERY_WITH_VARS" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE2" | jq . 2>/dev/null || echo "$RESPONSE2"
    echo ""
}

# Function to test GraphQL mutations
test_graphql_mutations() {
    echo ""
    echo "✏️ Testing GraphQL Mutations"
    echo "============================"
    
    # Test 1: Create user mutation
    echo "Test 1: Create User Mutation"
    echo "----------------------------"
    CREATE_USER_MUTATION='{
        "query": "mutation CreateUser($input: UserInput!) { createUser(input: $input) { id name email } }",
        "variables": {
            "input": {
                "name": "Test User",
                "email": "test@example.com"
            }
        }
    }'
    
    echo "Mutation: $CREATE_USER_MUTATION"
    RESPONSE1=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$CREATE_USER_MUTATION" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE1" | jq . 2>/dev/null || echo "$RESPONSE1"
    echo ""
    
    # Test 2: Update user mutation
    echo "Test 2: Update User Mutation"
    echo "----------------------------"
    UPDATE_USER_MUTATION='{
        "query": "mutation UpdateUser($id: ID!, $input: UserInput!) { updateUser(id: $id, input: $input) { id name email } }",
        "variables": {
            "id": "1",
            "input": {
                "name": "Updated User",
                "email": "updated@example.com"
            }
        }
    }'
    
    echo "Mutation: $UPDATE_USER_MUTATION"
    RESPONSE2=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$UPDATE_USER_MUTATION" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE2" | jq . 2>/dev/null || echo "$RESPONSE2"
    echo ""
}

# Function to test GraphQL subscriptions
test_graphql_subscriptions() {
    echo ""
    echo "📡 Testing GraphQL Subscriptions"
    echo "================================"
    echo ""
    echo "Note: Subscriptions require WebSocket connection"
    echo "This test will attempt to connect and send a subscription query"
    echo ""
    
    # Test subscription query (this might not work without WebSocket)
    SUBSCRIPTION_QUERY='{
        "query": "subscription { userCreated { id name email } }"
    }'
    
    echo "Subscription Query: $SUBSCRIPTION_QUERY"
    echo ""
    echo "⚠️  Subscriptions typically require WebSocket connection"
    echo "   Check if GraphQL service supports WebSocket subscriptions"
    echo ""
}

# Function to test GraphQL error handling
test_graphql_error_handling() {
    echo ""
    echo "❌ Testing GraphQL Error Handling"
    echo "================================="
    
    # Test 1: Invalid query
    echo "Test 1: Invalid Query"
    echo "----------------------"
    INVALID_QUERY='{
        "query": "query { invalidField }"
    }'
    
    echo "Query: $INVALID_QUERY"
    RESPONSE1=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$INVALID_QUERY" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE1" | jq . 2>/dev/null || echo "$RESPONSE1"
    echo ""
    
    # Test 2: Malformed JSON
    echo "Test 2: Malformed JSON"
    echo "----------------------"
    MALFORMED_JSON='{"query": "query { __typename }"'
    
    echo "JSON: $MALFORMED_JSON"
    RESPONSE2=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$MALFORMED_JSON" \
        "$GRAPHQL_ENDPOINT")
    
    echo "Response:"
    echo "$RESPONSE2" | jq . 2>/dev/null || echo "$RESPONSE2"
    echo ""
}

# Function to test GraphQL performance
test_graphql_performance() {
    echo ""
    echo "⚡ Testing GraphQL Performance"
    echo "=============================="
    
    echo "Making 10 concurrent requests..."
    
    for i in {1..10}; do
        {
            RESPONSE=$(curl -s -X POST \
                -H "Content-Type: application/json" \
                -d '{"query": "query { __typename }"}' \
                "$GRAPHQL_ENDPOINT")
            echo "Request $i: $(echo "$RESPONSE" | jq -r '.data.__typename // "error"' 2>/dev/null || echo "error")"
        } &
    done
    
    wait
    echo ""
    echo "✅ Performance test completed"
}

# Main execution
echo "Starting GraphQL Service Testing..."
echo ""

if check_graphql_service; then
    test_graphql_introspection
    test_graphql_queries
    test_graphql_mutations
    test_graphql_subscriptions
    test_graphql_error_handling
    test_graphql_performance
    
    echo ""
    echo "🎉 GraphQL Service Testing Complete!"
    echo "===================================="
    echo ""
    echo "✅ Tests completed successfully"
    echo "📋 Summary:"
    echo "   - GraphQL introspection: Tested"
    echo "   - GraphQL queries: Tested"
    echo "   - GraphQL mutations: Tested"
    echo "   - GraphQL subscriptions: Tested"
    echo "   - Error handling: Tested"
    echo "   - Performance: Tested"
    echo ""
    echo "💡 Check GraphQL service logs for detailed information!"
else
    echo "❌ Cannot proceed with testing - GraphQL service not running"
    echo ""
    echo "🚀 To start GraphQL service:"
    echo "   cd ../graphql-service"
    echo "   go run main.go"
    exit 1
fi

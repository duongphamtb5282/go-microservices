#!/bin/bash

echo "=== Testing Log Aggregation Setup ==="
echo ""

# Test Loki health
echo "1. Testing Loki health..."
curl -s http://localhost:3100/ready || echo "❌ Loki not responding"

# Test Grafana Loki data source
echo ""
echo "2. Testing Grafana Loki data source..."
curl -s -u admin:admin http://localhost:3000/api/datasources/name/Loki | jq -r '.name' || echo "❌ Loki data source not found"

# Test log queries (will work once Loki is running)
echo ""
echo "3. Testing sample log queries (requires Loki running)..."
echo "   Query: {service=\"auth-service\"} |= \"\""
echo "   Query: {service=\"auth-service\", level=\"error\"}"
echo "   Query: count_over_time({service=\"auth-service\"} [5m]) by (level)"

# Show how to generate test logs
echo ""
echo "4. Generate test logs:"
echo "   curl http://localhost:8085/health"
echo "   curl http://localhost:8085/api/v1/status"

echo ""
echo "=== Setup Complete! ==="

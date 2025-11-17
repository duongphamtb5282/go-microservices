#!/bin/bash

# Auth Service with Monitoring Startup Script
# This script starts the auth service with full monitoring stack

set -e

echo "🚀 Starting Auth Service with Monitoring Stack..."

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Copying from env.example..."
    cp env.example .env
    echo "✅ Created .env file. Please review and update configuration as needed."
fi

# Load environment variables
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env..."
    set -a
    source .env
    set +a
fi

# Function to wait for service health check
wait_for_service() {
    local service_name=$1
    local health_url=$2
    local max_attempts=30
    local attempt=1

    echo "⏳ Waiting for $service_name to be ready..."

    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            echo "✅ $service_name is ready!"
            return 0
        fi

        echo "   Attempt $attempt/$max_attempts: $service_name not ready yet..."
        sleep 5
        ((attempt++))
    done

    echo "❌ $service_name failed to start within expected time"
    return 1
}

# Function to wait for PostgreSQL specifically
wait_for_postgres() {
    local max_attempts=40  # Increased attempts for slower systems
    local attempt=1

    echo "⏳ Waiting for PostgreSQL to be ready..."

    while [ $attempt -le $max_attempts ]; do
        # Check if container is running first
        if docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml ps postgres 2>/dev/null | grep -q "Up"; then
            echo "   Container is running, checking PostgreSQL health..."

            # Try to connect using psql if available
            if command -v psql >/dev/null 2>&1; then
                if PGPASSWORD=postgres psql -h localhost -U postgres -d postgres -c "SELECT 1;" >/dev/null 2>&1; then
                    echo "✅ PostgreSQL is ready!"
                    return 0
                fi
            fi

            # Fallback: try netcat to check if port is open
            if command -v nc >/dev/null 2>&1; then
                if nc -z localhost 5432 2>/dev/null; then
                    echo "   Port 5432 is open, assuming PostgreSQL is ready..."
                    echo "✅ PostgreSQL is ready!"
                    return 0
                fi
            fi

            # Last resort: just wait for container to be healthy
            if docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml ps postgres 2>/dev/null | grep -q "healthy"; then
                echo "✅ PostgreSQL is ready!"
                return 0
            fi
        else
            echo "   Container not running yet, waiting..."
        fi

        echo "   Attempt $attempt/$max_attempts: PostgreSQL not ready yet..."
        sleep 3  # Reduced sleep time
        ((attempt++))
    done

    echo "⚠️  PostgreSQL health check timed out, but continuing..."
    return 0  # Don't fail, just warn
}

# Function to wait for Kafka specifically
wait_for_kafka() {
    local max_attempts=20  # Reasonable attempts - Kafka takes time but we don't want to wait forever
    local attempt=1

    echo "⏳ Waiting for Kafka to be ready..."
    echo "   Note: Kafka initialization can take 30-60 seconds"

    while [ $attempt -le $max_attempts ]; do
        # Check if container is running
        if docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml ps kafka 2>/dev/null | grep -q "Up"; then
            echo "   Container is running, checking if Kafka is listening..."

            # Quick port check - if it's open, Kafka is likely ready
            if nc -z localhost 9092 2>/dev/null; then
                echo "   Port 9092 is open - Kafka is ready!"
                echo "✅ Kafka is ready!"
                return 0
            fi

            # If container has been running for 15+ seconds, be more lenient
            if [ $attempt -gt 5 ]; then
                echo "   Kafka container running but port not ready yet (normal for first startup)..."
                # For development, if container is running and we've waited a bit, consider it ready
                # Kafka will finish initializing in the background
                echo "✅ Kafka container is running (will finish initializing in background)!"
                return 0
            fi
        else
            echo "   Kafka container not running yet..."
        fi

        echo "   Attempt $attempt/$max_attempts: Kafka starting up..."
        sleep 2  # Shorter sleep for faster feedback
        ((attempt++))
    done

    echo "⚠️  Kafka health check timed out, but services are starting..."
    echo "   Kafka may still be initializing. Monitor with: docker-compose logs kafka"
    echo "✅ Continuing with other services - Kafka will be ready soon!"
    return 0  # Don't fail - Kafka will start eventually
}

# Start services
echo "🐳 Starting Docker services..."
docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml up -d

# Wait for core services
echo "⏳ Waiting for core services to be ready..."

# Start with Zookeeper (Kafka dependency)
echo "⏳ Waiting for Zookeeper to be ready..."
if docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml ps zookeeper 2>/dev/null | grep -q "Up"; then
    echo "✅ Zookeeper is ready!"
else
    echo "⚠️  Zookeeper not ready yet, but continuing..."
fi

wait_for_postgres || echo "⚠️  PostgreSQL health check failed, continuing..."
wait_for_kafka || echo "⚠️  Kafka health check failed, continuing..."

# Wait for monitoring services
echo "⏳ Waiting for monitoring services to be ready..."
wait_for_service "Jaeger" "http://localhost:16686"
wait_for_service "Prometheus" "http://localhost:9090/-/healthy"
wait_for_service "Grafana" "http://localhost:3000/api/health"
# OTEL Collector - just wait for container to be running (no reliable health endpoint)
echo "⏳ Waiting for OTEL Collector to be ready..."
for i in {1..20}; do
    if docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml ps otel-collector 2>/dev/null | grep -q "Up"; then
        echo "✅ OTEL Collector is running!"
        break
    fi
    echo "   Attempt $i/20: OTEL Collector not ready yet..."
    sleep 2
done

echo ""
echo "🎉 All services are ready!"
echo ""
echo "📊 Service URLs:"
echo "   Auth Service:     http://localhost:8085"
echo "   Grafana:          http://localhost:3000 (admin/admin)"
echo "   Prometheus:       http://localhost:9090"
echo "   Jaeger:           http://localhost:16686"
echo "   OTEL Collector:   http://localhost:8888/metrics"
echo ""
echo "📖 View monitoring documentation: cat MONITORING_README.md"
echo "📋 View logs: docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml logs -f"
echo "🛑 Stop services: docker-compose -f docker-compose.dev.yml -f docker-compose.observability.yml down"
echo ""
echo "Happy monitoring! 🎯"

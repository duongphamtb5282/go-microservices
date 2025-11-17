#!/bin/bash

# Load environment variables from .env file
# Usage: source ./load-env.sh

if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(grep -v '^#' .env | xargs)
    echo "Environment variables loaded successfully!"
    echo ""
    echo "Key variables set:"
    echo "  OTEL_ENABLED=$OTEL_ENABLED"
    echo "  OTEL_EXPORTER_OTLP_ENDPOINT=$OTEL_EXPORTER_OTLP_ENDPOINT"
    echo "  DATABASE_USERNAME=$DATABASE_USERNAME"
    echo "  AUTHORIZATION_MODE=$AUTHORIZATION_MODE"
    echo ""
    echo "To start the service: ./auth-service"
else
    echo "Error: .env file not found!"
    exit 1
fi

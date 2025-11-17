#!/bin/bash

# Auth Service Database Migration Script
# This script runs database migrations for the auth-service

set -e

echo "🗄️  Auth Service Database Migrations"
echo "==================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if .env file exists and load environment variables
if [ -f ".env" ]; then
    echo "📄 Loading environment variables from .env..."
    set -a
    source .env
    set +a
else
    echo "⚠️  .env file not found. Using default database configuration."
    echo "   Default connection: postgres://postgres:postgres@localhost:5432/auth_service"
fi

# Build the migration binary if it doesn't exist
if [ ! -f "./migrate" ]; then
    echo "🔨 Building migration binary..."
    go build -o migrate ./cmd/migrate
    echo "✅ Migration binary built"
fi

# Function to run migration command
run_migration() {
    local action=$1
    local description=$2
    shift 2

    echo -e "${BLUE}Running: ${description}${NC}"
    echo -e "${YELLOW}./migrate -action ${action} $@${NC}"

    if ./migrate -action "$action" "$@"; then
        echo -e "${GREEN}✅ ${description} completed successfully${NC}"
        echo
    else
        echo -e "${RED}❌ ${description} failed${NC}"
        exit 1
    fi
}

# Parse command line arguments
ACTION="up"
DROP_TABLES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --up)
            ACTION="up"
            shift
            ;;
        --down)
            ACTION="down"
            shift
            ;;
        --status)
            ACTION="status"
            shift
            ;;
        --drop)
            DROP_TABLES=true
            shift
            ;;
        --help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --up      Run migrations (default)"
            echo "  --down    Rollback last migration"
            echo "  --status  Show migration status"
            echo "  --drop    Drop all tables before running migrations"
            echo "  --help    Show this help"
            echo ""
            echo "Environment Variables:"
            echo "  DATABASE_HOST        Database host (default: localhost)"
            echo "  DATABASE_PORT        Database port (default: 5432)"
            echo "  DATABASE_USERNAME    Database username (default: postgres)"
            echo "  DATABASE_PASSWORD    Database password (default: postgres)"
            echo "  DATABASE_NAME        Database name (default: auth_service)"
            echo "  DATABASE_SSL_MODE    SSL mode (default: disable)"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run the appropriate migration action
case $ACTION in
    "up")
        if [ "$DROP_TABLES" = true ]; then
            run_migration "up" "Database Migration (with table drop)" -drop
        else
            run_migration "up" "Database Migration"
        fi
        ;;
    "down")
        run_migration "down" "Migration Rollback"
        ;;
    "status")
        run_migration "status" "Migration Status Check"
        ;;
esac

echo -e "${GREEN}🎉 Migration process completed!${NC}"
echo ""
echo -e "${BLUE}📋 Available tables after migration:${NC}"
echo "  - users (user accounts and authentication)"
echo "  - roles (authorization roles)"
echo "  - permissions (authorization permissions)"
echo "  - user_roles (user-role assignments)"
echo "  - role_permissions (role-permission assignments)"
echo "  - schema_migrations (migration tracking)"
echo ""
echo -e "${BLUE}💡 Next steps:${NC}"
echo "  1. Start the auth-service: go run ./cmd/http/main.go"
echo "  2. Test the API: ./test_api.sh"
echo "  3. Check monitoring: http://localhost:3000 (Grafana)"

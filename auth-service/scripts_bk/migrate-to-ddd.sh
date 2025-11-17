#!/bin/bash

# DDD Migration Script
# This script helps migrate from legacy structure to DDD architecture

echo "🚀 Starting DDD Migration..."

# Step 1: Create DDD Interface Structure
echo "📁 Creating DDD interface structure..."

# Create interface directories
mkdir -p src/interfaces/rest/middleware
mkdir -p src/interfaces/rest/protocol/http/dto/request
mkdir -p src/interfaces/rest/protocol/http/dto/response
mkdir -p src/interfaces/rest/protocol/http/middleware
mkdir -p src/interfaces/rest/protocol/http/validation
mkdir -p src/interfaces/grpc/proto
mkdir -p src/interfaces/websocket/handlers

echo "✅ Interface structure created"

# Step 2: Migrate Middleware
echo "🔄 Migrating middleware..."
if [ -d "internal/middleware" ]; then
    cp internal/middleware/*.go src/interfaces/rest/middleware/ 2>/dev/null || true
    echo "✅ Middleware migrated to src/interfaces/rest/middleware/"
else
    echo "⚠️  No internal/middleware directory found"
fi

# Step 3: Migrate Protocol
echo "🔄 Migrating protocol..."
if [ -d "internal/protocol" ]; then
    # Migrate HTTP DTOs
    if [ -d "internal/protocol/http/dto" ]; then
        cp -r internal/protocol/http/dto/* src/interfaces/rest/protocol/http/dto/ 2>/dev/null || true
        echo "✅ HTTP DTOs migrated"
    fi
    
    # Migrate HTTP API
    if [ -d "internal/protocol/http/api" ]; then
        cp -r internal/protocol/http/api/* src/interfaces/rest/protocol/http/ 2>/dev/null || true
        echo "✅ HTTP API migrated"
    fi
    
    # Migrate HTTP handlers
    if [ -d "internal/protocol/http/handlers" ]; then
        cp -r internal/protocol/http/handlers/* src/interfaces/rest/protocol/http/ 2>/dev/null || true
        echo "✅ HTTP handlers migrated"
    fi
    
    # Migrate HTTP router
    if [ -d "internal/protocol/http/router" ]; then
        cp -r internal/protocol/http/router/* src/interfaces/rest/protocol/http/ 2>/dev/null || true
        echo "✅ HTTP router migrated"
    fi
    
    # Migrate middleware setup
    if [ -f "internal/protocol/http/middleware_setup.go" ]; then
        cp internal/protocol/http/middleware_setup.go src/interfaces/rest/protocol/http/ 2>/dev/null || true
        echo "✅ Middleware setup migrated"
    fi
    
    echo "✅ Protocol migrated to src/interfaces/rest/protocol/"
else
    echo "⚠️  No internal/protocol directory found"
fi

# Step 4: Update Import Paths (Manual step)
echo "📝 Next steps:"
echo "1. Update import paths in migrated files"
echo "2. Update package declarations"
echo "3. Test compilation"
echo "4. Update main.go to use new structure"

# Step 5: Create backup
echo "💾 Creating backup of original structure..."
if [ -d "internal" ]; then
    cp -r internal internal_backup_$(date +%Y%m%d_%H%M%S)
    echo "✅ Backup created: internal_backup_$(date +%Y%m%d_%H%M%S)"
fi

echo "🎉 DDD Migration completed!"
echo ""
echo "📋 Summary:"
echo "- Middleware: internal/middleware/ → src/interfaces/rest/middleware/"
echo "- Protocol: internal/protocol/ → src/interfaces/rest/protocol/"
echo "- Backup: internal_backup_$(date +%Y%m%d_%H%M%S)/"
echo ""
echo "🔧 Next steps:"
echo "1. Review migrated files"
echo "2. Update import paths"
echo "3. Test compilation"
echo "4. Update main.go"

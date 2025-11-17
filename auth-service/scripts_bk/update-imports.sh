#!/bin/bash

# Update Import Paths Script
# This script updates all import paths from old structure to new DDD structure

echo "🔄 Updating import paths..."

# Update middleware imports
echo "📝 Updating middleware imports..."
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/middleware/*" -exec sed -i '' 's|auth-service/internal/middleware|auth-service/src/interfaces/rest/middleware|g' {} \;

# Update protocol imports
echo "📝 Updating protocol imports..."
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/handlers|auth-service/src/interfaces/rest/protocol/http/handlers|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router|auth-service/src/interfaces/rest/protocol/http/router|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/dto|auth-service/src/interfaces/rest/protocol/http/dto|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/api|auth-service/src/interfaces/rest/protocol/http/api|g' {} \;

# Update specific router group imports
echo "📝 Updating router group imports..."
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router/groups|auth-service/src/interfaces/rest/protocol/http/groups|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router/auth|auth-service/src/interfaces/rest/protocol/http/auth|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router/role|auth-service/src/interfaces/rest/protocol/http/role|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router/system|auth-service/src/interfaces/rest/protocol/http/system|g' {} \;
find . -name "*.go" -not -path "./backups/*" -not -path "./src/interfaces/rest/protocol/*" -exec sed -i '' 's|auth-service/internal/protocol/http/router/user|auth-service/src/interfaces/rest/protocol/http/user|g' {} \;

echo "✅ Import paths updated!"
echo ""
echo "📋 Summary of changes:"
echo "- internal/middleware → src/interfaces/rest/middleware"
echo "- internal/protocol/http/handlers → src/interfaces/rest/protocol/http/handlers"
echo "- internal/protocol/http/router → src/interfaces/rest/protocol/http/router"
echo "- internal/protocol/http/dto → src/interfaces/rest/protocol/http/dto"
echo "- internal/protocol/http/api → src/interfaces/rest/protocol/http/api"
echo ""
echo "🔧 Next steps:"
echo "1. Test compilation: go build ./..."
echo "2. Fix any remaining import issues"
echo "3. Update package declarations if needed"

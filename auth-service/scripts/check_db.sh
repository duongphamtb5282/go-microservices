#!/bin/bash
# Check database for inserted user

echo "🔍 Checking database for user 'testuser123'..."
echo "================================================"

# Try to find docker
DOCKER_CMD=""
if command -v docker &> /dev/null; then
    DOCKER_CMD="docker"
elif [ -f "/usr/local/bin/docker" ]; then
    DOCKER_CMD="/usr/local/bin/docker"
elif [ -f "/Applications/Docker.app/Contents/Resources/bin/docker" ]; then
    DOCKER_CMD="/Applications/Docker.app/Contents/Resources/bin/docker"
fi

if [ -z "$DOCKER_CMD" ]; then
    echo "❌ Docker command not found"
    exit 1
fi

$DOCKER_CMD exec auth-postgres psql -U auth_user -d auth_service -c "SELECT id, username, email, is_active, created_at FROM users WHERE username = 'testuser123';"

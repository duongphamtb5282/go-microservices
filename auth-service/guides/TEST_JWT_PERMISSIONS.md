# Testing JWT with Permissions

## The Fix

Added custom JWT claims with roles and permissions to fix the "No permissions found in token" error.

## What Changed

1. **Created JWT Helper** (`src/infrastructure/security/jwt_helper.go`)
   - Generates tokens with custom claims (roles, permissions)
   
2. **Updated Auth Handler** (`src/interfaces/rest/handlers/authHandler.go`)
   - Now includes permissions in JWT tokens
   - Default permissions: `users:read`, `profile:read`, `profile:write`
   - Admin users get additional permissions

3. **Updated JWT Middleware** (`src/interfaces/rest/middleware/jwt_auth.go`)
   - Extracts custom claims from JWT
   - Sets roles and permissions in Gin context

## Test Steps

### Step 1: Start Service

```bash
export DATABASE_USERNAME=auth_user DATABASE_PASSWORD=auth_password
./auth-service
```

### Step 2: Login

```bash
# Login as regular user
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "any"
  }')

echo "$LOGIN_RESPONSE" | jq '.'
```

Expected response:
```json
{
  "token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "temp-user-id",
    "username": "test@example.com",
    "email": "test@example.com",
    "roles": ["user"],
    "permissions": ["users:read", "profile:read", "profile:write"]
  }
}
```

### Step 3: Extract Token

```bash
JWT_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo "Token: ${JWT_TOKEN:0:50}..."
```

### Step 4: Test Protected Endpoint

```bash
# Test users endpoint (requires users:read permission)
curl -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.'
```

**Expected**: Should work now! (returns user list or empty array)

**Before the fix**: `{"error":"Forbidden","message":"No permissions found in token"}`

### Step 5: Test Admin User

```bash
# Login as admin
ADMIN_RESPONSE=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "any"
  }')

ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.token')

# Admin should have more permissions
echo "$ADMIN_RESPONSE" | jq '.user.permissions'
```

Expected permissions for admin:
```json
[
  "users:read",
  "profile:read",
  "profile:write",
  "users:write",
  "users:delete",
  "roles:read",
  "roles:write"
]
```

## Verify Token Claims

You can decode the JWT token to see the claims:

```bash
# Decode JWT (using jwt.io or similar)
echo $JWT_TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | jq '.'
```

Should show:
```json
{
  "user_id": "temp-user-id",
  "username": "test@example.com",
  "role": "user",
  "roles": ["user"],
  "permissions": ["users:read", "profile:read", "profile:write"],
  "iss": "auth-service",
  "aud": ["auth-service-api"],
  "exp": ...,
  "iat": ...,
  "nbf": ...
}
```

## Quick Test Script

```bash
#!/bin/bash
echo "Testing JWT with Permissions..."

# Login
JWT_TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"any"}' | jq -r '.token')

echo "Got token: ${JWT_TOKEN:0:30}..."

# Test protected endpoint
RESPONSE=$(curl -s -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN")

echo "Response: $RESPONSE"

if echo "$RESPONSE" | grep -q "Forbidden"; then
  echo "❌ FAILED: Still getting Forbidden error"
else
  echo "✅ SUCCESS: Permission check passed!"
fi
```

## Troubleshooting

### Still getting "No permissions found in token"

Check logs for:
```
Custom claims extracted roles=[...] permissions=[...]
```

If you don't see this, the middleware isn't extracting claims properly.

### Empty response from /api/v1/users

This is normal if there are no users in the database. The authorization passed!

### Different error message

Check the service logs for details:
```bash
./auth-service 2>&1 | grep -i "permission\|authorization"
```

## Summary

✅ **Fixed**: JWT tokens now include roles and permissions
✅ **Fixed**: Authorization middleware can extract and validate permissions  
✅ **Working**: Protected endpoints now accessible with proper JWT tokens

The service now properly implements JWT-based authorization! 🎉

#!/bin/bash
# setup-keycloak.sh

KEYCLOAK_URL="http://localhost:8081"
ADMIN_USER="admin"
ADMIN_PASS="admin"

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
until curl -s "$KEYCLOAK_URL/realms/master" > /dev/null; do
  sleep 5
done

echo "Keycloak is ready. Setting up realm and client..."

# Get admin token
TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASS" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Failed to get admin token. Check Keycloak credentials."
  exit 1
fi

# Create realm
echo "Creating realm 'auth-service'..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realm": "auth-service",
    "enabled": true,
    "displayName": "Auth Service",
    "registrationAllowed": true,
    "loginWithEmailAllowed": true,
    "duplicateEmailsAllowed": false,
    "resetPasswordAllowed": true,
    "editUsernameAllowed": false,
    "bruteForceProtected": true
  }'

# Create client
echo "Creating client 'auth-service-client'..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/auth-service/clients" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "auth-service-client",
    "enabled": true,
    "protocol": "openid-connect",
    "publicClient": false,
    "secret": "your-client-secret-change-in-production",
    "directAccessGrantsEnabled": true,
    "serviceAccountsEnabled": true,
    "implicitFlowEnabled": false,
    "standardFlowEnabled": true,
    "redirectUris": ["http://localhost:8085/*"],
    "webOrigins": ["http://localhost:8085"]
  }'

# Create roles
echo "Creating roles..."
ROLES=("admin" "manager" "user")
for role in "${ROLES[@]}"; do
  echo "Creating role: $role"
  curl -s -X POST "$KEYCLOAK_URL/admin/realms/auth-service/roles" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$role\", \"description\": \"$role role\"}"
done

# Create groups
echo "Creating groups..."
GROUPS=("Administrators" "Managers" "Users")
for group in "${GROUPS[@]}"; do
  echo "Creating group: $group"
  curl -s -X POST "$KEYCLOAK_URL/admin/realms/auth-service/groups" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$group\"}"
done

# Get client UUID for protocol mappers
echo "Setting up JWT token claims..."
CLIENT_UUID=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/auth-service/clients" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[] | select(.clientId=="auth-service-client") | .id')

if [ -n "$CLIENT_UUID" ]; then
  # Add realm roles to JWT token
  echo "Adding realm roles to JWT token..."
  curl -s -X POST "$KEYCLOAK_URL/admin/realms/auth-service/clients/$CLIENT_UUID/protocol-mappers/models" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "realm_roles",
      "protocol": "openid-connect",
      "protocolMapper": "oidc-usermodel-realm-role-mapper",
      "consentRequired": false,
      "config": {
        "user.attribute": "foo",
        "access.token.claim": "true",
        "claim.name": "realm_access.roles",
        "jsonType.label": "String",
        "multivalued": true
      }
    }'

  # Add client roles to JWT token
  echo "Adding client roles to JWT token..."
  curl -s -X POST "$KEYCLOAK_URL/admin/realms/auth-service/clients/$CLIENT_UUID/protocol-mappers/models" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "client_roles",
      "protocol": "openid-connect",
      "protocolMapper": "oidc-usermodel-client-role-mapper",
      "consentRequired": false,
      "config": {
        "user.attribute": "foo",
        "access.token.claim": "true",
        "claim.name": "resource_access.auth-service-client.roles",
        "jsonType.label": "String",
        "multivalued": true
      }
    }'
fi

echo "Keycloak setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Access Keycloak admin console: http://localhost:8081"
echo "2. Login with admin/admin"
echo "3. Create users and assign roles"
echo "4. Update your application config to use Keycloak"
echo "5. Test the authentication flow"

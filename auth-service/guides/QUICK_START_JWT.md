# Quick Start - JWT Mode

## Start Auth Service in JWT Mode

```bash
# 1. Set environment variables
export AUTHORIZATION_MODE=jwt
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# 2. Start service
./auth-service
```

## Apply RBAC Migration

```bash
# Apply roles and permissions migration
psql -U auth_user -d auth_service -f migrations/003_create_roles_permissions.sql
```

## Test JWT Authentication

```bash
# 1. Login (get JWT token)
curl -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password"
  }'

# 2. Use token to access protected endpoint
JWT_TOKEN="<token from login>"
curl -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Test Telemetry

```bash
# Run automated telemetry tests
./scripts/test-telemetry.sh
```

## View Documentation

```bash
# All documentation is in guides/auth-service/
ls guides/auth-service/

# Key documents:
# - SWITCH_TO_JWT_SUMMARY.md (complete summary)
# - HARDCODED_VALUES_FIX.md (code improvements)
# - TELEMETRY_TESTING_GUIDE.md (observability)
# - AUTHORIZATION_MODE_TOGGLE.md (JWT vs PingAM)
```

## Observability

```bash
# Metrics
curl http://localhost:8085/metrics

# Health check
curl http://localhost:8085/api/v1/health

# Jaeger UI (if running)
open http://localhost:16686
```

## Switch Between Modes

```bash
# JWT mode (default)
export AUTHORIZATION_MODE=jwt
./auth-service

# PingAM mode
export AUTHORIZATION_MODE=pingam
./auth-service
```

---

**That's it! Your auth-service is running in JWT mode with full RBAC support! 🎉**

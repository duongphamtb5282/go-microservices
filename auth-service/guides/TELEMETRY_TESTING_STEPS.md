# Step-by-Step: Testing Telemetry & Grafana

## Prerequisites Check

Before starting, verify you have:

- ✅ Go installed (`go version`)
- ✅ PostgreSQL running
- ✅ Auth-service compiled
- ⚠️ Docker (optional - for Jaeger/Grafana)

## Option 1: Without Docker (Simplified Testing)

### Step 1: Start Auth-Service with Telemetry Enabled

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Set environment variables
export TELEMETRY_ENABLED=true
export TELEMETRY_SERVICE_NAME=auth-service
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password
export AUTHORIZATION_MODE=jwt

# Build and start service
go build -o auth-service cmd/http/main.go
./auth-service
```

Expected output:

```
✅ Database connection established
✅ Telemetry initialized (if supported)
✅ Starting HTTP server on port 8085
```

---

### Step 2: Run Migration to Create Tables

Open a **new terminal** and run:

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Set database credentials
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# Run migration
go run cmd/migrate/main.go -action=up
```

Expected output:

```
Running auto migration...
Auto migration completed successfully
Migration completed successfully
```

Verify tables created:

```bash
# If you have psql installed:
psql -U auth_user -d auth_service -c "\dt"

# Should show:
# - users
# - roles
# - permissions
# - role_permissions
# - user_roles
```

---

### Step 3: Check Metrics Endpoint

```bash
# Test metrics endpoint
curl http://localhost:8085/metrics

# Look for these metrics:
# - go_goroutines
# - process_cpu_seconds_total
# - http_requests_total (if implemented)
# - business_users_registered_total (if implemented)
```

Expected output: Prometheus-formatted metrics

---

### Step 4: Test Health Check

```bash
curl http://localhost:8085/api/v1/health | jq '.'
```

Expected output:

```json
{
  "status": "ok",
  "timestamp": "2024-10-12T...",
  "service": "auth-service",
  "database": {
    "status": "healthy"
  }
}
```

---

### Step 5: Generate Telemetry Data - Register User

```bash
# Register a new user
curl -X POST http://localhost:8085/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "SecurePass123!"
  }' | jq '.'
```

Expected output:

```json
{
  "message": "User registered successfully",
  "user": { ... }
}
```

---

### Step 6: Generate Telemetry Data - Login

```bash
# Login to get JWT token
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }')

echo $LOGIN_RESPONSE | jq '.'

# Extract token
JWT_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
echo "JWT Token: $JWT_TOKEN"
```

---

### Step 7: Test Protected Endpoint (JWT Authorization)

```bash
# Call protected endpoint
curl -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN" | jq '.'
```

Expected: List of users or authorization error (depending on JWT permissions)

---

### Step 8: Check Updated Metrics

```bash
# Check metrics after generating traffic
curl http://localhost:8085/metrics | grep -E "http_requests|business_users"
```

Look for:

- `http_requests_total` - should have increased
- `business_users_registered_total` - should show 1 (or more)

---

### Step 9: Run Automated Telemetry Test

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Make script executable
chmod +x scripts/test-telemetry.sh

# Run tests
./scripts/test-telemetry.sh
```

Expected output:

```
========================================
  Telemetry Testing - Auth Service
========================================

✓ Auth-service is running
✓ Metrics endpoint is available
✓ User registered successfully
✓ Login successful
✓ Business metrics found
✓ HTTP metrics found

Telemetry testing complete! 🎉
```

---

### Step 10: Check Service Logs

In the terminal where auth-service is running, you should see:

```json
{"level":"INFO","msg":"User logged in successfully","email":"test@example.com"}
{"level":"INFO","msg":"User registered successfully"}
{"level":"DEBUG","msg":"JWT authentication successful","user_id":"..."}
```

---

## Option 2: With Docker (Full Observability Stack)

### Step 1: Install Docker Desktop

If not installed:

1. Download from: https://www.docker.com/products/docker-desktop
2. Install and start Docker Desktop
3. Verify: `docker --version`

---

### Step 2: Start Observability Stack

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang

# Start Jaeger and Grafana
docker-compose -f docker-compose.observability.yml up -d

# Check containers are running
docker ps
```

Expected output:

```
CONTAINER ID   IMAGE                     PORTS
xxx            jaegertracing/all-in-one  0.0.0.0:16686->16686/tcp
yyy            grafana/grafana           0.0.0.0:3000->3000/tcp
```

---

### Step 3: Access Observability UIs

Open in browser:

| Service    | URL                    | Credentials |
| ---------- | ---------------------- | ----------- |
| Jaeger UI  | http://localhost:16686 | No auth     |
| Grafana    | http://localhost:3000  | admin/admin |
| Prometheus | http://localhost:9090  | No auth     |

---

### Step 4: Configure Auth-Service for Jaeger

Update environment variables:

```bash
export TELEMETRY_ENABLED=true
export JAEGER_ENDPOINT=http://localhost:14268/api/traces
export JAEGER_SAMPLE_RATE=1.0
export TELEMETRY_SERVICE_NAME=auth-service
```

Restart auth-service with these variables.

---

### Step 5: Generate Traffic with Tracing

```bash
# Generate 10 requests
for i in {1..10}; do
  curl -s http://localhost:8085/api/v1/health > /dev/null
  echo "Request $i completed"
done
```

---

### Step 6: View Traces in Jaeger

1. Open http://localhost:16686
2. Select service: `auth-service`
3. Click "Find Traces"
4. You should see traces for each request
5. Click on a trace to see:
   - Span duration
   - HTTP method and status code
   - Database queries
   - Error details (if any)

---

### Step 7: Configure Grafana Dashboard

1. Open http://localhost:3000
2. Login with admin/admin
3. Go to **Configuration** → **Data Sources**
4. Add Prometheus:
   - URL: `http://prometheus:9090`
   - Click "Save & Test"
5. Add Jaeger:
   - URL: `http://jaeger:16686`
   - Click "Save & Test"

---

### Step 8: Create Dashboard in Grafana

1. Click **+** → **Dashboard**
2. Add panels:

**Panel 1: Request Rate**

```promql
rate(http_requests_total[5m])
```

**Panel 2: Error Rate**

```promql
rate(http_requests_total{status=~"5.."}[5m])
```

**Panel 3: Response Time (P95)**

```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**Panel 4: User Registrations**

```promql
increase(business_users_registered_total[1h])
```

3. Save dashboard as "Auth Service Overview"

---

### Step 9: Load Testing

Generate more traffic:

```bash
# Install apache bench if needed
# brew install httpd (macOS)

# Run load test
ab -n 1000 -c 10 http://localhost:8085/api/v1/health

# Results will show:
# - Requests per second
# - Time per request
# - Error rate
```

---

### Step 10: Analyze in Jaeger & Grafana

**In Jaeger:**

1. Search for traces with longest duration
2. Look for errors (traces with error=true)
3. Analyze span hierarchy

**In Grafana:**

1. Check request rate graphs
2. Monitor error rate
3. View P95 latency
4. Track business metrics

---

## Verification Checklist

✅ **Basic Functionality**

- [ ] Auth-service starts successfully
- [ ] Metrics endpoint responds (http://localhost:8085/metrics)
- [ ] Health check passes
- [ ] Can register new user
- [ ] Can login and get JWT token
- [ ] Protected endpoints work with JWT

✅ **Telemetry**

- [ ] Metrics are being collected
- [ ] Business metrics increment correctly
- [ ] HTTP request metrics are accurate
- [ ] Logs include structured data

✅ **Observability (with Docker)**

- [ ] Jaeger UI accessible
- [ ] Traces appear in Jaeger
- [ ] Grafana UI accessible
- [ ] Can create dashboards in Grafana
- [ ] Data sources connected

---

## Troubleshooting

### Issue: Service won't start

**Check:**

```bash
# Is PostgreSQL running?
pg_isready -h localhost -U auth_user

# Are environment variables set?
echo $DATABASE_USERNAME
echo $AUTHORIZATION_MODE

# Check logs
./auth-service 2>&1 | tee service.log
```

### Issue: Metrics endpoint returns 404

**Solution:**

- Verify service is running on port 8085
- Check if `/metrics` endpoint is registered
- Look for metrics middleware initialization in logs

### Issue: No traces in Jaeger

**Check:**

```bash
# Is Jaeger running?
curl http://localhost:14268/api/traces

# Is JAEGER_ENDPOINT set?
echo $JAEGER_ENDPOINT

# Check service logs for telemetry initialization
grep -i "telemetry\|jaeger" service.log
```

### Issue: Tables not created

**Solution:**

```bash
# Run migration
go run cmd/migrate/main.go -action=up

# Or manually check:
psql -U auth_user -d auth_service -c "\dt"
```

---

## Summary

**Without Docker (Minimum):**

1. ✅ Start auth-service
2. ✅ Run migrations
3. ✅ Test endpoints
4. ✅ Check metrics
5. ✅ Run automated tests

**With Docker (Full Stack):**

1. ✅ Start observability stack
2. ✅ Configure telemetry
3. ✅ View traces in Jaeger
4. ✅ Create dashboards in Grafana
5. ✅ Analyze performance

---

## Next Steps

1. **Set up alerts** in Grafana for:

   - Error rate > 1%
   - P95 latency > 500ms
   - Failed login attempts > 10/min

2. **Add custom metrics** for:

   - Active sessions
   - Token generation rate
   - Cache hit rate

3. **Implement distributed tracing** across:

   - Auth-service → Database
   - Auth-service → Redis
   - Auth-service → Kafka

4. **Create SLO dashboards** for:
   - 99.9% availability
   - <200ms P95 latency
   - <0.1% error rate

---

**You're all set! Your auth-service now has full observability! 🎉**

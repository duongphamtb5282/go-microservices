# Quick Start: Jaeger Tracing (5 Minutes)

## Architecture

```
auth-service → Jaeger OTLP Receiver (port 4318) → Jaeger Backend → Jaeger UI (port 16686)
```

**Note**: We use Jaeger's built-in OTLP receiver. No separate collector needed!

## Step 1: Start Jaeger (1 min)

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang
docker-compose -f docker-compose.observability.yml up -d jaeger
```

**Verify Jaeger is running:**
```bash
curl -s http://localhost:16686 | grep -q "Jaeger" && echo "✅ Jaeger UI is up!" || echo "❌ Jaeger not running"
```

## Step 2: Start Auth-Service with Telemetry (1 min)

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Use the startup script (RECOMMENDED)
./start-with-telemetry.sh
```

**Look for this in logs:**
```
"message":"OpenTelemetry initialized successfully"
"otlp_endpoint":"localhost:4318"
```

If you DON'T see this, telemetry is NOT enabled!

## Step 3: Generate Traces (1 min)

```bash
# In a NEW terminal (keep service running)
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Test health endpoint
curl http://localhost:8085/health

# Test JWT login (generates trace with permissions)
JWT_TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"any"}' | jq -r '.token')

echo "Got token: ${JWT_TOKEN:0:50}..."

# Test protected endpoint
curl -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN"

# Generate 10 more traces
for i in {1..10}; do 
  curl -s http://localhost:8085/health > /dev/null
  echo "Sent request $i"
  sleep 1
done
```

## Step 4: View Traces in Jaeger (2 min)

1. **Open Jaeger UI**: http://localhost:16686

2. **Wait 30-60 seconds** for traces to appear

3. **Select service**: Click "Service" dropdown → Select `auth-service`

4. **Click "Find Traces"**

5. **You should see:**
   - `GET /health`
   - `POST /api/v1/auth/login`  
   - `GET /api/v1/users`

6. **Click on any trace** to see:
   - Span hierarchy
   - Duration breakdown
   - HTTP method, status code
   - Request/response details

## Troubleshooting

### Problem: No "auth-service" in dropdown

**Check 1: Is telemetry enabled?**
```bash
grep -i "opentelemetry initialized" auth-service.log
```
Expected: `"message":"OpenTelemetry initialized successfully"`

**Check 2: Are environment variables set?**
```bash
echo $OTEL_ENABLED
echo $OTEL_EXPORTER_OTLP_ENDPOINT
```
Expected: `true` and `localhost:4318`

**Fix:** Restart service WITH environment variables:
```bash
./start-with-telemetry.sh
```

### Problem: "Connection refused" errors

**Check if Jaeger OTLP endpoint is accessible:**
```bash
curl -v http://localhost:4318
```

If connection refused, restart Jaeger:
```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang
docker-compose -f docker-compose.observability.yml restart jaeger
```

### Problem: Port 13133 not found

Port 13133 is for the separate otel-collector health check.
We're using **Jaeger's built-in OTLP receiver** instead, so this port is not needed!

## Diagnostic Tool

Run the full diagnostic:
```bash
./scripts/diagnose-jaeger.sh
```

## Expected Timeline

- **Traces appear in Jaeger**: 30-60 seconds after making requests
- **Traces are batched**: Exported every 5-10 seconds by default
- **Be patient**: First traces may take up to 1 minute

## Success Checklist

- [ ] Jaeger UI accessible at http://localhost:16686
- [ ] Service started with `OTEL_ENABLED=true`
- [ ] Logs show "OpenTelemetry initialized successfully"
- [ ] Made at least 10 HTTP requests
- [ ] Waited 60 seconds
- [ ] Refreshed Jaeger UI
- [ ] "auth-service" appears in Service dropdown
- [ ] Traces visible when clicking "Find Traces"

## Summary

✅ **Simple setup** - No separate collector needed  
✅ **Fast** - Traces appear in ~30 seconds  
✅ **Automatic** - HTTP requests traced automatically  
✅ **Production-ready** - Full distributed tracing

Your auth-service now has complete observability! 🎉

# OpenTelemetry + Jaeger Integration Guide

## Overview

The `auth-service` now uses **OpenTelemetry** to send distributed traces to **Jaeger** via the **OTLP (OpenTelemetry Protocol)** collector.

## Architecture

```
auth-service → OTLP HTTP Exporter → OpenTelemetry Collector → Jaeger Backend
```

## Environment Variables

Set these before starting the service:

```bash
# OpenTelemetry Configuration
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=auth-service
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_ENVIRONMENT=development

# OTLP Exporter (connects to otel-collector)
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318

# Database credentials
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password
```

## Step-by-Step Setup

### Step 1: Start Observability Stack

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang

# Start Jaeger, Prometheus, Grafana, and OTLP Collector
docker-compose -f docker-compose.observability.yml up -d
```

**Services started:**

- Jaeger UI: http://localhost:16686
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- OTLP Collector: http://localhost:4318 (HTTP), localhost:4317 (gRPC)

### Step 2: Verify OTLP Collector

Check that the OTLP collector is running:

```bash
curl http://localhost:13133/
# Expected: {"status":"Server available"}
```

### Step 3: Start Auth Service with Telemetry

**⚠️ CRITICAL: You MUST set environment variables or telemetry will NOT work!**

**Option A: Use the startup script (RECOMMENDED)**

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service
./start-with-telemetry.sh
```

**Option B: Set environment variables manually**

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/auth-service

# Set environment variables (REQUIRED!)
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=auth-service
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# Start the service with logging
./auth-service 2>&1 | tee auth-service.log
```

**Expected logs:**

```
"message":"OpenTelemetry initialized successfully"
"service":"auth-service"
"otlp_endpoint":"localhost:4318"
"enabled":true
```

### Step 4: Generate Traces

Make API calls to generate trace data:

```bash
# 1. Health check
curl http://localhost:8085/health

# 2. Login (generates trace with JWT generation)
JWT_TOKEN=$(curl -s -X POST http://localhost:8085/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "any"
  }' | jq -r '.token')

echo "Got token: ${JWT_TOKEN:0:30}..."

# 3. Protected endpoint (generates trace with auth middleware)
curl -X GET http://localhost:8085/api/v1/users \
  -H "Authorization: Bearer $JWT_TOKEN"

# 4. Multiple requests to generate more traces
for i in {1..5}; do
  curl -s http://localhost:8085/health > /dev/null
  echo "Request $i sent"
  sleep 1
done
```

### Step 5: View Traces in Jaeger

1. Open Jaeger UI: http://localhost:16686

2. In the **Service** dropdown, select `auth-service`

3. Click **Find Traces**

4. You should see traces for:
   - `GET /health`
   - `POST /api/v1/auth/login`
   - `GET /api/v1/users`

### Step 6: Analyze Trace Details

Click on any trace to see:

**Span Details:**

- Operation name (e.g., `GET /health`)
- Duration
- HTTP method, status code, route
- Span ID, Trace ID
- Service name: `auth-service`

**Example Trace Structure:**

```
auth-service: GET /api/v1/users (500ms)
  ├─ HTTP GET /api/v1/users (450ms)
  │   ├─ JWT Validation (50ms)
  │   ├─ Authorization Check (100ms)
  │   └─ Database Query (300ms)
  └─ Middleware Processing (50ms)
```

## Troubleshooting

### Issue 1: No traces in Jaeger

**Check 1: Service is sending traces**

```bash
# Look for OpenTelemetry initialization in logs
./auth-service 2>&1 | grep -i "opentelemetry\|otlp"
```

Expected:

```
"message":"OpenTelemetry initialized successfully"
```

**Check 2: OTLP Collector is running**

```bash
docker ps | grep otel-collector
curl http://localhost:13133/
```

**Check 3: Environment variables are set**

```bash
echo $OTEL_ENABLED
echo $OTEL_EXPORTER_OTLP_ENDPOINT
```

### Issue 2: OTLP endpoint connection refused

**Error:** `failed to create OTLP exporter: connection refused`

**Solution:**

```bash
# Check if otel-collector is running
docker-compose -f docker-compose.observability.yml ps

# Restart if needed
docker-compose -f docker-compose.observability.yml restart otel-collector
```

### Issue 3: Service shows telemetry disabled

**Check logs for:**

```
"message":"Failed to initialize OpenTelemetry"
```

**Solution:** Set environment variables:

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
```

### Issue 4: "auth-service" not in dropdown

Wait 30-60 seconds after making requests, then refresh Jaeger UI.

If still not appearing:

```bash
# Check OTLP collector logs
docker logs otel-collector

# Check Jaeger logs
docker logs jaeger
```

## Configuration Files

### docker-compose.observability.yml

Located at: `/Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/docker-compose.observability.yml`

**Services:**

- `jaeger`: All-in-one Jaeger with OTLP receiver
- `otel-collector`: OpenTelemetry Collector (optional, if using separate collector)
- `prometheus`: Metrics collection
- `grafana`: Visualization

### otel-collector-config.yml

Located at: `/Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/otel-collector-config.yml`

**Key configuration:**

```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger]
```

## Testing Checklist

- [ ] Observability stack is running (docker-compose up)
- [ ] OTLP collector is accessible (curl localhost:13133)
- [ ] Environment variables are set (OTEL_ENABLED=true)
- [ ] Auth service starts successfully
- [ ] Logs show "OpenTelemetry initialized successfully"
- [ ] API requests generate traces
- [ ] Jaeger UI shows "auth-service" in dropdown
- [ ] Traces appear in Jaeger UI
- [ ] Trace details show span information

## Advanced: Custom Spans

To add custom spans in your code:

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

func MyFunction(ctx context.Context) error {
    tracer := otel.Tracer("auth-service")
    ctx, span := tracer.Start(ctx, "MyFunction")
    defer span.End()

    // Add attributes
    span.SetAttributes(
        attribute.String("user.id", userID),
        attribute.String("operation", "create"),
    )

    // Your code here

    return nil
}
```

## Performance Impact

OpenTelemetry adds minimal overhead:

- **Latency:** ~1-5ms per request
- **CPU:** ~1-2% increase
- **Memory:** ~10-20MB additional

## Production Recommendations

1. **Use OTLP over gRPC** (more efficient):

   ```bash
   export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
   ```

2. **Configure sampling** (reduce trace volume):

   - Sample 10% in production
   - Sample 100% in development

3. **Use proper OTLP Collector setup**:

   - Dedicated collector instance
   - Load balancing
   - Buffering and retry logic

4. **Monitor collector health**:
   - Collector metrics endpoint
   - Alert on dropped spans

## Summary

✅ **OpenTelemetry integrated** - Traces sent to Jaeger via OTLP  
✅ **Automatic instrumentation** - HTTP requests traced automatically  
✅ **Distributed tracing** - Full request lifecycle visibility  
✅ **Production-ready** - Configurable sampling and exporters

Your auth-service now has full observability! 🎉

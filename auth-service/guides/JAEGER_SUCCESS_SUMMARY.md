# 🎉 Jaeger Integration Success Summary

## Problem

The `auth-service` was not appearing in Jaeger UI despite:

- ✅ OpenTelemetry configured correctly
- ✅ OTLP endpoint set (`localhost:4318`)
- ✅ Jaeger accepting connections
- ✅ Middleware being added

## Root Cause

**Gin middleware execution order issue**: In Gin, middleware added with `.Use()` only applies to routes registered **AFTER** the middleware is added.

The original code was:

```go
// 1. Routes registered first
ginRouter := router.SetupRoutes()  // All routes defined here

// 2. Middleware added AFTER routes
ginRouter.Use(otelTelemetry.GinMiddleware())  // TOO LATE!
```

Since all routes (including `/health`) were registered in `SetupRoutes()` BEFORE the OpenTelemetry middleware was added, **none of the routes were being traced**.

## Solution

Restructured the initialization order in `service_factory.go`:

```go
// 1. Create Gin engine FIRST
ginRouter := gin.New()

// 2. Add OpenTelemetry middleware BEFORE any routes
if otelTelemetry != nil {
    ginRouter.Use(otelTelemetry.GinMiddleware())
}

// 3. Add standard middleware
ginRouter.Use(gin.Logger())
ginRouter.Use(gin.Recovery())

// 4. NOW register routes (they will be traced)
router := providers.RouterProvider(...)
router.RegisterRoutes(ginRouter)
```

Also refactored `router.go` to separate route registration from engine creation:

- `SetupRoutes()` - Creates engine + registers routes (for backward compatibility)
- `RegisterRoutes(router)` - Only registers routes on provided engine

## Verification

Created a minimal test (`test_trace.go`) that proved:

1. OpenTelemetry initialization works
2. OTLP exporter works
3. Jaeger receives traces

This confirmed the issue was **NOT** with OpenTelemetry configuration, but with **middleware ordering** in the auth-service.

## Files Changed

### 1. `src/applications/service_factory.go`

- Restructured `CreateRouter()` to create Gin engine first
- Added OpenTelemetry middleware before routes
- Added debug `fmt.Printf` statements for troubleshooting

### 2. `src/interfaces/rest/router/router.go`

- Added `RegisterRoutes(router *gin.Engine)` method
- Refactored `SetupRoutes()` to call `RegisterRoutes()`

## Result

```bash
$ curl -s "http://localhost:16686/api/services" | jq -r '.data[]'
jaeger-all-in-one
test-service
auth-service  ← ✅ SUCCESS!
```

Traces are now visible in Jaeger UI at http://localhost:16686

## Key Learnings

1. **Middleware order matters in Gin**: Always add tracing middleware BEFORE routes
2. **Test incrementally**: The minimal test helped isolate the issue to middleware ordering
3. **Debug output is essential**: `fmt.Printf` statements revealed that:
   - OTLP endpoint was configured
   - Telemetry object was created
   - But middleware was added too late

## Quick Start (With Jaeger)

```bash
# 1. Start Jaeger
docker-compose -f docker-compose.observability.yml up -d

# 2. Set environment variables
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# 3. Build and run
go build -o auth-service cmd/http/main.go
./auth-service

# 4. Generate traces
for i in {1..30}; do curl -s http://localhost:8085/health; sleep 1; done

# 5. View in Jaeger (after 60 seconds)
open http://localhost:16686
```

## Environment Variables

| Variable                      | Value            | Purpose                   |
| ----------------------------- | ---------------- | ------------------------- |
| `OTEL_ENABLED`                | `true`           | Enable OpenTelemetry      |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4318` | Jaeger OTLP HTTP endpoint |
| `OTEL_SERVICE_NAME`           | `auth-service`   | Service name in Jaeger    |

## Troubleshooting

### Service still not showing?

1. **Check environment variables are set in the RUNNING process**:

   ```bash
   ps aux | grep auth-service  # Find PID
   cat /proc/PID/environ | tr '\0' '\n' | grep OTEL
   ```

2. **Check startup logs for**:

   ```
   🔍 OTLP CONFIG: endpoint='localhost:4318' enabled=true
   ✅ otelTelemetry object created successfully
   ✅ OpenTelemetry middleware added to Gin
   ```

3. **Verify Jaeger is running**:

   ```bash
   curl http://localhost:4318/v1/traces -H "Content-Type: application/json" -d '{}'
   # Should return HTTP 200
   ```

4. **Wait 60 seconds** after generating traces - export is batched

## References

- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/languages/go/)
- [Jaeger All-in-One](https://www.jaegertracing.io/docs/1.50/getting-started/)
- [otelgin Middleware](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin)

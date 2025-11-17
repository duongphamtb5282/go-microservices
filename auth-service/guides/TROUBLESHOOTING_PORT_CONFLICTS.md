# Troubleshooting Port Conflicts

## Problem: Port Already Allocated

### Error Message

```
Error response from daemon: failed to set up container networking:
driver failed programming external connectivity on endpoint golang-otel-collector-1:
Bind for 0.0.0.0:4317 failed: port is already allocated
```

This means another container (likely Jaeger) is already using the OTLP ports.

---

## Quick Fix (3 Solutions)

### Solution 1: Stop Conflicting Containers (Recommended)

Since you're using Jaeger's built-in OTLP receiver, you don't need a separate otel-collector.

**Step 1: Check running containers**

```bash
docker ps
```

**Step 2: Stop the otel-collector container**

```bash
# Find the container name
docker ps | grep otel-collector

# Stop it
docker stop golang-otel-collector-1

# Or stop all observability services and restart
docker-compose -f docker-compose.observability.yml down
docker-compose -f docker-compose.observability.yml up -d
```

### Solution 2: Use Jaeger's Built-in OTLP (Current Setup)

Your `docker-compose.observability.yml` already has Jaeger with OTLP support. You don't need a separate otel-collector!

**Verify Jaeger configuration:**

```yaml
# docker-compose.observability.yml
jaeger:
  image: jaegertracing/all-in-one:1.50
  ports:
    - "16686:16686" # Jaeger UI
    - "14268:14268" # Jaeger collector
    - "4317:4317" # OTLP gRPC ← Already here!
    - "4318:4318" # OTLP HTTP ← Already here!
  environment:
    - COLLECTOR_OTLP_ENABLED=true
```

**Your service should use:**

```bash
# .env
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
```

### Solution 3: Change Ports

If you really need both, change the otel-collector ports:

**Edit your docker-compose:**

```yaml
otel-collector:
  ports:
    - "14317:4317" # Map to different host port
    - "14318:4318" # Map to different host port
```

**Then update your service:**

```bash
# .env
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:14318
```

---

## Step-by-Step Resolution

### Step 1: Check What's Running

```bash
# List all running containers
docker ps

# Check for Jaeger
docker ps | grep jaeger

# Check for otel-collector
docker ps | grep otel-collector
```

### Step 2: Identify the Conflict

If you see BOTH:

- ✅ `jaeger` container with ports `4317:4317` and `4318:4318`
- ❌ `otel-collector` trying to use the same ports

**Decision**: Keep Jaeger (it has built-in OTLP), remove otel-collector

### Step 3: Clean Up

```bash
# Stop all observability services
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang
docker-compose -f docker-compose.observability.yml down

# Remove any orphan containers
docker ps -a | grep otel-collector
docker rm -f golang-otel-collector-1  # If it exists
```

### Step 4: Restart Clean

```bash
# Start only Jaeger, Prometheus, and Grafana
docker-compose -f docker-compose.observability.yml up -d jaeger prometheus grafana

# Verify ports
docker ps --format "table {{.Names}}\t{{.Ports}}"
```

You should see:

```
jaeger       0.0.0.0:16686->16686/tcp, 0.0.0.0:4317->4317/tcp, 0.0.0.0:4318->4318/tcp
prometheus   0.0.0.0:9090->9090/tcp
grafana      0.0.0.0:3000->3000/tcp
```

### Step 5: Test

```bash
# Test OTLP HTTP endpoint (Jaeger)
curl http://localhost:4318/v1/traces -H "Content-Type: application/json" -d '{}'

# Should return 200 OK

# Start auth-service
cd auth-service
./scripts/run.sh
```

---

## Understanding the Ports

### OTLP Ports

| Port | Protocol | Service       | Your Setup      |
| ---- | -------- | ------------- | --------------- |
| 4317 | gRPC     | OTLP receiver | Jaeger has this |
| 4318 | HTTP     | OTLP receiver | Jaeger has this |

### Why Port Conflict?

You have TWO services trying to use the same ports:

1. **Jaeger** (jaegertracing/all-in-one:1.50)

   - Built-in OTLP receiver
   - Ports: 4317 (gRPC), 4318 (HTTP)

2. **OTEL Collector** (separate container)
   - Standalone OTLP collector
   - Trying to use: 4317, 4318 ← **CONFLICT!**

### Recommended Setup (Current)

**Use Jaeger's built-in OTLP** (no separate collector needed):

```
auth-service → OTLP (4318) → Jaeger → Traces stored
             ↓
          localhost:4318
```

---

## Verification Commands

### Check Port Usage

```bash
# macOS/Linux
lsof -i :4317
lsof -i :4318

# Check which container
docker ps --filter "publish=4317"
docker ps --filter "publish=4318"
```

### Check Container Logs

```bash
# Jaeger logs
docker logs jaeger

# OTEL Collector logs (if running)
docker logs golang-otel-collector-1
```

### Test Connectivity

```bash
# Test OTLP HTTP (should work)
curl -v http://localhost:4318/v1/traces

# Test OTLP gRPC (requires grpcurl)
grpcurl -plaintext localhost:4317 list
```

---

## Common Scenarios

### Scenario 1: "I see both Jaeger and otel-collector"

**Solution**: Stop otel-collector, use Jaeger's built-in OTLP

```bash
docker stop golang-otel-collector-1
docker rm golang-otel-collector-1
```

### Scenario 2: "I only want otel-collector, not Jaeger"

**Solution**: Update docker-compose to remove port conflict

```yaml
# Option A: Don't expose Jaeger's OTLP ports
jaeger:
  ports:
    - "16686:16686" # UI only
    # Remove 4317 and 4318

# Option B: Change otel-collector ports
otel-collector:
  ports:
    - "14317:4317"
    - "14318:4318"
```

### Scenario 3: "I want both (advanced)"

**Solution**: Use different ports and configure forwarding

```yaml
# docker-compose.yml
otel-collector:
  ports:
    - "14317:4317" # External apps use 14317
    - "14318:4318" # External apps use 14318

jaeger:
  ports:
    - "16686:16686"
    # Don't expose 4317/4318 externally

# otel-collector forwards to Jaeger internally
exporters:
  jaeger:
    endpoint: jaeger:14250 # Internal gRPC
```

---

## Prevention

### Best Practice: Use One OTLP Receiver

For most projects, choose ONE:

**Option 1: Jaeger with built-in OTLP (Simplest)** ✅

```yaml
jaeger:
  image: jaegertracing/all-in-one:1.50
  ports:
    - "16686:16686"
    - "4317:4317"
    - "4318:4318"
  environment:
    - COLLECTOR_OTLP_ENABLED=true
```

**Option 2: OTEL Collector → Jaeger (Advanced)**

```yaml
otel-collector:
  ports:
    - "4317:4317"
    - "4318:4318"

jaeger:
  ports:
    - "16686:16686"
    # No OTLP ports exposed
```

### Check Before Starting

```bash
# Before docker-compose up
lsof -i :4317
lsof -i :4318

# If ports are free, proceed
docker-compose -f docker-compose.observability.yml up -d
```

---

## Quick Resolution Script

Save as `scripts/fix-port-conflict.sh`:

```bash
#!/bin/bash

echo "🔧 Fixing OTLP Port Conflict..."
echo ""

# Check what's using the ports
echo "1️⃣  Checking port 4317..."
PORT_4317=$(lsof -ti :4317 2>/dev/null)
if [ -n "$PORT_4317" ]; then
    echo "   Port 4317 in use by PID: $PORT_4317"
else
    echo "   Port 4317 is free"
fi

echo "2️⃣  Checking port 4318..."
PORT_4318=$(lsof -ti :4318 2>/dev/null)
if [ -n "$PORT_4318" ]; then
    echo "   Port 4318 in use by PID: $PORT_4318"
else
    echo "   Port 4318 is free"
fi

echo ""
echo "3️⃣  Checking Docker containers..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "jaeger|otel"

echo ""
echo "4️⃣  Stopping conflicting containers..."
docker stop golang-otel-collector-1 2>/dev/null || echo "   No otel-collector to stop"

echo ""
echo "5️⃣  Restarting observability stack..."
docker-compose -f docker-compose.observability.yml down
docker-compose -f docker-compose.observability.yml up -d jaeger prometheus grafana

echo ""
echo "6️⃣  Waiting for services..."
sleep 5

echo ""
echo "7️⃣  Testing OTLP endpoint..."
curl -s http://localhost:4318/v1/traces -H "Content-Type: application/json" -d '{}' && echo "   ✅ OTLP HTTP working!" || echo "   ❌ OTLP not responding"

echo ""
echo "✅ Done! Services running:"
docker ps --format "table {{.Names}}\t{{.Ports}}" | grep -E "jaeger|prometheus|grafana"

echo ""
echo "🌐 Access:"
echo "   Jaeger:     http://localhost:16686"
echo "   Grafana:    http://localhost:3000"
echo "   Prometheus: http://localhost:9090"
```

Make executable and run:

```bash
chmod +x scripts/fix-port-conflict.sh
./scripts/fix-port-conflict.sh
```

---

## Summary

✅ **Problem**: Port 4317/4318 already in use

✅ **Cause**: Multiple containers trying to use OTLP ports

✅ **Solution**: Use Jaeger's built-in OTLP receiver (simplest)

✅ **Quick Fix**:

```bash
# Stop everything
docker-compose -f docker-compose.observability.yml down

# Start clean (Jaeger has OTLP built-in)
docker-compose -f docker-compose.observability.yml up -d

# Verify
docker ps
curl http://localhost:4318/v1/traces
```

✅ **Your Service**: Use `OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318`

No separate otel-collector needed! 🎉

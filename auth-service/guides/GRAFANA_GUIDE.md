# Grafana Guide - Complete Observability Dashboard

## Table of Contents

1. [Overview](#overview)
2. [What is Grafana?](#what-is-grafana)
3. [Setup and Installation](#setup-and-installation)
4. [First Time Access](#first-time-access)
5. [Connecting Data Sources](#connecting-data-sources)
6. [Creating Dashboards](#creating-dashboards)
7. [Auth-Service Dashboards](#auth-service-dashboards)
8. [Alerting](#alerting)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## Overview

This guide shows you how to use Grafana to visualize metrics, logs, and traces from your `auth-service`.

### What You'll Learn

- ✅ Start Grafana with Docker Compose
- ✅ Connect to Prometheus (metrics)
- ✅ Connect to Jaeger (traces)
- ✅ Create custom dashboards
- ✅ Monitor auth-service performance
- ✅ Set up alerts
- ✅ Best practices for observability

---

## What is Grafana?

**Grafana** is an open-source analytics and monitoring platform that:

- 📊 **Visualizes** metrics, logs, and traces
- 🔗 **Connects** to multiple data sources (Prometheus, Jaeger, Loki, etc.)
- 📈 **Creates** interactive dashboards
- 🚨 **Alerts** when problems occur
- 🎨 **Customizes** visualizations with rich UI

### Why Use Grafana?

| Feature          | Benefit                                    |
| ---------------- | ------------------------------------------ |
| **Unified View** | See metrics, traces, and logs in one place |
| **Real-time**    | Monitor live data as it happens            |
| **Alerting**     | Get notified of issues automatically       |
| **Templating**   | Create reusable dashboards                 |
| **Sharing**      | Share dashboards with your team            |

---

## Setup and Installation

### Option 1: Docker Compose (Recommended)

The `auth-service` already has Grafana configured in `docker-compose.observability.yml`.

#### Start Observability Stack

```bash
# From project root
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang

# Start Grafana + Prometheus + Jaeger
docker-compose -f docker-compose.observability.yml up -d

# Check services are running
docker-compose -f docker-compose.observability.yml ps
```

Expected output:

```
NAME           SERVICE      STATUS    PORTS
grafana        grafana      running   0.0.0.0:3000->3000/tcp
jaeger         jaeger       running   0.0.0.0:16686->16686/tcp, ...
prometheus     prometheus   running   0.0.0.0:9090->9090/tcp
```

#### Verify Grafana is Running

```bash
# Check Grafana health
curl -s http://localhost:3000/api/health | jq

# Should return:
# {
#   "commit": "...",
#   "database": "ok",
#   "version": "..."
# }
```

### Option 2: Standalone Docker

```bash
docker run -d \
  --name=grafana \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana/grafana:latest
```

### Option 3: Local Installation

**macOS**:

```bash
brew install grafana
brew services start grafana
```

**Linux**:

```bash
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
sudo apt-get update
sudo apt-get install grafana
sudo systemctl start grafana-server
```

---

## First Time Access

### 1. Open Grafana

Open your browser and go to:

```
http://localhost:3000
```

### 2. Login

**Default credentials**:

- Username: `admin`
- Password: `admin`

You'll be prompted to change the password on first login.

### 3. Home Screen

After login, you'll see:

- 📊 **Dashboards** - View and create dashboards
- 🔍 **Explore** - Query data sources directly
- 🚨 **Alerting** - Manage alerts
- ⚙️ **Configuration** - Data sources, plugins, settings

---

## Connecting Data Sources

### Add Prometheus (Metrics)

Prometheus collects metrics from your `auth-service`.

#### Step 1: Navigate to Data Sources

1. Click the **⚙️ gear icon** (Configuration) in the left sidebar
2. Click **Data Sources**
3. Click **Add data source**
4. Select **Prometheus**

#### Step 2: Configure Prometheus

**Connection settings**:

```
Name: Prometheus
URL: http://prometheus:9090
```

If running locally (not Docker):

```
URL: http://localhost:9090
```

**Other settings**:

- HTTP Method: `GET`
- Access: `Server (default)`

#### Step 3: Save & Test

Click **Save & Test**

You should see: ✅ **Data source is working**

### Add Jaeger (Traces)

Jaeger provides distributed tracing for your `auth-service`.

#### Step 1: Add Jaeger Data Source

1. Click **⚙️ Configuration** → **Data Sources**
2. Click **Add data source**
3. Select **Jaeger**

#### Step 2: Configure Jaeger

**Connection settings**:

```
Name: Jaeger
URL: http://jaeger:16686
```

If running locally:

```
URL: http://localhost:16686
```

#### Step 3: Save & Test

Click **Save & Test**

You should see: ✅ **Data source is working**

### Optional: Add Loki (Logs)

If you're using Loki for log aggregation:

1. Add data source → Select **Loki**
2. URL: `http://loki:3100`
3. Save & Test

---

## Creating Dashboards

### Create Your First Dashboard

#### Method 1: Quick Start

1. Click **+ (Create)** icon in left sidebar
2. Select **Dashboard**
3. Click **Add new panel**
4. Select data source (Prometheus)
5. Write your query
6. Click **Apply**

#### Method 2: Import Dashboard

1. Click **+ (Create)** → **Import**
2. Enter dashboard ID or upload JSON
3. Select data sources
4. Click **Import**

**Popular dashboard IDs**:

- Go Processes: `6671`
- Prometheus Stats: `2`
- Node Exporter: `1860`

### Dashboard Components

#### Panel Types

| Type          | Use Case         | Example                    |
| ------------- | ---------------- | -------------------------- |
| **Graph**     | Time series data | Request rate over time     |
| **Stat**      | Single value     | Total users                |
| **Gauge**     | Percentage/ratio | CPU usage                  |
| **Table**     | Tabular data     | Error logs                 |
| **Heatmap**   | Distribution     | Response time distribution |
| **Bar Chart** | Comparison       | Requests by endpoint       |

#### Basic Panel Example

**Query**: HTTP request rate

```promql
rate(http_requests_total[5m])
```

**Configuration**:

- Title: "HTTP Requests/sec"
- Unit: "requests/sec"
- Visualization: Time series
- Legend: Show

---

## Auth-Service Dashboards

Let me create pre-configured dashboards for your `auth-service`.

### Dashboard 1: Overview

**Panels to include**:

1. Total Users (Stat)
2. Request Rate (Graph)
3. Response Time (Graph)
4. Error Rate (Graph)
5. Active Sessions (Stat)
6. Database Connections (Gauge)

### Dashboard 2: Performance

**Panels to include**:

1. API Response Times (Heatmap)
2. Request Duration by Endpoint (Graph)
3. Cache Hit Rate (Gauge)
4. Database Query Duration (Graph)

### Dashboard 3: Errors & Issues

**Panels to include**:

1. Error Rate by Endpoint (Bar Chart)
2. Failed Logins (Table)
3. JWT Validation Failures (Graph)
4. Database Errors (Table)

### Dashboard 4: Business Metrics

**Panels to include**:

1. User Registrations (Graph)
2. Login Success Rate (Gauge)
3. Active Users (Stat)
4. Most Used Endpoints (Bar Chart)

---

## Example Queries (PromQL)

### HTTP Metrics

**Request rate (requests per second)**:

```promql
rate(http_requests_total{service="auth-service"}[5m])
```

**Request rate by endpoint**:

```promql
sum by (endpoint) (rate(http_requests_total{service="auth-service"}[5m]))
```

**Request duration (average)**:

```promql
rate(http_request_duration_seconds_sum{service="auth-service"}[5m])
/
rate(http_request_duration_seconds_count{service="auth-service"}[5m])
```

**Request duration (95th percentile)**:

```promql
histogram_quantile(0.95,
  rate(http_request_duration_seconds_bucket{service="auth-service"}[5m])
)
```

**Error rate**:

```promql
sum(rate(http_requests_total{service="auth-service", status=~"5.."}[5m]))
/
sum(rate(http_requests_total{service="auth-service"}[5m]))
* 100
```

### Database Metrics

**Active connections**:

```promql
db_connections_active{service="auth-service"}
```

**Query duration**:

```promql
rate(db_query_duration_seconds_sum[5m])
/
rate(db_query_duration_seconds_count[5m])
```

**Slow queries (>1s)**:

```promql
sum(rate(db_query_duration_seconds_bucket{le="1"}[5m]))
```

### Cache Metrics

**Cache hit rate**:

```promql
sum(rate(cache_hits_total{service="auth-service"}[5m]))
/
(
  sum(rate(cache_hits_total{service="auth-service"}[5m]))
  +
  sum(rate(cache_misses_total{service="auth-service"}[5m]))
)
* 100
```

**Cache operations**:

```promql
sum by (operation) (rate(cache_operations_total{service="auth-service"}[5m]))
```

### Business Metrics

**User registrations**:

```promql
increase(user_registrations_total{service="auth-service"}[1h])
```

**Login success rate**:

```promql
sum(rate(user_login_total{service="auth-service", status="success"}[5m]))
/
sum(rate(user_login_total{service="auth-service"}[5m]))
* 100
```

**Active sessions**:

```promql
user_sessions_active{service="auth-service"}
```

---

## Step-by-Step: Create Auth-Service Dashboard

### Step 1: Create New Dashboard

1. Click **+ (Create)** → **Dashboard**
2. Click **Add new panel**

### Step 2: Add Request Rate Panel

**Query**:

```promql
sum(rate(http_requests_total{service="auth-service"}[5m])) by (method, endpoint)
```

**Panel Settings**:

- Title: "HTTP Requests Rate"
- Description: "Requests per second by endpoint"
- Visualization: Time series
- Unit: req/s
- Legend: Show as table
- Legend values: Last, Min, Max, Avg

**Display**:

- Line interpolation: Smooth
- Fill opacity: 10
- Point size: 5

Click **Apply**

### Step 3: Add Response Time Panel

Click **Add panel**

**Query**:

```promql
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket{service="auth-service"}[5m])) by (le, endpoint)
)
```

**Panel Settings**:

- Title: "Response Time (95th percentile)"
- Unit: seconds (s)
- Visualization: Time series
- Thresholds:
  - Green: 0-0.1s
  - Yellow: 0.1-0.5s
  - Red: >0.5s

Click **Apply**

### Step 4: Add Error Rate Panel

Click **Add panel**

**Query**:

```promql
sum(rate(http_requests_total{service="auth-service", status=~"5.."}[5m]))
/
sum(rate(http_requests_total{service="auth-service"}[5m]))
* 100
```

**Panel Settings**:

- Title: "Error Rate"
- Unit: percent (0-100)
- Visualization: Stat
- Thresholds:
  - Green: 0-1%
  - Yellow: 1-5%
  - Red: >5%

Click **Apply**

### Step 5: Save Dashboard

1. Click **💾 (Save dashboard)** icon at top
2. Name: "Auth-Service Overview"
3. Folder: Create "Auth-Service" folder
4. Click **Save**

---

## Dashboard JSON Templates

### Template 1: Auth-Service Overview

Save this as `auth-service-overview.json`:

```json
{
  "dashboard": {
    "title": "Auth-Service Overview",
    "tags": ["auth-service", "overview"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"auth-service\"}[5m])) by (endpoint)",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time (p95)",
        "type": "graph",
        "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service=\"auth-service\"}[5m])) by (le))",
            "legendFormat": "p95"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "stat",
        "gridPos": { "h": 4, "w": 6, "x": 0, "y": 8 },
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{service=\"auth-service\", status=~\"5..\"}[5m])) / sum(rate(http_requests_total{service=\"auth-service\"}[5m])) * 100"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "thresholds": {
              "steps": [
                { "value": 0, "color": "green" },
                { "value": 1, "color": "yellow" },
                { "value": 5, "color": "red" }
              ]
            }
          }
        }
      }
    ]
  }
}
```

### Import Dashboard

1. Click **+ (Create)** → **Import**
2. Paste JSON or upload file
3. Select Prometheus data source
4. Click **Import**

---

## Alerting

### Create an Alert

#### Step 1: Edit Panel

1. Open dashboard
2. Click on panel title → **Edit**

#### Step 2: Create Alert Rule

1. Go to **Alert** tab
2. Click **Create alert rule from this panel**

#### Step 3: Configure Alert

**Name**: High Error Rate

**Condition**:

```
WHEN avg() OF query(A, 5m, now) IS ABOVE 5
```

**Evaluate every**: `1m`
**For**: `5m` (alert after 5 minutes)

#### Step 4: Notifications

**Contact point**: Create or select existing

**Message**:

```
Auth-Service Error Rate Alert

Error rate: {{ $value }}%
Threshold: 5%

Dashboard: {{ $externalURL }}/d/{{ $dashboardUID }}
```

#### Step 5: Save

Click **Save rule and exit**

### Contact Points

#### Email Notification

1. **Alerting** → **Contact points**
2. Click **New contact point**
3. Name: "Team Email"
4. Integration: **Email**
5. Addresses: `team@example.com`
6. Click **Test** and **Save**

#### Slack Notification

1. Create Slack webhook URL
2. **New contact point**
3. Integration: **Slack**
4. Webhook URL: `https://hooks.slack.com/services/...`
5. Channel: `#alerts`
6. Click **Test** and **Save**

---

## Best Practices

### Dashboard Design

1. **Keep it simple**

   - 6-8 panels per dashboard
   - Most important metrics at top
   - Use consistent colors

2. **Use variables**

   - Service: `$service`
   - Environment: `$environment`
   - Time range: `$__interval`

3. **Organize dashboards**

   - Overview → Details → Troubleshooting
   - Group by service/team
   - Use folders

4. **Performance**
   - Limit time range for heavy queries
   - Use recording rules for complex queries
   - Cache dashboard snapshots

### Naming Conventions

**Dashboards**:

- `Service Overview` - High-level metrics
- `Service Performance` - Detailed performance
- `Service Errors` - Error tracking
- `Service Business` - Business KPIs

**Panels**:

- Clear, descriptive titles
- Include units (req/s, ms, %)
- Use consistent naming

### Query Optimization

1. **Use rate() for counters**

   ```promql
   rate(http_requests_total[5m])
   ```

2. **Use increase() for totals**

   ```promql
   increase(http_requests_total[1h])
   ```

3. **Aggregate early**

   ```promql
   sum by (endpoint) (rate(...))
   ```

4. **Use recording rules**
   ```yaml
   # prometheus.yml
   groups:
     - name: auth_service
       interval: 30s
       rules:
         - record: job:http_requests:rate5m
           expr: sum(rate(http_requests_total{job="auth-service"}[5m]))
   ```

---

## Troubleshooting

### Problem 1: Grafana Not Loading

**Symptoms**: Can't access http://localhost:3000

**Solutions**:

```bash
# Check if Grafana is running
docker ps | grep grafana

# Check logs
docker logs grafana

# Restart Grafana
docker-compose -f docker-compose.observability.yml restart grafana
```

### Problem 2: No Data in Dashboards

**Symptoms**: Panels show "No data"

**Solutions**:

1. **Check data source**:

   - Configuration → Data Sources
   - Click **Test** on Prometheus/Jaeger
   - Should show "Data source is working"

2. **Check Prometheus has data**:

   ```bash
   curl http://localhost:9090/api/v1/query?query=up
   ```

3. **Check time range**:

   - Click time picker (top right)
   - Try "Last 15 minutes"

4. **Check metric names**:
   - Open Prometheus: http://localhost:9090
   - Go to Graph tab
   - Check metric exists

### Problem 3: Slow Dashboards

**Solutions**:

1. **Reduce time range**:

   - Use shorter intervals (5m, 15m)
   - Limit historical data

2. **Optimize queries**:

   - Use recording rules
   - Aggregate data
   - Limit label cardinality

3. **Increase refresh interval**:
   - Dashboard settings
   - Auto-refresh: 30s or 1m

### Problem 4: Alerts Not Firing

**Solutions**:

1. **Check alert rule**:

   - Alerting → Alert rules
   - Check "State" column
   - Review evaluation results

2. **Check contact points**:

   - Alerting → Contact points
   - Click **Test** to verify

3. **Check notification policy**:
   - Alerting → Notification policies
   - Ensure alerts route to correct contact point

---

## Complete Setup Script

Save as `scripts/setup-grafana.sh`:

```bash
#!/bin/bash

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║        Setting up Grafana for Auth-Service                 ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# 1. Start observability stack
echo "1️⃣  Starting Grafana, Prometheus, and Jaeger..."
docker-compose -f docker-compose.observability.yml up -d
sleep 5

# 2. Wait for Grafana to be ready
echo "2️⃣  Waiting for Grafana to start..."
until curl -s http://localhost:3000/api/health > /dev/null; do
    echo "   Waiting for Grafana..."
    sleep 2
done
echo "   ✅ Grafana is ready!"

# 3. Configure Prometheus data source
echo "3️⃣  Configuring Prometheus data source..."
curl -X POST http://admin:admin@localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Prometheus",
    "type": "prometheus",
    "url": "http://prometheus:9090",
    "access": "proxy",
    "isDefault": true
  }' 2>/dev/null
echo "   ✅ Prometheus configured"

# 4. Configure Jaeger data source
echo "4️⃣  Configuring Jaeger data source..."
curl -X POST http://admin:admin@localhost:3000/api/datasources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jaeger",
    "type": "jaeger",
    "url": "http://jaeger:16686",
    "access": "proxy"
  }' 2>/dev/null
echo "   ✅ Jaeger configured"

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                  ✅ Setup Complete! ✅                      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "🌐 Access Grafana:"
echo "   URL: http://localhost:3000"
echo "   Username: admin"
echo "   Password: admin"
echo ""
echo "📊 Next steps:"
echo "   1. Login to Grafana"
echo "   2. Change password"
echo "   3. Import dashboards from guides/GRAFANA_GUIDE.md"
echo "   4. Create custom panels"
echo ""
```

Make executable:

```bash
chmod +x scripts/setup-grafana.sh
```

---

## Quick Reference

### Essential URLs

| Service    | URL                    | Credentials   |
| ---------- | ---------------------- | ------------- |
| Grafana    | http://localhost:3000  | admin / admin |
| Prometheus | http://localhost:9090  | None          |
| Jaeger     | http://localhost:16686 | None          |

### Common PromQL Functions

| Function               | Purpose         | Example                         |
| ---------------------- | --------------- | ------------------------------- |
| `rate()`               | Per-second rate | `rate(requests[5m])`            |
| `increase()`           | Total increase  | `increase(requests[1h])`        |
| `histogram_quantile()` | Percentile      | `histogram_quantile(0.95, ...)` |
| `sum()`                | Aggregate       | `sum(requests) by (endpoint)`   |
| `avg()`                | Average         | `avg(duration)`                 |

### Keyboard Shortcuts

| Shortcut       | Action            |
| -------------- | ----------------- |
| `Ctrl/Cmd + S` | Save dashboard    |
| `Ctrl/Cmd + H` | Toggle row        |
| `Ctrl/Cmd + K` | Toggle kiosk mode |
| `d + n`        | New dashboard     |
| `d + f`        | Open folder       |
| `e`            | Edit panel        |
| `v`            | View panel        |

---

## Summary

✅ **You learned**:

- How to start Grafana with Docker Compose
- How to connect Prometheus and Jaeger
- How to create custom dashboards
- How to write PromQL queries
- How to set up alerts
- Best practices for observability

✅ **Quick Start**:

```bash
# 1. Start Grafana
docker-compose -f docker-compose.observability.yml up -d

# 2. Open browser
open http://localhost:3000

# 3. Login (admin/admin)

# 4. Add data sources (Prometheus, Jaeger)

# 5. Create dashboards
```

✅ **Resources**:

- Grafana Docs: https://grafana.com/docs/
- PromQL Guide: https://prometheus.io/docs/prometheus/latest/querying/basics/
- Dashboard Examples: https://grafana.com/grafana/dashboards/

🎉 **You're ready to monitor your auth-service with Grafana!**

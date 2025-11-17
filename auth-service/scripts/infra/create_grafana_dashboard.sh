#!/bin/bash

# Create Auth Service Overview Dashboard
DASHBOARD_JSON='{
  "dashboard": {
    "title": "Auth Service Overview",
    "tags": ["auth-service", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "HTTP Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Go Routines",
        "type": "graph",
        "targets": [
          {
            "expr": "go_goroutines",
            "legendFormat": "Goroutines",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_resident_memory_bytes / 1024 / 1024",
            "legendFormat": "Memory (MB)",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 4,
        "title": "GC Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "go_gc_duration_seconds",
            "legendFormat": "{{quantile}} quantile",
            "refId": "A"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      }
    ],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s"
  }
}'

curl -X POST -H "Content-Type: application/json" -u admin:admin \
  http://localhost:3000/api/dashboards/db \
  -d "$DASHBOARD_JSON"

echo "Dashboard created!"

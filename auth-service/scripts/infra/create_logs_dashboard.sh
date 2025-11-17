#!/bin/bash

# Create Logs Dashboard for Loki
LOGS_DASHBOARD_JSON='{
  "dashboard": {
    "title": "Auth Service Logs",
    "tags": ["logs", "auth-service", "loki"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "All Application Logs",
        "type": "logs",
        "targets": [
          {
            "expr": "{service=\"auth-service\"} |= \"\"",
            "refId": "A",
            "datasource": "Loki"
          }
        ],
        "gridPos": {"h": 12, "w": 24, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Error Logs",
        "type": "logs",
        "targets": [
          {
            "expr": "{service=\"auth-service\", level=\"error\"} |= \"\"",
            "refId": "A",
            "datasource": "Loki"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 12}
      },
      {
        "id": 3,
        "title": "Warning Logs",
        "type": "logs",
        "targets": [
          {
            "expr": "{service=\"auth-service\", level=\"warn\"} |= \"\"",
            "refId": "A",
            "datasource": "Loki"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 12}
      },
      {
        "id": 4,
        "title": "Log Volume by Level",
        "type": "barchart",
        "targets": [
          {
            "expr": "count_over_time({service=\"auth-service\"} [5m]) by (level)",
            "refId": "A",
            "datasource": "Loki"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 20}
      },
      {
        "id": 5,
        "title": "HTTP Request Logs",
        "type": "logs",
        "targets": [
          {
            "expr": "{service=\"auth-service\"} |= \"GET\" or \"POST\" or \"PUT\" or \"DELETE\"",
            "refId": "A",
            "datasource": "Loki"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 20}
      }
    ],
    "time": {"from": "now-1h", "to": "now"},
    "refresh": "30s",
    "timepicker": {
      "refresh_intervals": ["5s", "10s", "30s", "1m", "5m", "15m", "30m", "1h", "2h", "1d"]
    }
  }
}'

curl -X POST -H "Content-Type: application/json" -u admin:admin \
  http://localhost:3000/api/dashboards/db \
  -d "$LOGS_DASHBOARD_JSON"

echo "Logs dashboard created!"

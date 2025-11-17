#!/bin/bash

# Add Loki as a data source in Grafana
LOKI_DATASOURCE_JSON='{
  "name": "Loki",
  "type": "loki",
  "url": "http://loki:3100",
  "access": "proxy",
  "isDefault": false,
  "jsonData": {
    "maxLines": 1000,
    "timeout": "60s"
  }
}'

curl -X POST -H "Content-Type: application/json" -u admin:admin \
  http://localhost:3000/api/datasources \
  -d "$LOKI_DATASOURCE_JSON"

echo "Loki data source added to Grafana!"

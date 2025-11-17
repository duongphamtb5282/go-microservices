#!/bin/bash

echo "=== Starting Infrastructure Exporters ==="

# Clean up any existing containers
docker rm -f node-exporter-test postgres-exporter-test kafka-exporter-test cadvisor-test 2>/dev/null || true

echo "1. Starting Node Exporter..."
docker run -d \
  --name node-exporter-test \
  --network host \
  -v /proc:/host/proc:ro \
  -v /sys:/host/sys:ro \
  -v /:/rootfs:ro \
  prom/node-exporter \
  --path.procfs=/host/proc \
  --path.rootfs=/rootfs \
  --path.sysfs=/host/sys \
  --collector.filesystem.mount-points-exclude="^/(sys|proc|dev|host|etc)" \
  --collector.filesystem.ignored-mount-points="^/(run|var|snap)"

echo "2. Starting PostgreSQL Exporter..."
docker run -d \
  --name postgres-exporter-test \
  --network host \
  -e DATA_SOURCE_NAME="postgresql://auth_user:auth_password@localhost:5432/auth_service?sslmode=disable" \
  prometheuscommunity/postgres-exporter

echo "3. Starting Kafka Exporter..."
docker run -d \
  --name kafka-exporter-test \
  --network host \
  -e KAFKA_SERVER=localhost:9092 \
  danielqsj/kafka-exporter \
  --kafka.server=localhost:9092

echo "4. Starting cAdvisor..."
docker run -d \
  --name cadvisor-test \
  --network host \
  -v /:/rootfs:ro \
  -v /var/run:/var/run:ro \
  -v /sys:/sys:ro \
  -v /var/lib/docker:/var/lib/docker:ro \
  -v /dev/disk:/dev/disk:ro \
  --privileged \
  gcr.io/cadvisor/cadvisor

echo ""
echo "=== Checking Status ==="
sleep 3
docker ps | grep -E "(node-exporter|postgres-exporter|kafka-exporter|cadvisor)" || echo "Some exporters may not have started"

echo ""
echo "=== Testing Metrics Endpoints ==="
echo "Node Exporter: http://localhost:9100/metrics"
curl -s http://localhost:9100/metrics | head -3 || echo "Node exporter not responding"

echo ""
echo "PostgreSQL Exporter: http://localhost:9187/metrics"
curl -s http://localhost:9187/metrics | head -3 || echo "PostgreSQL exporter not responding"

echo ""
echo "Kafka Exporter: http://localhost:9308/metrics"
curl -s http://localhost:9308/metrics | head -3 || echo "Kafka exporter not responding"

echo ""
echo "cAdvisor: http://localhost:8080/metrics"
curl -s http://localhost:8080/metrics | head -3 || echo "cAdvisor not responding"

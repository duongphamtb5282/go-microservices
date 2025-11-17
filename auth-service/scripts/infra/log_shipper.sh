#!/bin/bash
LOG_FILE="/Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/go/golang_keycloak/infrastructure/logs/auth-service.log"
LOKI_URL="http://localhost:3100/loki/api/v1/push"

# Function to send log to Loki
send_to_loki() {
    local log_line="$1"
    local timestamp=$(date +%s000000000)
    
    # Escape quotes in log line
    log_line=$(echo "$log_line" | sed 's/"/\\"/g')
    
    # Format for Loki
    local payload="{\"streams\": [{\"stream\": {\"job\": \"auth-service-local\", \"service\": \"auth-service\", \"component\": \"application\"}, \"values\": [[\"$timestamp\", \"$log_line\"]]}]}"
    
    echo "Sending: $payload" >&2
    curl -s -X POST -H "Content-Type: application/json" -d "$payload" "$LOKI_URL" >&2
}

# Tail the log file and send new lines to Loki
tail -f "$LOG_FILE" | while read -r line; do
    if [ -n "$line" ]; then
        send_to_loki "$line"
    fi
done

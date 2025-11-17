#!/bin/bash

# Load Testing Script for Auth Service
# Usage: ./load-test.sh [test-type] [duration]

set -euo pipefail

# Configuration
SERVICE_URL="${SERVICE_URL:-https://auth.yourcompany.com}"
TEST_TYPE="${1:-full}"
DURATION="${2:-5m}"
K6_VERSION="${K6_VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check if k6 is installed
    if ! command -v k6 &> /dev/null; then
        log_info "k6 not found. Installing k6..."
        install_k6
    fi

    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi

    log_success "Prerequisites check passed"
}

# Install k6
install_k6() {
    case "$(uname -s)" in
        Linux*)
            if command -v apt-get &> /dev/null; then
                sudo apt-get update
                sudo apt-get install -y k6
            elif command -v yum &> /dev/null; then
                sudo yum install -y k6
            else
                log_error "Could not install k6. Please install manually."
                exit 1
            fi
            ;;
        Darwin*)
            if command -v brew &> /dev/null; then
                brew install k6
            else
                log_error "Homebrew not found. Please install k6 manually."
                exit 1
            fi
            ;;
        *)
            log_error "Unsupported OS. Please install k6 manually."
            exit 1
            ;;
    esac
}

# Basic health check
health_check() {
    log_info "Running basic health check..."

    if curl -f --max-time 10 "${SERVICE_URL}/health" &> /dev/null; then
        log_success "Health check passed"
        return 0
    else
        log_error "Health check failed"
        return 1
    fi
}

# Smoke test - basic functionality test
smoke_test() {
    log_info "Running smoke test..."

    # Test health endpoint
    if ! health_check; then
        log_error "Smoke test failed - service not healthy"
        exit 1
    fi

    # Test metrics endpoint (should be accessible)
    if curl -f --max-time 10 "${SERVICE_URL}/metrics" &> /dev/null; then
        log_success "Metrics endpoint accessible"
    else
        log_warn "Metrics endpoint not accessible"
    fi

    # Test basic API endpoints
    log_info "Testing basic API endpoints..."

    # This would typically test actual API endpoints
    # For now, just verify the service responds
    if curl -f --max-time 10 "${SERVICE_URL}/api/v1/health" &> /dev/null; then
        log_success "API health check passed"
    else
        log_warn "API health check failed (this may be expected if endpoints don't exist)"
    fi

    log_success "Smoke test completed"
}

# Load test with k6
load_test() {
    local duration="$1"
    local test_file=""

    case "$TEST_TYPE" in
        "auth")
            test_file="k6/auth-load-test.js"
            ;;
        "api")
            test_file="k6/api-load-test.js"
            ;;
        "spike")
            test_file="k6/spike-test.js"
            ;;
        "stress")
            test_file="k6/stress-test.js"
            ;;
        "full")
            test_file="k6/full-load-test.js"
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            echo "Available test types: auth, api, spike, stress, full"
            exit 1
            ;;
    esac

    if [ ! -f "$test_file" ]; then
        log_warn "Test file $test_file not found. Creating default test..."
        create_default_test "$test_file"
    fi

    log_info "Running $TEST_TYPE load test for $duration..."
    log_info "Test file: $test_file"
    log_info "Service URL: $SERVICE_URL"

    # Run k6 test
    k6 run \
        --duration="$duration" \
        --out json=results.json \
        --out csv=results.csv \
        --out html=report.html \
        "$test_file"

    log_success "Load test completed"
    log_info "Results saved to results.json, results.csv, and report.html"
}

# Create default test file if it doesn't exist
create_default_test() {
    local test_file="$1"

    mkdir -p "$(dirname "$test_file")"

    cat > "$test_file" << EOF
import http from 'k6/http';
import { check, sleep } from 'k6';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '${DURATION}', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '2m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% requests < 500ms
    http_req_failed: ['rate<0.1'],    // Error rate < 10%
    http_reqs: ['rate>100'],          // At least 100 requests per second
  },
};

export default function () {
  // Health check
  const healthResponse = http.get('${SERVICE_URL}/health');
  check(healthResponse, {
    'health status is 200': (r) => r.status === 200,
  });

  // API test (adjust based on your actual endpoints)
  const apiResponse = http.get('${SERVICE_URL}/api/v1/health');
  check(apiResponse, {
    'api status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}

export function handleSummary(data) {
  return {
    'results.json': JSON.stringify(data, null, 2),
    'report.html': htmlReport(data),
  };
}
EOF

    log_info "Created default test file: $test_file"
}

# Performance analysis
analyze_results() {
    log_info "Analyzing test results..."

    if [ -f "results.json" ]; then
        # Extract key metrics from JSON results
        local total_requests=$(jq '.metrics.http_reqs.values.count' results.json)
        local failed_requests=$(jq '.metrics.http_req_failed.values.rate' results.json)
        local avg_response_time=$(jq '.metrics.http_req_duration.values.avg' results.json)
        local p95_response_time=$(jq '.metrics.http_req_duration.values["95%"]' results.json)
        local p99_response_time=$(jq '.metrics.http_req_duration.values["99%"]' results.json)

        echo ""
        echo "=== Performance Analysis ==="
        echo "Total Requests: $total_requests"
        echo "Failed Requests: $(echo "$failed_requests * 100" | bc -l | cut -d'.' -f1)%"
        echo "Average Response Time: ${avg_response_time}ms"
        echo "95th Percentile: ${p95_response_time}ms"
        echo "99th Percentile: ${p99_response_time}ms"

        # Performance assessment
        echo ""
        echo "=== Performance Assessment ==="

        if (( $(echo "$avg_response_time < 100" | bc -l) )); then
            log_success "Excellent performance - average response time under 100ms"
        elif (( $(echo "$avg_response_time < 500" | bc -l) )); then
            log_success "Good performance - average response time under 500ms"
        elif (( $(echo "$avg_response_time < 1000" | bc -l) )); then
            log_warn "Acceptable performance - average response time under 1s"
        else
            log_error "Poor performance - average response time over 1s"
        fi

        if (( $(echo "$failed_requests < 0.01" | bc -l) )); then
            log_success "Excellent reliability - error rate under 1%"
        elif (( $(echo "$failed_requests < 0.05" | bc -l) )); then
            log_success "Good reliability - error rate under 5%"
        elif (( $(echo "$failed_requests < 0.1" | bc -l) )); then
            log_warn "Acceptable reliability - error rate under 10%"
        else
            log_error "Poor reliability - error rate over 10%"
        fi

        if (( $(echo "$p95_response_time < 500" | bc -l) )); then
            log_success "Excellent 95th percentile - under 500ms"
        elif (( $(echo "$p95_response_time < 1000" | bc -l) )); then
            log_success "Good 95th percentile - under 1s"
        else
            log_warn "95th percentile over 1s - consider optimization"
        fi
    else
        log_warn "No results file found for analysis"
    fi
}

# Generate performance report
generate_report() {
    log_info "Generating performance report..."

    local report_file="performance-report-$(date +%Y%m%d-%H%M%S).md"

    cat > "$report_file" << EOF
# Auth Service Performance Report

**Test Date:** $(date)
**Test Type:** $TEST_TYPE
**Duration:** $DURATION
**Service URL:** $SERVICE_URL

## Test Results

EOF

    if [ -f "results.json" ]; then
        # Add results to report
        echo "- **Total Requests:** $(jq '.metrics.http_reqs.values.count' results.json)" >> "$report_file"
        echo "- **Error Rate:** $(jq '.metrics.http_req_failed.values.rate * 100' results.json | cut -d'.' -f1)%" >> "$report_file"
        echo "- **Average Response Time:** $(jq '.metrics.http_req_duration.values.avg' results.json)ms" >> "$report_file"
        echo "- **95th Percentile:** $(jq '.metrics.http_req_duration.values["95%"]' results.json)ms" >> "$report_file"
        echo "- **99th Percentile:** $(jq '.metrics.http_req_duration.values["99%"]' results.json)ms" >> "$report_file"
    fi

    cat >> "$report_file" << EOF

## Recommendations

- Monitor system resources during peak load
- Consider implementing caching for frequently accessed data
- Review database queries for optimization opportunities
- Implement circuit breakers for external service calls

## Next Steps

1. Review the detailed HTML report for visual analysis
2. Check system metrics during the test period
3. Identify any bottlenecks or performance issues
4. Plan capacity based on expected traffic patterns

EOF

    log_success "Performance report generated: $report_file"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test artifacts..."

    # Remove temporary files
    rm -f results.json results.csv report.html

    log_success "Cleanup completed"
}

# Main execution
main() {
    log_info "Starting load testing for Auth Service"
    log_info "Test type: $TEST_TYPE"
    log_info "Duration: $DURATION"
    log_info "Service URL: $SERVICE_URL"

    case "$TEST_TYPE" in
        "health")
            check_prerequisites
            health_check
            ;;
        "smoke")
            check_prerequisites
            smoke_test
            ;;
        "load"|"auth"|"api"|"spike"|"stress"|"full")
            check_prerequisites
            smoke_test
            load_test "$DURATION"
            analyze_results
            generate_report
            cleanup
            ;;
        "analyze")
            analyze_results
            ;;
        "report")
            generate_report
            ;;
        *)
            log_error "Unknown test type: $TEST_TYPE"
            echo "Available test types: health, smoke, load, auth, api, spike, stress, full, analyze, report"
            echo "Usage: $0 [test-type] [duration]"
            echo "Example: $0 full 5m"
            exit 1
            ;;
    esac

    log_success "Load testing completed!"
}

# Handle script interruption
trap 'log_error "Load testing interrupted"; cleanup; exit 1' INT TERM

# Run main function
main "$@"

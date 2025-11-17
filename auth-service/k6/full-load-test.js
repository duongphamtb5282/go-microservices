import http from 'k6/http';
import { check, sleep } from 'k6';
import { htmlReport } from 'https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js';

// Test configuration
export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
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
  const healthResponse = http.get('https://auth.yourcompany.com/health');
  check(healthResponse, {
    'health status is 200': (r) => r.status === 200,
  });

  // API test (adjust based on your actual endpoints)
  const apiResponse = http.get('https://auth.yourcompany.com/api/v1/health');
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

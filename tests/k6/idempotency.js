
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

// Override this when you want to repeat a test with a fresh key.
const RUN_ID = __ENV.RUN_ID || `run-${Date.now()}`;

export const options = {
  scenarios: {
    concurrent_duplicates: {
      executor: 'shared-iterations',
      vus: 20,
      iterations: 20,
      maxDuration: '30s',
    },
  },
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const idempotencyKey = `k6-concurrent-${RUN_ID}-001`;

  const event = {
    type: 'order.created',
    data: {
      order_id: `concurrent-${RUN_ID}`,
      source: 'k6-idempotency-test',
    },
  };

  const response = http.post(
    `${BASE_URL}/v1/events`,
    JSON.stringify(event),
    {
      headers: {
        'Content-Type': 'application/json',
        'Idempotency-Key': idempotencyKey,
      },
      tags: { endpoint: 'concurrent_idempotency' },
    }
  );

  let body = {};
  try {
    body = response.json();
  } catch (_) {
    // The checks below will fail if the response is not valid JSON.
  }

  check(response, {
    'returns 202 or 200': (r) =>
      r.status === 202 || r.status === 200,
    'returns JSON': (r) =>
      r.headers['Content-Type']?.includes('application/json'),
    'contains event ID': () =>
      typeof body.id === 'string' && body.id.length > 0,
    'contains event status': () =>
      typeof body.status === 'string' && body.status.length > 0,
  });
}


import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  scenarios: {
    event_ingestion: {
      executor: 'constant-vus',
      vus: 1,
      duration: '20s',
    },
  },

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
    checks: ['rate>0.99'],
  },
};

export default function () {
  const event = {
    type: 'order.created',
    data: {
      order_id: `order-${__VU}-${__ITER}`,
      source: 'k6-load-test',
    },
  };

  const response = http.post(
    `${BASE_URL}/v1/events`,
    JSON.stringify(event),
    {
      headers: {
        'Content-Type': 'application/json',
        'Idempotency-Key': `k6-${__VU}-${__ITER}`,
      },
      tags: { endpoint: 'ingest_event' },
    }
  );

  if (response.status !== 202) {
  console.error(
    `status=${response.status} body=${response.body}`
  );
 }

  let body;

  try {
    body = response.json();
  } catch {
    body = {};
  }

  check(response, {
    'returns 202 Accepted': (res) => res.status === 202,
    'returns JSON': (res) =>
      res.headers['Content-Type']?.includes('application/json'),
    'contains event ID': () =>
      typeof body.id === 'string' && body.id.length > 0,
    'contains event status': () =>
      typeof body.status === 'string',
  });

  sleep(0.1);
}

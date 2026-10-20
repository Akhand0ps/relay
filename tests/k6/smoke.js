
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:3000';

export const options = {
  scenarios: {
    load_test: {
      executor: 'shared-iterations',
      vus: 10,
      iterations: 1000,
      maxDuration: '2m',
    },
  },

  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
    checks: ['rate>0.99'],
  },
};

export default function () {
  const response = http.get(`${BASE_URL}/health`);

  check(response, {
    'status is 200': (res) => res.status === 200,
    'content type is JSON': (res) =>
      res.headers['Content-Type']?.includes('application/json'),
    'health status is ok': (res) => {
      try {
        return JSON.parse(res.body).status === 'ok';
      } catch {
        return false;
      }
    },
  });

  sleep(0.1);
}

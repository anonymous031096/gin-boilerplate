import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api';
const DEVICE_ID = 'k6-load-test';

export const options = {
  stages: [
    { duration: '10s', target: 2000 },   // ramp up
    { duration: '20s', target: 10000 },  // ramp up to 10k VUs
    { duration: '30s', target: 10000 },  // hold 10k VUs
    { duration: '10s', target: 0 },      // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<2000', 'p(95)<1000', 'p(50)<500'],  // enterprise SLA
    http_req_failed: ['rate<0.001'],                                // < 0.1% errors
  },
};

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({
      email: 'admin@example.com',
      password: 'Abc@1234',
    }),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Device-Id': DEVICE_ID,
      },
    }
  );

  check(loginRes, {
    'login status 200': (r) => r.status === 200,
  });

  const body = JSON.parse(loginRes.body);
  return { token: body.accessToken };
}

export default function (data) {
  const res = http.get(`${BASE_URL}/auth/me`, {
    headers: {
      Authorization: `Bearer ${data.token}`,
      'X-Device-Id': DEVICE_ID,
    },
  });

  check(res, {
    'status 200': (r) => r.status === 200,
    'has id': (r) => JSON.parse(r.body).id !== undefined,
  });
}

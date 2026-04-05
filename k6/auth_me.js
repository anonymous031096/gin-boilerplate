import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080/api';
const DEVICE_ID = 'k6-load-test';

export const options = {
  stages: [
    { duration: '10s', target: 500 },    // warm up
    { duration: '30s', target: 1000 },   // hold 1k VUs
    { duration: '10s', target: 0 },      // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(99)<5000', 'p(95)<3000', 'p(50)<1000'],  // enterprise SLA
    http_req_failed: ['rate<0.01'],                                  // < 1% errors
  },
};

export function setup() {
  const loginRes = http.post(
    `${BASE_URL}/auth/login`,
    JSON.stringify({
      email: 'admin@init.com',
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
  const res = http.get(`${BASE_URL}/users/me`, {
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

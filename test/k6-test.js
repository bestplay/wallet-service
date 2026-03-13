import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = 'http://localhost:8080';

const errorRate = new Rate('errors');

export function setup() {
  const wallets = [];

  for (let i = 0; i < 10; i++) {
    const res = http.post(`${BASE_URL}/wallets`);
    if (res.status === 201) {
      const walletId = res.json().id;
      wallets.push(walletId);

      // 给每个钱包充值 10000
      const rechargePayload = JSON.stringify({
        walletId: walletId,
        amount: 10000,
      });
      http.post(`${BASE_URL}/wallets/recharge`, rechargePayload, {
        headers: { 'Content-Type': 'application/json' },
      });
    }
  }

  return { wallets };
}

export default function (data) {
  const wallets = data.wallets;
  if (wallets.length < 2) {
    return;
  }

  const scenario = Math.random();

  if (scenario < 0.3) {
    testCreateWallet();
  } else if (scenario < 0.8) {
    testGetWallet(wallets);
  } else {
    testTransfer(wallets);
  }

  sleep(1);
}

function testCreateWallet() {
  const res = http.post(`${BASE_URL}/wallets`);
  const success = check(res, {
    'create wallet status is 201': (r) => r.status === 201,
    'create wallet response has id': (r) => r.json().id !== undefined,
  });
  errorRate.add(!success);
}

function testGetWallet(wallets) {
  const walletId = wallets[Math.floor(Math.random() * wallets.length)];
  const res = http.get(`${BASE_URL}/wallets/${walletId}`);
  const success = check(res, {
    'get wallet status is 200': (r) => r.status === 200,
    'get wallet response has id': (r) => r.json().id === walletId,
  });
  errorRate.add(!success);
}

function testTransfer(wallets) {
  const sourceIdx = Math.floor(Math.random() * wallets.length);
  let destIdx = Math.floor(Math.random() * wallets.length);
  while (destIdx === sourceIdx) {
    destIdx = Math.floor(Math.random() * wallets.length);
  }

  const sourceId = wallets[sourceIdx];
  const destId = wallets[destIdx];
  const amount = (Math.random() * 100).toFixed(2);

  const payload = JSON.stringify({
    sourceId: sourceId,
    destId: destId,
    amount: parseFloat(amount),
  });

  const res = http.post(`${BASE_URL}/wallets/transfer`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  const success = check(res, {
    'transfer status is 200 or 400': (r) => r.status === 200 || r.status === 400,
  });
  errorRate.add(!success && res.status !== 400);
}

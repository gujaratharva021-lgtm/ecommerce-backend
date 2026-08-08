import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL = __ENV.BASE_URL || 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

// Coordinates near "Default Warehouse" (Mumbai, 19.1197, 72.8468, 10km radius)
// so checkout's serviceability check passes.
const TEST_LAT = 19.1197;
const TEST_LNG = 72.8468;

// How many distinct logged-in "sessions" to prepare before the load stages
// begin. /auth/send-otp is rate-limited to 5 requests/minute PER IP, and
// every k6 VU on your machine shares one IP -- so logins happen once, up
// front, in small batches that respect that limit, not once per iteration.
// This also matches how a real app behaves: users log in once and reuse
// that session for many requests, they don't re-login on every action.
const SESSION_POOL_SIZE = Number(__ENV.SESSION_POOL_SIZE || 8);

const checkoutFailures = new Counter('checkout_failures');
const journeyDuration = new Trend('full_journey_duration', true);

export const options = {
  setupTimeout: '5m', // logging in a pool of users respects a 5/min rate limit, so this takes a bit
  stages: [
    { duration: '30s', target: 5 },
    { duration: '1m', target: 5 },
    { duration: '30s', target: 10 },
    { duration: '1m', target: 10 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],
    http_req_failed: ['rate<0.05'],
    checkout_failures: ['count<5'],
  },
};

function jsonHeaders(token) {
  const h = { 'Content-Type': 'application/json' };
  if (token) h.Authorization = `Bearer ${token}`;
  return { headers: h };
}

function uniquePhone(n) {
  return `9${String(n).padStart(9, '0')}`.slice(0, 10);
}

// ---------------------------------------------------------------------------
// setup() runs once, before any VU starts iterating, and does NOT count
// toward the load stages/metrics above. This is where we build our pool of
// logged-in sessions (with a saved address) that the VUs will reuse.
// ---------------------------------------------------------------------------
export function setup() {
  const sessions = [];

  for (let i = 0; i < SESSION_POOL_SIZE; i++) {
    // Respect the 5-requests-per-minute-per-IP limit on /auth/send-otp:
    // after every 5 logins, wait out the rest of the window.
    if (i > 0 && i % 5 === 0) {
      console.log('Pausing ~65s to respect the send-otp rate limit...');
      sleep(65);
    }

    const phone = uniquePhone(i);

    const sendRes = http.post(
      `${BASE_URL}/auth/send-otp`,
      JSON.stringify({ phone }),
      jsonHeaders()
    );
    if (sendRes.status !== 200) {
      console.log(`setup: send-otp failed for session ${i}: ${sendRes.status} ${sendRes.body}`);
      continue;
    }
    const otp = JSON.parse(sendRes.body).otp;

    const verifyRes = http.post(
      `${BASE_URL}/auth/verify-otp`,
      JSON.stringify({ phone, otp, name: `LoadTest Session ${i}` }),
      jsonHeaders()
    );
    if (verifyRes.status !== 200) {
      console.log(`setup: verify-otp failed for session ${i}: ${verifyRes.status} ${verifyRes.body}`);
      continue;
    }
    const token = JSON.parse(verifyRes.body).token;

    // Save one address per session, with valid coordinates, so checkout
    // never needs to create one mid-iteration.
    const addrRes = http.post(
      `${BASE_URL}/addresses`,
      JSON.stringify({
        label: 'Home',
        full_name: `LoadTest User ${i}`,
        phone,
        line1: '123 Test Street',
        line2: '',
        city: 'Mumbai',
        state: 'Maharashtra',
        pincode: '400001',
        lat: TEST_LAT,
        lng: TEST_LNG,
        is_default: true,
      }),
      jsonHeaders(token)
    );
    let addressId;
    if (addrRes.status >= 200 && addrRes.status < 300) {
      try {
        const body = JSON.parse(addrRes.body);
        addressId = body.id || (body.address && body.address.id);
      } catch (e) {
        // ignore
      }
    }
    if (!addressId) {
      console.log(`setup: address creation failed for session ${i}: ${addrRes.status} ${addrRes.body}`);
      continue;
    }

    sessions.push({ token, addressId });
  }

  if (sessions.length === 0) {
    throw new Error('setup() could not create any logged-in sessions - aborting load test.');
  }

  console.log(`setup: ${sessions.length}/${SESSION_POOL_SIZE} sessions ready.`);
  return { sessions };
}

// ---------------------------------------------------------------------------
// Main VU flow - reuses one of the pre-authenticated sessions from setup().
// ---------------------------------------------------------------------------
export default function (data) {
  const session = data.sessions[__VU % data.sessions.length];
  const { token, addressId } = session;
  const start = Date.now();

  let productId;

  group('1. Browse products', function () {
    const listRes = http.get(`${BASE_URL}/products?page=1&limit=20`, jsonHeaders(token));
    check(listRes, {
      'products status 200': (r) => r.status === 200,
      'products list non-empty': (r) => {
        try {
          return (JSON.parse(r.body).products || []).length > 0;
        } catch (e) {
          return false;
        }
      },
    });
    try {
      const products = JSON.parse(listRes.body).products || [];
      if (products.length > 0) {
        // Pick a random product each iteration so load (and stock use)
        // spreads across the catalog instead of hammering a single item.
        productId = products[Math.floor(Math.random() * products.length)].id;
      }
    } catch (e) {
      // leave productId undefined
    }

    const searchRes = http.get(`${BASE_URL}/products?search=milk&limit=10`, jsonHeaders(token));
    check(searchRes, { 'search status 200': (r) => r.status === 200 });
  });

  sleep(1);

  if (!productId) {
    sleep(1);
    return;
  }

  group('2. Add to cart', function () {
    const addRes = http.post(
      `${BASE_URL}/cart`,
      JSON.stringify({ product_id: productId, quantity: 1 }),
      jsonHeaders(token)
    );
    check(addRes, { 'add to cart status 2xx': (r) => r.status >= 200 && r.status < 300 });

    const cartRes = http.get(`${BASE_URL}/cart`, jsonHeaders(token));
    check(cartRes, { 'get cart status 200': (r) => r.status === 200 });
  });

  sleep(1);

  group('3. Checkout (COD)', function () {
    const checkoutRes = http.post(
      `${BASE_URL}/orders/checkout`,
      JSON.stringify({ address_id: addressId, payment_method: 'cod' }),
      jsonHeaders(token)
    );
    const ok = check(checkoutRes, {
      'checkout status 2xx': (r) => r.status >= 200 && r.status < 300,
    });
    if (!ok) {
      checkoutFailures.add(1);
      // Log the actual error once in a while so failures are diagnosable,
      // without flooding the console on every single iteration.
      if (Math.random() < 0.2) {
        console.log(`checkout failed: ${checkoutRes.status} ${checkoutRes.body}`);
      }
    }
  });

  sleep(1);

  group('4. View my orders', function () {
    const ordersRes = http.get(`${BASE_URL}/orders?page=1&limit=10`, jsonHeaders(token));
    check(ordersRes, { 'orders status 200': (r) => r.status === 200 });
  });

  journeyDuration.add(Date.now() - start);

  sleep(1);
}

import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL = __ENV.BASE_URL || 'https://ecommerce-backend-dd4u.onrender.com/api/v1';
const ADMIN_PHONE = __ENV.ADMIN_PHONE || '9999999999';
const SESSION_POOL_SIZE = Number(__ENV.SESSION_POOL_SIZE || 8);

// Coordinates near "Default Warehouse" (Mumbai, 19.1197, 72.8468, 10km radius)
// so checkout's serviceability check passes.
const TEST_LAT = 19.1197;
const TEST_LNG = 72.8468;

const checkoutFailures = new Counter('checkout_failures');
const otherFailures = new Counter('other_failures');

export const options = {
  setupTimeout: '5m',
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

function uniquePhone(prefix, n) {
  return `${prefix}${String(n).padStart(9, '0')}`.slice(0, 10);
}

function loginViaOtp(sendUrl, verifyUrl, phone, extra) {
  const sendRes = http.post(sendUrl, JSON.stringify({ phone }), jsonHeaders());
  if (sendRes.status !== 200) return null;
  const otp = JSON.parse(sendRes.body).otp;
  const verifyRes = http.post(verifyUrl, JSON.stringify({ phone, otp, ...extra }), jsonHeaders());
  if (verifyRes.status !== 200) return null;
  return JSON.parse(verifyRes.body).token;
}

// ---------------------------------------------------------------------------
// setup() - logs in a pool of customer sessions (respecting the shared
// /auth/send-otp rate limit of 5/min/IP), gives each a saved address, and
// grabs a real active coupon code (if one exists) to exercise coupon
// validation with a genuine happy path.
// ---------------------------------------------------------------------------
export function setup() {
  let authCount = 0;
  function pace() {
    authCount++;
    if (authCount % 5 === 0) {
      console.log('Pausing ~65s to respect the /auth/send-otp rate limit...');
      sleep(65);
    }
  }

  // Admin login, used only to look up an existing active coupon code.
  pace();
  const adminToken = loginViaOtp(
    `${BASE_URL}/auth/send-otp`,
    `${BASE_URL}/auth/verify-otp`,
    ADMIN_PHONE
  );
  let couponCode = null;
  if (adminToken) {
    const couponRes = http.get(`${BASE_URL}/admin/coupons`, jsonHeaders(adminToken));
    if (couponRes.status === 200) {
      try {
        const coupons = JSON.parse(couponRes.body) || [];
        const active = coupons.find((c) => c.is_active);
        if (active) couponCode = active.code;
      } catch (e) {}
    }
  }
  if (!couponCode) console.log('setup: no active coupon found - coupon validation will only be error-path tested.');

  const sessions = [];
  for (let i = 0; i < SESSION_POOL_SIZE; i++) {
    pace();
    const phone = uniquePhone('9', i);
    const token = loginViaOtp(
      `${BASE_URL}/auth/send-otp`,
      `${BASE_URL}/auth/verify-otp`,
      phone,
      { name: `LoadTest Session ${i}` }
    );
    if (!token) continue;

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
        addressId = JSON.parse(addrRes.body).id;
      } catch (e) {}
    }
    if (!addressId) continue;

    sessions.push({ token, addressId });
  }

  if (sessions.length === 0) throw new Error('setup: could not create any logged-in sessions.');

  console.log(`setup complete: ${sessions.length}/${SESSION_POOL_SIZE} sessions ready. Coupon: ${couponCode || 'none'}`);
  return { sessions, couponCode };
}

// ---------------------------------------------------------------------------
// Main VU flow - exercises the full customer-facing API surface using a
// reused, pre-authenticated session.
// ---------------------------------------------------------------------------
export default function (data) {
  const session = data.sessions[__VU % data.sessions.length];
  const { token, addressId } = session;

  let productId;

  group('1. Browse & search products', function () {
    const listRes = http.get(`${BASE_URL}/products?page=1&limit=20`, jsonHeaders(token));
    check(listRes, { 'products status 200': (r) => r.status === 200 });
    try {
      const products = JSON.parse(listRes.body).products || [];
      if (products.length > 0) productId = products[Math.floor(Math.random() * products.length)].id;
    } catch (e) {}

    const searchRes = http.get(`${BASE_URL}/products?search=milk&limit=10`, jsonHeaders(token));
    check(searchRes, { 'search status 200': (r) => r.status === 200 });
  });

  sleep(1);

  if (!productId) {
    sleep(1);
    return;
  }

  group('2. Product reviews', function () {
    const getRes = http.get(`${BASE_URL}/products/${productId}/reviews`, jsonHeaders(token));
    check(getRes, { 'get reviews status 200': (r) => r.status === 200 });

    const postRes = http.post(
      `${BASE_URL}/products/${productId}/reviews`,
      JSON.stringify({ rating: 1 + Math.floor(Math.random() * 5), comment: 'Load test review' }),
      jsonHeaders(token)
    );
    const ok = check(postRes, { 'post review 2xx': (r) => r.status >= 200 && r.status < 300 });
    if (!ok) otherFailures.add(1);
  });

  sleep(1);

  group('3. Wishlist', function () {
    const addRes = http.post(
      `${BASE_URL}/wishlist`,
      JSON.stringify({ product_id: productId }),
      jsonHeaders(token)
    );
    check(addRes, { 'add wishlist 2xx': (r) => r.status >= 200 && r.status < 300 });

    const getRes = http.get(`${BASE_URL}/wishlist`, jsonHeaders(token));
    check(getRes, { 'get wishlist status 200': (r) => r.status === 200 });
  });

  sleep(1);

  group('4. Coupon validate', function () {
    const code = data.couponCode || 'INVALIDCODE';
    const res = http.post(
      `${BASE_URL}/coupons/validate`,
      JSON.stringify({ code, order_amount: 500 }),
      jsonHeaders(token)
    );
    if (data.couponCode) {
      check(res, { 'coupon validate 200 (real code)': (r) => r.status === 200 });
    } else {
      // No real coupon exists in this environment - just confirm the
      // endpoint responds sanely (400 for an unknown code is correct).
      check(res, { 'coupon validate responds': (r) => r.status === 200 || r.status === 400 });
    }
  });

  sleep(1);

  group('5. Wallet & notifications', function () {
    const walletRes = http.get(`${BASE_URL}/wallet`, jsonHeaders(token));
    check(walletRes, { 'wallet status 200': (r) => r.status === 200 });

    const notifRes = http.get(`${BASE_URL}/notifications`, jsonHeaders(token));
    check(notifRes, { 'notifications status 200': (r) => r.status === 200 });
  });

  sleep(1);

  group('6. Add to cart', function () {
    const addRes = http.post(
      `${BASE_URL}/cart`,
      JSON.stringify({ product_id: productId, quantity: 1 }),
      jsonHeaders(token)
    );
    check(addRes, { 'add to cart 2xx': (r) => r.status >= 200 && r.status < 300 });

    const cartRes = http.get(`${BASE_URL}/cart`, jsonHeaders(token));
    check(cartRes, { 'get cart status 200': (r) => r.status === 200 });
  });

  sleep(1);

  group('7. Checkout (COD)', function () {
    const checkoutRes = http.post(
      `${BASE_URL}/orders/checkout`,
      JSON.stringify({ address_id: addressId, payment_method: 'cod' }),
      jsonHeaders(token)
    );
    const ok = check(checkoutRes, { 'checkout status 2xx': (r) => r.status >= 200 && r.status < 300 });
    if (!ok) {
      checkoutFailures.add(1);
      if (Math.random() < 0.2) console.log(`checkout failed: ${checkoutRes.status} ${checkoutRes.body}`);
    }
  });

  sleep(1);

  group('8. View my orders & tracking', function () {
    const ordersRes = http.get(`${BASE_URL}/orders?page=1&limit=10`, jsonHeaders(token));
    const ok = check(ordersRes, { 'orders status 200': (r) => r.status === 200 });
    if (ok) {
      try {
        const orders = JSON.parse(ordersRes.body).orders || [];
        if (orders.length > 0) {
          const trackRes = http.get(`${BASE_URL}/orders/${orders[0].id}/tracking`, jsonHeaders(token));
          check(trackRes, { 'order tracking status 200': (r) => r.status === 200 });
        }
      } catch (e) {}
    }
  });

  sleep(1);
}

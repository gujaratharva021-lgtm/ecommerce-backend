import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL = __ENV.BASE_URL || 'https://ecommerce-backend-dd4u.onrender.com/api/v1';

// Custom metrics (visible in the end-of-run summary)
const checkoutFailures = new Counter('checkout_failures');
const otpFailures = new Counter('otp_failures');
const journeyDuration = new Trend('full_journey_duration', true);

// ---------------------------------------------------------------------------
// Load profile - tweak stages to match what you want to test.
// Start small (this hits a free-tier Render backend) and increase gradually.
// ---------------------------------------------------------------------------
export const options = {
  stages: [
    { duration: '30s', target: 5 },   // ramp up to 5 virtual users
    { duration: '1m', target: 5 },    // hold at 5
    { duration: '30s', target: 15 },  // ramp up to 15
    { duration: '1m', target: 15 },   // hold at 15
    { duration: '30s', target: 0 },   // ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],   // 95% of requests should be under 3s
    http_req_failed: ['rate<0.05'],      // less than 5% of requests should fail
    checkout_failures: ['count<5'],
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function uniquePhone() {
  // 10-digit numeric phone, unique per VU + iteration so each simulated user
  // is a distinct account (User.Phone has a unique index in the DB).
  const vu = String(__VU).padStart(4, '0');
  const iter = String(__ITER).padStart(4, '0');
  return `9${vu}${iter}`.slice(0, 10).padEnd(10, '0');
}

function authHeaders(token) {
  return {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  };
}

// ---------------------------------------------------------------------------
// Main VU flow - one simulated customer's full journey
// ---------------------------------------------------------------------------
export default function () {
  const start = Date.now();
  const phone = uniquePhone();

  let token;

  group('1. Login via OTP', function () {
    const sendRes = http.post(
      `${BASE_URL}/auth/send-otp`,
      JSON.stringify({ phone }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    const sendOk = check(sendRes, {
      'send-otp status 200': (r) => r.status === 200,
      'send-otp returns code': (r) => {
        try {
          return !!JSON.parse(r.body).otp;
        } catch (e) {
          return false;
        }
      },
    });
    if (!sendOk) {
      otpFailures.add(1);
      return;
    }
    const otp = JSON.parse(sendRes.body).otp;

    const verifyRes = http.post(
      `${BASE_URL}/auth/verify-otp`,
      JSON.stringify({ phone, otp, name: `LoadTest VU${__VU}` }),
      { headers: { 'Content-Type': 'application/json' } }
    );
    const verifyOk = check(verifyRes, {
      'verify-otp status 200': (r) => r.status === 200,
      'verify-otp returns token': (r) => {
        try {
          return !!JSON.parse(r.body).token;
        } catch (e) {
          return false;
        }
      },
    });
    if (!verifyOk) {
      otpFailures.add(1);
      return;
    }
    token = JSON.parse(verifyRes.body).token;
  });

  if (!token) {
    sleep(1);
    return; // can't continue journey without auth
  }

  sleep(1);

  let firstProductId;

  group('2. Browse products', function () {
    const listRes = http.get(`${BASE_URL}/products?page=1&limit=20`, authHeaders(token));
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
      if (products.length > 0) firstProductId = products[0].id;
    } catch (e) {
      // leave firstProductId undefined; later steps will skip
    }

    // Simulate a search too
    const searchRes = http.get(`${BASE_URL}/products?search=milk&limit=10`, authHeaders(token));
    check(searchRes, { 'search status 200': (r) => r.status === 200 });
  });

  sleep(1);

  if (!firstProductId) {
    sleep(1);
    return; // nothing to add to cart
  }

  group('3. Add to cart', function () {
    const addRes = http.post(
      `${BASE_URL}/cart`,
      JSON.stringify({ product_id: firstProductId, quantity: 1 }),
      authHeaders(token)
    );
    check(addRes, { 'add to cart status 2xx': (r) => r.status >= 200 && r.status < 300 });

    const cartRes = http.get(`${BASE_URL}/cart`, authHeaders(token));
    check(cartRes, { 'get cart status 200': (r) => r.status === 200 });
  });

  sleep(1);

  let addressId;

  group('4. Add delivery address', function () {
    const addrRes = http.post(
      `${BASE_URL}/addresses`,
      JSON.stringify({
        label: 'Home',
        full_name: `LoadTest User ${__VU}`,
        phone,
        line1: '123 Test Street',
        line2: '',
        city: 'Mumbai',
        state: 'Maharashtra',
        pincode: '400001',
        is_default: true,
      }),
      authHeaders(token)
    );
    const ok = check(addrRes, {
      'create address status 2xx': (r) => r.status >= 200 && r.status < 300,
    });
    if (ok) {
      try {
        const body = JSON.parse(addrRes.body);
        addressId = body.id || (body.address && body.address.id);
      } catch (e) {
        // ignore
      }
    }
  });

  sleep(1);

  group('5. Checkout (COD)', function () {
    const payload = { payment_method: 'cod' };
    if (addressId) payload.address_id = addressId;

    const checkoutRes = http.post(
      `${BASE_URL}/orders/checkout`,
      JSON.stringify(payload),
      authHeaders(token)
    );
    const ok = check(checkoutRes, {
      'checkout status 2xx': (r) => r.status >= 200 && r.status < 300,
    });
    if (!ok) checkoutFailures.add(1);
  });

  sleep(1);

  group('6. View my orders', function () {
    const ordersRes = http.get(`${BASE_URL}/orders?page=1&limit=10`, authHeaders(token));
    check(ordersRes, { 'orders status 200': (r) => r.status === 200 });
  });

  journeyDuration.add(Date.now() - start);

  sleep(1);
}

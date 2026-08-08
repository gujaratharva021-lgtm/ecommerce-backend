import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter } from 'k6/metrics';

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const BASE_URL = __ENV.BASE_URL || 'https://ecommerce-backend-dd4u.onrender.com/api/v1';
const ADMIN_PHONE = __ENV.ADMIN_PHONE || '9999999999';

const DELIVERY_PARTNER_POOL_SIZE = Number(__ENV.DELIVERY_PARTNER_POOL_SIZE || 3);
const ORDERS_PER_PARTNER = Number(__ENV.ORDERS_PER_PARTNER || 2);
const WAREHOUSE_STAFF_POOL_SIZE = Number(__ENV.WAREHOUSE_STAFF_POOL_SIZE || 4); // split across 2 warehouses

const deliveryActionFailures = new Counter('delivery_action_failures');
const warehouseActionFailures = new Counter('warehouse_action_failures');

export const options = {
  setupTimeout: '8m', // creating partners/staff/orders respects OTP rate limits, so this takes a few minutes
  stages: [
    { duration: '30s', target: 4 },
    { duration: '1m', target: 4 },
    { duration: '30s', target: 8 },
    { duration: '1m', target: 8 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'],
    http_req_failed: ['rate<0.05'],
  },
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
function jsonHeaders(token) {
  const h = { 'Content-Type': 'application/json' };
  if (token) h.Authorization = `Bearer ${token}`;
  return { headers: h };
}

function loginViaOtp(sendUrl, verifyUrl, phone, extra) {
  const sendRes = http.post(sendUrl, JSON.stringify({ phone }), jsonHeaders());
  if (sendRes.status !== 200) {
    console.log(`login failed at send-otp for ${phone}: ${sendRes.status} ${sendRes.body}`);
    return null;
  }
  const otp = JSON.parse(sendRes.body).otp;
  const verifyRes = http.post(
    verifyUrl,
    JSON.stringify({ phone, otp, ...extra }),
    jsonHeaders()
  );
  if (verifyRes.status !== 200) {
    console.log(`login failed at verify-otp for ${phone}: ${verifyRes.status} ${verifyRes.body}`);
    return null;
  }
  return JSON.parse(verifyRes.body).token;
}

// /auth/send-otp, /delivery/send-otp and /warehouse/send-otp each have their
// OWN independent 5-requests/minute-per-IP budget (separate rate limiter
// instances). This helper waits out that budget when a given bucket's count
// crosses a multiple of 5.
function respectRateLimit(counterRef) {
  counterRef.count++;
  if (counterRef.count > 0 && counterRef.count % 5 === 0) {
    console.log(`Pausing ~65s to respect rate limit (bucket: ${counterRef.name}, count: ${counterRef.count})...`);
    sleep(65);
  }
}

// ---------------------------------------------------------------------------
// setup() - runs once. Builds delivery partners, warehouse staff, and a
// handful of confirmed orders for the delivery partners to work with.
// ---------------------------------------------------------------------------
export function setup() {
  const authBucket = { name: 'auth', count: 0 };
  const deliveryBucket = { name: 'delivery', count: 0 };
  const warehouseBucket = { name: 'warehouse', count: 0 };

  // --- 1. Admin login --------------------------------------------------
  respectRateLimit(authBucket);
  const adminToken = loginViaOtp(
    `${BASE_URL}/auth/send-otp`,
    `${BASE_URL}/auth/verify-otp`,
    ADMIN_PHONE
  );
  if (!adminToken) throw new Error('setup: admin login failed - check ADMIN_PHONE and that it is a real admin.');

  // --- 2. Fetch warehouses ----------------------------------------------
  const whRes = http.get(`${BASE_URL}/admin/warehouses`, jsonHeaders(adminToken));
  const warehouses = (whRes.status === 200 ? JSON.parse(whRes.body).warehouses : []) || [];
  if (warehouses.length < 2) {
    throw new Error(`setup: need at least 2 warehouses to test stock transfers, found ${warehouses.length}.`);
  }
  const warehouseA = warehouses[0];
  const warehouseB = warehouses[1];

  // --- 3. Fetch a few product ids for cart/transfer use ------------------
  const prodRes = http.get(`${BASE_URL}/products?page=1&limit=10`, jsonHeaders(adminToken));
  const products = (prodRes.status === 200 ? JSON.parse(prodRes.body).products : []) || [];
  const productIds = products.map((p) => p.id);
  if (productIds.length === 0) throw new Error('setup: no products found.');

  // --- 4. Create + log in delivery partners ------------------------------
  const deliveryPartners = [];
  for (let i = 0; i < DELIVERY_PARTNER_POOL_SIZE; i++) {
    const phone = `8${String(i).padStart(9, '0')}`.slice(0, 10);
    const createRes = http.post(
      `${BASE_URL}/admin/delivery-partners`,
      JSON.stringify({ name: `LoadTest Partner ${i}`, phone, vehicle_number: `MH01AB${1000 + i}` }),
      jsonHeaders(adminToken)
    );
    // 201 = newly created. 409 = already exists (e.g. from a previous run) -
    // that's fine, the phone is still usable to log in.
    if (createRes.status !== 201 && createRes.status !== 409) {
      console.log(`setup: could not create delivery partner ${i}: ${createRes.status} ${createRes.body}`);
      continue;
    }

    respectRateLimit(deliveryBucket);
    const token = loginViaOtp(
      `${BASE_URL}/delivery/send-otp`,
      `${BASE_URL}/delivery/verify-otp`,
      phone
    );
    if (!token) continue;
    deliveryPartners.push({ token });
  }
  if (deliveryPartners.length === 0) throw new Error('setup: no delivery partners could be created/logged in.');

  // --- 5. Create + log in warehouse staff (split across both warehouses) -
  const warehouseStaff = [];
  for (let i = 0; i < WAREHOUSE_STAFF_POOL_SIZE; i++) {
    const warehouse = i % 2 === 0 ? warehouseA : warehouseB;
    const phone = `7${String(i).padStart(9, '0')}`.slice(0, 10);
    const createRes = http.post(
      `${BASE_URL}/admin/warehouse-staff`,
      JSON.stringify({ name: `LoadTest Staff ${i}`, phone, warehouse_id: warehouse.id }),
      jsonHeaders(adminToken)
    );
    if (createRes.status !== 201 && createRes.status !== 409) {
      console.log(`setup: could not create warehouse staff ${i}: ${createRes.status} ${createRes.body}`);
      continue;
    }

    respectRateLimit(warehouseBucket);
    const token = loginViaOtp(
      `${BASE_URL}/warehouse/send-otp`,
      `${BASE_URL}/warehouse/verify-otp`,
      phone
    );
    if (!token) continue;
    warehouseStaff.push({ token, warehouseId: warehouse.id });
  }
  if (warehouseStaff.length === 0) throw new Error('setup: no warehouse staff could be created/logged in.');

  // --- 6. Create confirmed orders for delivery partners to work with -----
  // Each customer: login, save an address near warehouse A, add a random
  // product to cart, checkout COD (auto-confirmed), then admin assigns the
  // resulting order to a delivery partner (round robin).
  const totalOrdersNeeded = DELIVERY_PARTNER_POOL_SIZE * ORDERS_PER_PARTNER;
  let partnerCursor = 0;

  for (let i = 0; i < totalOrdersNeeded; i++) {
    respectRateLimit(authBucket);
    const phone = `6${String(i).padStart(9, '0')}`.slice(0, 10);
    const custToken = loginViaOtp(
      `${BASE_URL}/auth/send-otp`,
      `${BASE_URL}/auth/verify-otp`,
      phone,
      { name: `LoadTest Customer ${i}` }
    );
    if (!custToken) continue;

    const addrRes = http.post(
      `${BASE_URL}/addresses`,
      JSON.stringify({
        label: 'Home',
        full_name: `LoadTest Customer ${i}`,
        phone,
        line1: '123 Test Street',
        line2: '',
        city: 'Mumbai',
        state: 'Maharashtra',
        pincode: '400001',
        lat: warehouseA.lat,
        lng: warehouseA.lng,
        is_default: true,
      }),
      jsonHeaders(custToken)
    );
    let addressId;
    if (addrRes.status >= 200 && addrRes.status < 300) {
      try {
        const body = JSON.parse(addrRes.body);
        addressId = body.id || (body.address && body.address.id);
      } catch (e) {}
    }
    if (!addressId) {
      console.log(`setup: address failed for order-customer ${i}: ${addrRes.status} ${addrRes.body}`);
      continue;
    }

    const productId = productIds[Math.floor(Math.random() * productIds.length)];
    http.post(
      `${BASE_URL}/cart`,
      JSON.stringify({ product_id: productId, quantity: 1 }),
      jsonHeaders(custToken)
    );

    const checkoutRes = http.post(
      `${BASE_URL}/orders/checkout`,
      JSON.stringify({ address_id: addressId, payment_method: 'cod' }),
      jsonHeaders(custToken)
    );
    if (checkoutRes.status < 200 || checkoutRes.status >= 300) {
      console.log(`setup: checkout failed for order-customer ${i}: ${checkoutRes.status} ${checkoutRes.body}`);
      continue;
    }
    let orderId;
    try {
      const body = JSON.parse(checkoutRes.body);
      orderId = body.id || (body.order && body.order.id);
    } catch (e) {}
    if (!orderId) continue;

    // Assign to a delivery partner, round robin.
    const partner = deliveryPartners[partnerCursor % deliveryPartners.length];
    partnerCursor++;
    // We don't have the partner's numeric ID handy (only their token), so
    // fetch it via the admin delivery-partners list instead of tracking
    // separately -- simplest is to just look it up once per assignment.
    partner.orderReadyToAssign = orderId;
  }

  // Fetch delivery partner IDs (needed for the assignment call) and assign
  // any orders we tagged above. We match by name prefix since we only have
  // tokens, not numeric IDs, from the login step.
  const dpListRes = http.get(`${BASE_URL}/admin/delivery-partners`, jsonHeaders(adminToken));
  const dpList = (dpListRes.status === 200 ? JSON.parse(dpListRes.body).delivery_partners : []) || [];
  const loadTestPartners = dpList.filter((x) => (x.name || '').startsWith('LoadTest Partner'));
  let assignCursor = 0;
  for (const dp of deliveryPartners) {
    if (!dp.orderReadyToAssign) continue;
    const targetPartner = loadTestPartners[assignCursor % loadTestPartners.length];
    assignCursor++;
    if (!targetPartner) continue;
    const assignRes = http.put(
      `${BASE_URL}/admin/orders/${dp.orderReadyToAssign}/assign-delivery`,
      JSON.stringify({ delivery_partner_id: targetPartner.id }),
      jsonHeaders(adminToken)
    );
    if (assignRes.status !== 200) {
      console.log(`setup: assign-delivery failed for order ${dp.orderReadyToAssign}: ${assignRes.status} ${assignRes.body}`);
    }
  }

  console.log(
    `setup complete: ${deliveryPartners.length} delivery partners, ${warehouseStaff.length} warehouse staff, ` +
      `${assignCursor} orders assigned.`
  );

  return {
    adminToken,
    warehouses: [warehouseA, warehouseB],
    productIds,
    deliveryPartners,
    warehouseStaff,
  };
}

// ---------------------------------------------------------------------------
// Default VU function - alternates between the delivery-partner journey and
// the warehouse-staff journey based on VU index, each reusing a
// pre-authenticated session from the setup() pool.
// ---------------------------------------------------------------------------
export default function (data) {
  const isDeliveryVU = __VU % 2 === 0;

  if (isDeliveryVU) {
    runDeliveryPartnerJourney(data);
  } else {
    runWarehouseStaffJourney(data);
  }

  sleep(1);
}

function runDeliveryPartnerJourney(data) {
  const session = data.deliveryPartners[__VU % data.deliveryPartners.length];
  const token = session.token;
  const warehouse = data.warehouses[__VU % data.warehouses.length];

  group('Delivery: process an assigned order', function () {
    const listRes = http.get(`${BASE_URL}/delivery/orders?status=confirmed`, jsonHeaders(token));
    const ok = check(listRes, { 'my deliveries status 200': (r) => r.status === 200 });
    if (!ok) {
      deliveryActionFailures.add(1);
      return;
    }
    let orders = [];
    try {
      orders = JSON.parse(listRes.body).orders || [];
    } catch (e) {}

    if (orders.length > 0) {
      const orderId = orders[0].id;

      const shipRes = http.put(
        `${BASE_URL}/delivery/orders/${orderId}/status`,
        JSON.stringify({ status: 'shipped' }),
        jsonHeaders(token)
      );
      const shipOk = check(shipRes, { 'mark shipped 2xx': (r) => r.status >= 200 && r.status < 300 });
      if (!shipOk) deliveryActionFailures.add(1);

      sleep(1);

      const deliverRes = http.put(
        `${BASE_URL}/delivery/orders/${orderId}/deliver`,
        null,
        jsonHeaders(token)
      );
      const deliverOk = check(deliverRes, { 'mark delivered 2xx': (r) => r.status >= 200 && r.status < 300 });
      if (!deliverOk) deliveryActionFailures.add(1);
    }
  });

  sleep(1);

  group('Delivery: location ping', function () {
    const jitter = () => (Math.random() - 0.5) * 0.01; // ~1km jitter
    const locRes = http.put(
      `${BASE_URL}/delivery/location`,
      JSON.stringify({ lat: warehouse.lat + jitter(), lng: warehouse.lng + jitter() }),
      jsonHeaders(token)
    );
    check(locRes, { 'location ping 2xx': (r) => r.status >= 200 && r.status < 300 });
  });

  sleep(1);

  group('Delivery: view earnings', function () {
    const earnRes = http.get(`${BASE_URL}/delivery/earnings`, jsonHeaders(token));
    check(earnRes, { 'earnings status 200': (r) => r.status === 200 });
  });
}

function runWarehouseStaffJourney(data) {
  const session = data.warehouseStaff[__VU % data.warehouseStaff.length];
  const { token, warehouseId } = session;
  const otherWarehouse = data.warehouses.find((w) => w.id !== warehouseId) || data.warehouses[0];

  group('Warehouse: check my transfers', function () {
    const listRes = http.get(`${BASE_URL}/warehouse/stock-transfers`, jsonHeaders(token));
    const ok = check(listRes, { 'my transfers status 200': (r) => r.status === 200 });
    if (!ok) {
      warehouseActionFailures.add(1);
      return;
    }
    let transfers = [];
    try {
      transfers = JSON.parse(listRes.body).stock_transfers || [];
    } catch (e) {}

    // Approve one pending incoming transfer, if any.
    const pendingIncoming = transfers.find((t) => t.status === 'pending' && t.to_warehouse_id === warehouseId);
    if (pendingIncoming) {
      const approveRes = http.put(
        `${BASE_URL}/warehouse/stock-transfers/${pendingIncoming.id}/approve`,
        null,
        jsonHeaders(token)
      );
      const ok2 = check(approveRes, { 'approve transfer 2xx': (r) => r.status >= 200 && r.status < 300 });
      if (!ok2) {
        warehouseActionFailures.add(1);
        if (Math.random() < 0.3) console.log(`approve failed: ${approveRes.status} ${approveRes.body}`);
      }
    }

    // Receive one in-transit incoming transfer, if any.
    const inTransitIncoming = transfers.find((t) => t.status === 'in_transit' && t.to_warehouse_id === warehouseId);
    if (inTransitIncoming) {
      const receiveRes = http.put(
        `${BASE_URL}/warehouse/stock-transfers/${inTransitIncoming.id}/receive`,
        null,
        jsonHeaders(token)
      );
      const ok3 = check(receiveRes, { 'receive transfer 2xx': (r) => r.status >= 200 && r.status < 300 });
      if (!ok3) warehouseActionFailures.add(1);
    }
  });

  sleep(1);

  group('Warehouse: request a new transfer', function () {
    // Not every iteration requests a new transfer, to avoid runaway growth.
    if (Math.random() > 0.4) return;

    const productId = data.productIds[Math.floor(Math.random() * data.productIds.length)];
    const reqRes = http.post(
      `${BASE_URL}/warehouse/stock-transfers`,
      JSON.stringify({ product_id: productId, to_warehouse_id: otherWarehouse.id, quantity: 1 }),
      jsonHeaders(token)
    );
    const ok = check(reqRes, { 'request transfer 2xx': (r) => r.status >= 200 && r.status < 300 });
    if (!ok) warehouseActionFailures.add(1);
  });
}

# Ecommerce Backend — API Documentation

**Base URL (local):** `http://localhost:8080/api/v1` (confirm your `PORT` in `.env`)
**Auth:** Bearer JWT in `Authorization: Bearer <token>` header, except where marked *Public*.
**Content-Type:** `application/json` for all request bodies.

All endpoints below are taken directly from `internal/routes/routes.go` (develop branch) and verified against a live run of the Postman collection.

---

## Table of Contents
1. [Health](#health)
2. [Auth](#auth)
3. [Products & Categories (Public)](#products--categories-public)
4. [Cart](#cart)
5. [Addresses](#addresses)
6. [Coupons](#coupons)
7. [Checkout & Orders](#checkout--orders)
8. [Payment (Razorpay)](#payment-razorpay)
9. [Admin — Categories](#admin--categories)
10. [Admin — Products](#admin--products)
11. [Admin — Orders](#admin--orders)
12. [Admin — Coupons](#admin--coupons)
13. [Upload](#upload)
14. [Error Format](#error-format)
15. [Known Behaviors / Notes](#known-behaviors--notes)

---

## Health

### `GET /health`
*Public.* No `/api/v1` prefix — served directly at root.

**Response `200`**
```json
{
  "message": "Ecommerce backend is running",
  "status": "ok"
}
```

---

## Auth

### `POST /auth/send-otp`
*Public.* Rate-limited: **5 requests/minute** per the `RateLimit` middleware.

**Body**
```json
{ "phone": "9876543210" }
```
`phone`: required, exactly 10 digits, numeric.

**Response `200`**
```json
{
  "message": "OTP sent successfully",
  "expires_in_minutes": 5,
  "otp": "452305"
}
```
> ⚠️ **Dev-only field:** `otp` is only included in the response when `GIN_MODE != release`. It is logged server-side (`[DEV OTP] phone=... code=...`) either way. **Before deploying to production, remove this field from the response** — see the `SendOTP` handler comment; it's a placeholder for local testing without an SMS gateway.

---

### `POST /auth/verify-otp`
*Public.* Rate-limited: **10 requests/minute**.

Creates the user on first login, logs them in on repeat visits — this one endpoint covers both signup and login since phone + OTP is the only credential. OTP is valid for 5 minutes and single-use.

**Body**
```json
{ "phone": "9876543210", "otp": "452305" }
```
`phone`: required, 10 digits. `otp`: required, 6 digits.

**Response `200`**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 2,
    "name": "",
    "phone": "9876543210",
    "role": "customer",
    "created_at": "2026-07-30T12:41:27+05:30",
    "updated_at": "2026-07-30T12:41:27+05:30"
  }
}
```
`role` is `"customer"` for all new signups. To get an `"admin"` token, the user's row must be manually promoted in the DB (see [Admin auth note](#known-behaviors--notes)) and the OTP flow re-run so the new JWT is minted with `role=admin`.

---

### `GET /auth/me` 🔒
Returns the authenticated user's profile.

**Response `200`** — same `user` shape as above (flat, not wrapped).

---

### `PUT /auth/me` 🔒
Updates the display name only. Phone is intentionally excluded — it's the login identity and would need a separate OTP-verified flow to change.

**Body**
```json
{ "name": "Test Customer" }
```

---

## Products & Categories (Public)

### `GET /categories`
**Response `200`**
```json
{
  "categories": [
    { "id": 1, "name": "Electronics", "image_url": "", "created_at": "...", "updated_at": "..." }
  ]
}
```

### `GET /products`
Query params (all optional): `search`, `category_id`, `min_price`, `max_price`, `in_stock` (bool), `sort`, `page`, `limit`.

**Response `200`**
```json
{
  "products": [
    {
      "id": 1,
      "name": "Wireless Mouse",
      "description": "A smooth wireless mouse",
      "price": 499,
      "image_url": "",
      "category_id": 1,
      "category": { "...": "nested category object" },
      "inventory": { "id": 1, "product_id": 1, "stock": 47, "in_stock": true, "...": "..." },
      "created_at": "...",
      "updated_at": "..."
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 1,
  "total_pages": 1
}
```
> **Note:** the nested `inventory.product` field comes back as a mostly-zeroed placeholder object (id 0, empty strings) rather than being omitted — harmless, but worth knowing if you're strict-parsing the response on the frontend.

### `GET /products/:id`
Same object shape as a single item from the list above.

---

## Cart
All routes below require `Authorization: Bearer <token>`.

### `POST /cart`
Add a product to the current user's cart (creates the cart if it doesn't exist).

**Body**
```json
{ "product_id": 1, "quantity": 2 }
```
`product_id`: required, integer (do **not** quote it — it's `uint` on the Go side, sending it as a string returns a 400 unmarshal error). `quantity`: required, min 1.

**Response `200`** — returns the **full cart object**, not just the created item:
```json
{
  "id": 2,
  "items": [
    { "id": 6, "cart_id": 2, "product_id": 1, "product": { "...": "..." }, "quantity": 2, "created_at": "...", "updated_at": "..." }
  ],
  "total_items": 2,
  "total_amount": 998
}
```
> The cart-item id you need for the next two endpoints is `items[N].id` (e.g. `6`), **not** the top-level `id` (which is the cart's own id, e.g. `2`). Easy to mix up.

### `GET /cart`
Same response shape as above.

### `PUT /cart/:item_id`
**Body**
```json
{ "quantity": 3 }
```
`item_id` in the URL is the cart-item id (see note above).

### `DELETE /cart/:item_id`
No body. Removes the line item.

---

## Addresses
All routes require auth.

### `POST /addresses`
**Body**
```json
{
  "label": "Home",
  "full_name": "Test Customer",
  "phone": "9876543210",
  "line1": "123 MG Road",
  "line2": "Near City Mall",
  "city": "Ahmedabad",
  "state": "Gujarat",
  "pincode": "380001",
  "is_default": true
}
```
Required: `full_name`, `phone` (10 digits), `line1`, `city`, `state`, `pincode` (6 digits). `label` and `line2` are optional.

**Response `200`** — the created address object (flat, includes `id`).

### `GET /addresses`
```json
{ "addresses": [ { "...": "address object" } ] }
```

### `PUT /addresses/:id`
Same body shape as create.

### `PUT /addresses/:id/default`
No body — marks this address as the default, unsetting any previous default.

---

## Coupons

### `POST /coupons/validate` 🔒
Customer-facing — previews a discount before checkout.

**Body**
```json
{ "code": "WELCOME10", "order_amount": 1000 }
```

**Response `200`** (valid) or **`400`/`404`** with `{"error": "coupon not found"}` etc. for invalid/expired/inapplicable codes.

---

## Checkout & Orders
All routes require auth.

### `POST /orders/checkout`
Converts the current cart into an order and **empties the cart**.

**Body**
```json
{ "address_id": 2, "payment_method": "cod", "coupon_code": "" }
```
`address_id`: optional (uint). `payment_method`: optional, one of `"cod"` / `"online"` — if omitted, check your handler's default. `coupon_code`: optional.

**Response `200`**
```json
{
  "id": 6,
  "user_id": 2,
  "address_id": 2,
  "address": { "...": "..." },
  "items_amount": 1497,
  "delivery_charge": 0,
  "total_amount": 1497,
  "status": "pending",
  "payment_method": "cod",
  "payment_status": "pending",
  "items": [ { "...": "order line items, price snapshotted at time of order" } ],
  "created_at": "...",
  "updated_at": "..."
}
```
> **Delivery charge varies by payment method in current data** — a COD order for ₹1497 showed `delivery_charge: 0`, while an online order for ₹499 showed `delivery_charge: 50`. Confirm this is your intended pricing rule (e.g. "free delivery above ₹X" or "COD always free") rather than an accidental default.

### `GET /orders`
```json
{ "orders": [ { "...": "..." } ], "page": 1, "limit": 20, "total": 7, "total_pages": 1 }
```
Returns only the authenticated user's own orders.

### `GET /orders/:id`
Single order object, same shape as checkout response.

### `PUT /orders/:id/cancel`
No body. Sets `status: "cancelled"`. **Verified:** cancelling an order restores the deducted stock (e.g. stock went 44 → 47 after cancelling a qty-3 order) — confirm this still holds for partial/multi-item orders.

### `POST /orders/:id/payment`
Creates a Razorpay order for an order with `payment_method: "online"`. No body needed.

**Response `200`**
```json
{
  "razorpay_order_id": "order_TJd6a40xilDbIH",
  "amount": 54900,
  "currency": "INR",
  "key_id": "rzp_test_TJaWdH3HQ9JaWm",
  "order_id": 7
}
```
`amount` is in paise (₹549 → `54900`).

### `POST /orders/:id/payment/verify`
**Body**
```json
{
  "razorpay_order_id": "order_TJd6a40xilDbIH",
  "razorpay_payment_id": "pay_xxx",
  "razorpay_signature": "sig_xxx"
}
```
All three required. The signature is HMAC-verified server-side against `RAZORPAY_KEY_SECRET` — **cannot be produced without a real Razorpay test-mode checkout**. A fake/placeholder signature correctly returns:
```json
{ "error": "Payment signature verification failed" }
```
To test this end-to-end, run an actual test-mode payment (card `4111 1111 1111 1111`, any future expiry/CVV) through Razorpay Checkout against the `razorpay_order_id` above, then paste the real `payment_id`/`signature` here.

---

## Payment (Razorpay)
See `POST /orders/:id/payment` and `POST /orders/:id/payment/verify` above — grouped under Orders since they're order-scoped.

---

## Admin — Categories
All routes require auth **and** `role: admin` (`middleware.AdminOnly()`), returns `403`/`401` otherwise.

### `POST /admin/categories`
```json
{ "name": "Home Appliances", "image_url": "" }
```
`name` required. Duplicate names return:
```json
{ "error": "Category already exists or could not be created" }
```

### `PUT /admin/categories/:id`
Same body shape.

### `DELETE /admin/categories/:id`
No body.

---

## Admin — Products

### `POST /admin/products`
```json
{
  "name": "Wireless Mouse",
  "description": "2.4GHz wireless mouse",
  "price": 599,
  "image_url": "",
  "category_id": 4,
  "stock": 50
}
```
`category_id`: **integer, not a quoted string** — same `uint` unmarshal rule as `product_id` in the cart endpoint. `name`, `price` (>0), `category_id` required; `stock` optional (≥0, defaults to 0).

### `PUT /admin/products/:id`
Same body shape as create — full replace, not partial patch.

### `PUT /admin/products/:id/inventory`
```json
{ "stock": 75 }
```
Sets absolute stock (not a delta).

### `DELETE /admin/products/:id`
No body.

---

## Admin — Orders

### `GET /admin/orders`
Query params: `status`, `page`, `limit`. Returns **all** users' orders (unlike `GET /orders`).

### `PUT /admin/orders/:id/status`
```json
{ "status": "confirmed" }
```
Must be one of: `confirmed`, `shipped`, `delivered`, `cancelled`.

> ⚠️ **Worth reviewing:** in testing, an `online`-payment order was set to `status: "confirmed"` while its `payment_status` was still `"pending"` (unverified payment). If your business rule requires payment confirmation before order confirmation, add a check in `UpdateOrderStatus` to block `confirmed`/`shipped`/`delivered` transitions for online orders where `payment_status != "paid"`.

---

## Admin — Coupons

### `POST /admin/coupons`
```json
{
  "code": "WELCOME10",
  "discount_type": "percentage",
  "discount_value": 10,
  "min_order_amount": 500,
  "max_discount_amount": 200,
  "usage_limit": 100,
  "expiry_date": "2026-12-31"
}
```
`discount_type`: `"flat"` or `"percentage"`. `discount_value`: required, >0. `expiry_date` format: `"2006-01-02"` (Go reference format, i.e. `YYYY-MM-DD`). `max_discount_amount` is optional/nullable — omit it for flat-type coupons with no cap.

### `GET /admin/coupons`
Returns a **bare array**, not wrapped in an object:
```json
[
  { "id": 2, "code": "WELCOME10", "...": "..." },
  { "id": 1, "code": "SAVE50", "...": "..." }
]
```
> Inconsistent with most other list endpoints, which wrap results in `{ "orders": [...] }` / `{ "products": [...] }` etc. — worth aligning if you want a uniform response contract across the API.

### `PUT /admin/coupons/:id/status`
```json
{ "is_active": true }
```

---

## Upload

### `POST /upload` 🔒
Multipart form upload (not JSON). Structure is in place for product/category images but not yet wired into the product/category create flows — uploaded files are served back from `/uploads/<filename>` (static route registered at the root, no `/api/v1` prefix).

---

## Error Format

Errors are consistently shaped as:
```json
{ "error": "human-readable message" }
```
Validation errors (failed `binding` tags) return Gin's raw validator message, e.g.:
```json
{ "error": "Key: 'AddToCartRequest.Quantity' Error:Field validation for 'Quantity' failed on the 'min' tag" }
```
Malformed JSON types (e.g. sending a string where a number is expected) return:
```json
{ "error": "json: cannot unmarshal string into Go struct field ... of type uint" }
```

---

## Known Behaviors / Notes

- **Admin bootstrapping:** every new phone number signs up as `role: "customer"`. To create the first admin, sign up normally via OTP, then run:
  ```sql
  UPDATE users SET role='admin' WHERE phone='9999999999';
  ```
  ...and log in again (send-otp + verify-otp) so the new JWT is minted with `role=admin` baked into its claims — role is read from the JWT, not re-checked against the DB on every request.
- **Integer fields must not be quoted in JSON bodies:** `product_id`, `category_id`, etc. are Go `uint` — sending `"1"` instead of `1` fails with a 400 unmarshal error, not a friendlier validation message.
- **Rate limiting:** `/auth/send-otp` (5/min) and `/auth/verify-otp` (10/min) are rate-limited per the Day 7 security hardening pass. Expect `429`s if you hammer these in quick succession while testing.
- **CORS, JWT algorithm confusion, and OTP-leak fixes** were also part of the Day 7 hardening pass — see `internal/middleware/cors.go` and the auth middleware for specifics if you need to document those for a security review.

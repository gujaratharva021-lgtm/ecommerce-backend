# Ecommerce Backend (GoLang + PostgreSQL)

Developer 1 — Backend module. Day 1 deliverable: project setup, DB schema, and authentication APIs.

## Tech Stack
- Go 1.22+
- Gin (web framework)
- GORM (ORM) + PostgreSQL
- JWT (golang-jwt/jwt/v5) for authentication
- OTP-based login (phone number only — no email/password)
- Docker Compose for local PostgreSQL

## Folder Structure
```
ecommerce-backend/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── config/                  # Env config loader
│   ├── database/                # DB connection + migrations
│   ├── models/                  # GORM models (User, Product, Cart, Order, etc.)
│   ├── handlers/                # Request handlers (controllers)
│   ├── middleware/               # JWT auth middleware
│   ├── routes/                  # Route definitions
│   └── utils/                   # JWT helpers
├── migrations/                  # (raw SQL, optional — GORM auto-migrates by default)
├── docker-compose.yml           # Local Postgres + pgAdmin
├── .env.example
└── go.mod
```

## Setup Instructions

### 1. Clone and enter the project
```bash
git clone https://github.com/gujaratharva021-lgtm/ecommerce-backend.git
cd ecommerce-backend
```

### 2. Start PostgreSQL with Docker
```bash
docker compose up -d
```
This starts:
- PostgreSQL on `localhost:5432` (user: `postgres`, password: `postgres`, db: `ecommerce_db`)
- pgAdmin on `localhost:5050` (email: `admin@admin.com`, password: `admin`)

### 3. Configure environment variables
```bash
cp .env.example .env
```
Edit `.env` if you changed any Docker credentials. Change `JWT_SECRET` to a random string.

### 4. Install Go dependencies
```bash
go mod tidy
```

### 5. Run the server
```bash
go run cmd/api/main.go
```
Server starts on `http://localhost:8080`. On first run, GORM auto-migration creates all tables.

### 6. Verify it's running
```bash
curl http://localhost:8080/health
```

---

## API Documentation (Day 1)

Base URL: `http://localhost:8080/api/v1`

### Auth APIs — Phone Number + OTP (no email/password)

There is no separate signup vs login. `POST /auth/verify-otp` handles both:
first successful OTP verification for a phone number creates the account;
every verification after that just logs the existing user in.

#### 1. Send OTP
`POST /auth/send-otp`

Request body:
```json
{ "phone": "9999999999" }
```
Phone must be exactly 10 digits, numeric, no country code, no spaces/dashes.

Response `200 OK`:
```json
{
  "message": "OTP sent successfully",
  "otp": "042917",
  "expires_in_minutes": 5
}
```
⚠️ **Dev only**: no SMS gateway is wired up yet, so the OTP is returned directly in the response (and logged to the server console) purely so you can test locally. Before shipping this anywhere real, integrate an SMS provider (MSG91, Fast2SMS, Twilio, etc.) inside `SendOTP` in `auth_handler.go` and delete the `"otp"` field from the response — returning the code to the client defeats the point of OTP auth.

#### 2. Verify OTP (signup + login combined)
`POST /auth/verify-otp`

Request body:
```json
{ "phone": "9999999999", "otp": "042917" }
```

Response `200 OK`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "name": "",
    "phone": "9999999999",
    "role": "customer",
    "created_at": "2026-07-30T08:00:00Z",
    "updated_at": "2026-07-30T08:00:00Z"
  }
}
```
`name` is blank until the user sets it via a future "update profile" endpoint (Day 3).

Error responses:
- `400 Bad Request` — validation failed (phone not 10 digits, OTP not 6 digits)
- `401 Unauthorized` — wrong OTP, already-used OTP, or expired OTP (5 min validity)

#### 3. Get logged-in user profile (protected)
`GET /auth/me`

Headers:
```
Authorization: Bearer <token>
```

Response `200 OK`: the `user` object shown above.

Error responses:
- `401 Unauthorized` — missing/invalid/expired token

### Product APIs

#### 1. List products (search, filter, sort, paginate)
`GET /products`

Query params (all optional):
| Param | Type | Example | Notes |
|---|---|---|---|
| `search` | string | `?search=milk` | Matches name or description (case-insensitive) |
| `category_id` | uint | `?category_id=3` | |
| `min_price` / `max_price` | float | `?min_price=10&max_price=100` | |
| `in_stock` | bool | `?in_stock=true` | Requires an inventory row for the product |
| `sort` | string | `?sort=price_asc` | `price_asc`, `price_desc`, `name_asc`, `name_desc`, `newest` (default) |
| `page` | int | `?page=2` | Default `1` |
| `limit` | int | `?limit=20` | Default `20`, max `100` |

Response `200 OK`:
```json
{
  "products": [ { "id": 1, "name": "...", "price": 49.0, "category": {...}, "inventory": {...} } ],
  "page": 1,
  "limit": 20,
  "total": 57,
  "total_pages": 3
}
```

#### 2. Get product details
`GET /products/:id` → `200 OK` (single product with category + inventory) or `404 Not Found`

#### 3. List categories
`GET /categories` → `200 OK` → `{ "categories": [...] }`

### Cart APIs (all protected — require `Authorization: Bearer <token>`)

#### 1. Get current user's cart
`GET /cart`

Response `200 OK`:
```json
{
  "id": 1,
  "items": [ { "id": 4, "product_id": 2, "product": {...}, "quantity": 2 } ],
  "total_items": 2,
  "total_amount": 98.0
}
```

#### 2. Add item to cart
`POST /cart`

Request body:
```json
{ "product_id": 2, "quantity": 1 }
```
If the product is already in the cart, quantity is incremented rather than duplicated. Returns `201 Created` with the updated cart. `400` if the product is out of stock or doesn't have enough stock; `404` if the product doesn't exist.

#### 3. Update item quantity
`PUT /cart/:item_id`

Request body:
```json
{ "quantity": 3 }
```
Returns `200 OK` with the updated cart. `403 Forbidden` if the cart item doesn't belong to the logged-in user.

#### 4. Remove item from cart
`DELETE /cart/:item_id` → `200 OK` with the updated cart. `403 Forbidden` if it isn't your cart item.

### Upload API (protected)

`POST /upload` — `multipart/form-data` with a field named `image` (jpg/jpeg/png/webp, max 5MB).

Response `201 Created`:
```json
{ "image_url": "/uploads/1234567890.jpg" }
```
Use the returned `image_url` when creating/updating a Product or Category (Day 6 admin APIs). Files are served locally under `/uploads/...` for development — swap `SaveUploadedFile` for S3/GCS before production so images survive redeploys.

### Health Check
`GET /health` → `{"status": "ok", "message": "Ecommerce backend is running"}`

---

## Database Tables (Day 1)
- `users` — id, name, phone (unique), role, timestamps
- `otps` — id, phone, code, expires_at, verified, created_at
- `categories` — id, name, image_url, timestamps
- `products` — id, name, description, price, image_url, category_id, timestamps
- `inventories` — id, product_id, stock, in_stock, timestamps
- `carts` — id, user_id, timestamps
- `cart_items` — id, cart_id, product_id, quantity, timestamps
- `orders` — id, user_id, total_amount, status, timestamps
- `order_items` — id, order_id, product_id, quantity, price, created_at

Tables are auto-created via GORM `AutoMigrate` on server start — no manual SQL needed.

## Testing with Postman
1. Import base URL `http://localhost:8080/api/v1`
2. Call `POST /auth/send-otp` with a phone number, then `POST /auth/verify-otp` with that phone + the `otp` from the response → copy the `token`
3. For protected routes, add header: `Authorization: Bearer <token>`
4. Call `GET /auth/me` to verify token works

## Day 1 Checklist
- [x] Go project structure
- [x] PostgreSQL via Docker
- [x] Database schema (8 tables)
- [x] GORM models + auto-migration
- [x] Phone + OTP authentication (send-otp, verify-otp — combined signup/login)
- [x] JWT authentication
- [x] Auth middleware
- [x] API endpoints defined
- [x] API documentation (this file)
- [ ] Test APIs using Postman (do this after `go run`)
- [ ] Push to GitHub
- [ ] Share API docs with Flutter Developer

## Day 2 Checklist
- [x] Product APIs — list (search/filter/sort/paginate), detail
- [x] Category API — list
- [x] Image upload structure (local disk, ready to swap for S3/GCS)
- [x] Cart module — add / update quantity / remove / get, with per-user ownership checks
- [x] Stock validation on add-to-cart (checks `inventories.in_stock` + `inventories.stock`)
- [x] API documentation updated (this file)
- [ ] Test APIs using Postman (do this after `go run`)
- [ ] Push to GitHub
- [ ] Share new APIs with Flutter Developer

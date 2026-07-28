# Ecommerce Backend (GoLang + PostgreSQL)

Developer 1 — Backend module. Day 1 deliverable: project setup, DB schema, and authentication APIs.

## Tech Stack
- Go 1.22+
- Gin (web framework)
- GORM (ORM) + PostgreSQL
- JWT (golang-jwt/jwt/v5) for authentication
- bcrypt for password hashing
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
│   └── utils/                   # JWT + password helpers
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

### Auth APIs

#### 1. Signup
`POST /auth/signup`

Request body:
```json
{
  "name": "Harva Patel",
  "email": "harva@example.com",
  "password": "password123",
  "phone": "9999999999"
}
```

Response `201 Created`:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "name": "Harva Patel",
    "email": "harva@example.com",
    "phone": "9999999999",
    "role": "customer",
    "created_at": "2026-07-28T10:00:00Z",
    "updated_at": "2026-07-28T10:00:00Z"
  }
}
```

Error responses:
- `400 Bad Request` — validation failed (missing fields, invalid email, password < 6 chars)
- `409 Conflict` — email already registered

#### 2. Login
`POST /auth/login`

Request body:
```json
{
  "email": "harva@example.com",
  "password": "password123"
}
```

Response `200 OK`: same shape as Signup response (token + user).

Error responses:
- `400 Bad Request` — validation failed
- `401 Unauthorized` — invalid email or password

#### 3. Get logged-in user profile (protected)
`GET /auth/me`

Headers:
```
Authorization: Bearer <token>
```

Response `200 OK`:
```json
{
  "id": 1,
  "name": "Harva Patel",
  "email": "harva@example.com",
  "phone": "9999999999",
  "role": "customer",
  "created_at": "2026-07-28T10:00:00Z",
  "updated_at": "2026-07-28T10:00:00Z"
}
```

Error responses:
- `401 Unauthorized` — missing/invalid/expired token

---

### Placeholder Endpoints (structure ready, full logic on Day 2)

| Method | Endpoint | Protected | Status |
|--------|----------|-----------|--------|
| GET | `/products` | No | Stub |
| GET | `/products/:id` | No | Stub |
| GET | `/categories` | No | Stub |
| GET | `/cart` | Yes | Stub |
| POST | `/cart` | Yes | Stub |

### Health Check
`GET /health` → `{"status": "ok", "message": "Ecommerce backend is running"}`

---

## Database Tables (Day 1)
- `users` — id, name, email, password (hashed), phone, role, timestamps
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
2. Call `POST /auth/signup` → copy the `token` from response
3. For protected routes, add header: `Authorization: Bearer <token>`
4. Call `GET /auth/me` to verify token works

## Day 1 Checklist
- [x] Go project structure
- [x] PostgreSQL via Docker
- [x] Database schema (8 tables)
- [x] GORM models + auto-migration
- [x] Signup API + password hashing
- [x] Login API
- [x] JWT authentication
- [x] Auth middleware
- [x] API endpoints defined
- [x] API documentation (this file)
- [ ] Test APIs using Postman (do this after `go run`)
- [ ] Push to GitHub
- [ ] Share API docs with Flutter Developer

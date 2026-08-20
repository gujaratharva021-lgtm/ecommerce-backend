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

ecommerce-backend/
├── cmd/api/main.go # Entry point
├── internal/
│ ├── config/ # Env config loader
│ ├── database/ # DB connection + migrations
│ ├── models/ # GORM models (User, Product, Cart, Order, etc.)
│ ├── handlers/ # Request handlers (controllers)
│ ├── middleware/ # JWT auth middleware
│ ├── routes/ # Route definitions
│ └── utils/ # JWT helpers
├── migrations/ # (raw SQL, optional — GORM auto-migrates by default)
├── docker-compose.yml # Local Postgres + pgAdmin
├── .env.example
└── go.mod


## Setup Instructions

### 1. Clone and enter the project

git clone https://github.com/gujaratharva021-lgtm/ecommerce-backend.git
cd ecommerce-backend


### 2. Start PostgreSQL with Docker

docker compose up -d


This starts:

- PostgreSQL on `localhost:5432` (user: `postgres`, password: `postgres`, db: `ecommerce_db`)
- pgAdmin on `localhost:5050` (email: `admin@admin.com`, password: `admin`)

### 3. Configure environment variables

cp .env.example .env


Edit `.env` if you changed any Docker credentials. Change `JWT_SECRET` to a random string.

### 4. Install Go dependencies

go mod tidy


### 5. Run the server

go run cmd/api/main.go


Server starts on `http://localhost:8080`. On first run, GORM auto-migration creates all tables.

### 6. Verify it's running

curl http://localhost:8080/health


---

## Database Tables (Day 1–3)

- `users` — id, name, phone (unique), role, timestamps
- `otps` — id, phone, code, expires_at, verified, created_at
- `categories` — id, name, image_url, timestamps
- `products` — id, name, description, price, image_url, category_id, timestamps
- `inventories` — id, product_id, stock, in_stock, timestamps
- `carts` — id, user_id, timestamps
- `cart_items` — id, cart_id, product_id, quantity, timestamps
- `addresses` — id, user_id, label, full_name, phone, line1, line2, city, state, pincode, is_default, timestamps
- `orders` — id, user_id, address_id, items_amount, delivery_charge, total_amount, status, timestamps
- `order_items` — id, order_id, product_id, quantity, price, created_at

Tables are auto-created via GORM `AutoMigrate` on server start — no manual SQL needed.

## Day 1–5 summary

- Day 1: Project setup, DB schema, Auth APIs (phone + OTP)
- Day 2: Product/Category APIs, Cart module, image upload
- Day 3: Address module, Order management + checkout, profile update
- Day 4: Admin APIs — categories, products, inventory, order status management
- Day 5: Razorpay payment integration — online + COD checkout, signature verification

Full endpoint-level docs for these are in `API_DOCUMENTATION.md`.

## Day 6 — Finance & Accounting Module

### Overview
Full accounting/finance module added: vendor bills (accounts payable), a double-entry ledger with manual journal entries and trial balance, bank reconciliation, GST tracking (output GST from sales + vendor/purchase GST), and CA-ready Excel reports.

### New Features
- **Vendors & Vendor Bills** — track suppliers and their bills, with partial/full payment recording and GST amount per bill
- **Chart of Accounts & Ledger** — create accounts (asset/liability/equity/revenue/expense), post balanced manual journal entries, view trial balance
- **Bank Reconciliation** — record bank statement lines, match against internal records (vendor bill payments, etc.) or mark as ignored
- **GST Summary** — output GST (CGST/SGST/IGST) collected from sales invoices, broken down by rate and HSN code
- **Payments & Refunds** — collected/pending/refunded totals, online vs COD split, order counts by status
- **Custom Range Report** — pick any date range (or use 7/30/90-day presets) for a combined sales + GST + vendor-GST summary, with a 5-sheet Excel export (Summary, Orders, GST By Rate, GST By HSN, Vendor GST)
- **Delivery Partner Location** — admin can fetch a partner's last known GPS location (`GET /admin/delivery-partners/:id/location`)

### New API Endpoints

GET/POST/PUT/DELETE /api/v1/admin/finance/vendors
GET/POST/DELETE /api/v1/admin/finance/vendor-bills
POST /api/v1/admin/finance/vendor-bills/:id/pay
GET/POST/PUT /api/v1/admin/finance/accounts
GET/POST /api/v1/admin/finance/ledger
GET /api/v1/admin/finance/ledger/trial-balance
GET/POST /api/v1/admin/finance/bank-transactions
POST /api/v1/admin/finance/bank-transactions/:id/match
POST /api/v1/admin/finance/bank-transactions/:id/ignore
GET /api/v1/admin/finance/gst
GET /api/v1/admin/payments/reconciliation
GET /api/v1/admin/reports/range-sales
GET /api/v1/admin/reports/range-sales/export
GET /api/v1/admin/delivery-partners/:id/location


### Finance Panel (separate frontend)
New sidebar sections: Payments & Refunds, GST, and an "Accounting" group with Vendors, Vendor Bills, Chart of Accounts, Ledger, and Bank Reconciliation.

### Testing
- All endpoints tested live against the deployed backend
- End-to-end verified: created a vendor + GST-inclusive bill, recorded a partial payment, posted a balanced journal entry, matched a bank transaction, and downloaded/opened the range-report Excel export
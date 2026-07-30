# Deployment Checklist — Ecommerce Backend

Based on the actual code in `cmd/api/main.go`, `internal/config/config.go`, `internal/middleware/*.go`, and `migrations/README.md` as of the `develop` branch (commit `8e45ead`). Items are grouped by how much they'll bite you if skipped.

---

## 🔴 Must fix before any production deploy

- [ ] **Set a strong `JWT_SECRET`.** `config.go` already fails fast if `GIN_MODE=release` and the secret is still the placeholder (`default_secret_change_me`) — good. Just make sure the real value is a long random string (32+ bytes), not something guessable, and that it's injected via your host's secret manager, not committed to git.
  ```powershell
  # generate one
  node -e "console.log(require('crypto').randomBytes(48).toString('hex'))"
  ```
- [ ] **Set `GIN_MODE=release`.** This also strips the dev-only `otp` field from the `/auth/send-otp` response (see `SendOTP` handler) — without it, OTPs leak straight into the API response in production.
- [ ] **Remove the SMS gateway TODO.** `SendOTP` currently only logs the OTP to the server console (`log.Printf("[DEV OTP] ...")`) — there is no real SMS provider wired in yet. Without one, no real user can ever receive their OTP. Wire in MSG91 / Fast2SMS / Twilio (or similar) before this is usable by anyone outside your own testing.
- [ ] **Lock down CORS.** `internal/middleware/cors.go` currently sets `Access-Control-Allow-Origin: *` — any website can call this API from a browser using a logged-in user's token if it leaks. Restrict this to your actual frontend domain(s) before going live:
  ```go
  c.Header("Access-Control-Allow-Origin", "https://your-frontend-domain.com")
  ```
- [ ] **Set real Razorpay keys.** `.env.example` ships with `rzp_test_xxxxxxxxxxxxxx` placeholders. Confirm `RAZORPAY_KEY_ID` / `RAZORPAY_KEY_SECRET` are the **live** (not test) keys before accepting real payments, and that they're pulled from a secrets manager, not `.env` in the deployed image.
- [ ] **Set `DB_SSLMODE` correctly for your host.** `.env.example` defaults to `disable` (fine for local Docker Postgres). Most managed Postgres providers (RDS, Supabase, Neon, etc.) require `require` or `verify-full` — leaving it on `disable` against a public endpoint sends credentials and data unencrypted.
- [ ] **Move off `AutoMigrate` for schema changes going forward.** Per `migrations/README.md` (written by the project itself): `AutoMigrate` runs on *every* server start and can't do column renames, backfills, or rollbacks safely. Fine for the current dev phase, but before this has real customer data, switch to versioned migrations (`golang-migrate` or `goose`) — otherwise a routine deploy could silently mutate a live schema in ways you can't undo.

## 🟠 Should fix soon after launch

- [ ] **Rate limiter won't scale past one instance.** `internal/middleware/ratelimit.go` is explicitly in-memory (`sync.Mutex` + a map), keyed by client IP. The code comment says it plainly: *"if this app is ever scaled across multiple server instances, swap this for a shared store (e.g. Redis)."* If you deploy behind a load balancer with 2+ replicas, each instance tracks its own counts — so real limits become `limit × instance count`. Fine for a single-VM launch, not fine once you autoscale.
- [ ] **Confirm the delivery-charge rule is intentional.** Testing showed a COD order got `delivery_charge: 0` while a same-store online order got `delivery_charge: 50`. If this isn't an intentional "COD is always free" / "free above ₹X" rule, it's worth a quick look at the checkout handler before it ships to real customers.
- [ ] **Decide whether "confirmed" should require a paid online order.** `PUT /admin/orders/:id/status` currently lets an admin set `status: "confirmed"` on an online-payment order even while `payment_status` is still `"pending"`. If that's not intended, add a guard in `UpdateOrderStatus`.
- [ ] **Health check monitoring.** `GET /health` exists and works — hook it up to whatever uptime/monitoring tool your host uses (e.g. UptimeRobot, host's built-in health check, k8s liveness probe) so you get alerted on downtime instead of finding out from a customer.
- [ ] **Structured logging / error tracking.** Currently `log.Printf`/`log.Fatalf` go to stdout only. Consider piping to a log aggregator (or at minimum, making sure your host persists container logs) and wiring up an error tracker (Sentry or similar) so 500s surface somewhere you'll actually see them.

## 🟡 Good practice, not urgent

- [ ] **DB backups.** No backup strategy exists in the repo yet. At minimum, set up your DB host's automated daily backups (most managed Postgres offers this natively) and actually test a restore once before you need it for real.
- [ ] **Rotate `JWT_EXPIRY_HOURS` decision.** Currently 72h (3 days) — confirm this matches your product's desired session length; there's no refresh-token flow, so users get logged out and must re-OTP after this window.
- [ ] **`uploads/` storage.** `router.Static("/uploads", "./uploads")` serves files from local disk. This works on a single VM but **will not persist across deploys/restarts on most container platforms** (Render, Railway, Fly.io, Heroku-style ephemeral filesystems) and won't be shared across replicas. Move to S3 / Cloudinary / a managed object store before relying on uploaded images in production.
- [ ] **`.env` hygiene.** Confirm `.env` is in `.gitignore` (only `.env.example` should be committed) and that production secrets are injected via your host's environment variable / secrets UI, never baked into the Docker image.
- [ ] **API documentation kept in sync.** See `API_DOCUMENTATION.md` — update it alongside any route/handler changes so it doesn't drift from the real behavior.

---

## Suggested `.env` for production (fill in real values)

```env
PORT=8080
GIN_MODE=release

DB_HOST=<your-managed-postgres-host>
DB_PORT=5432
DB_USER=<db-user>
DB_PASSWORD=<strong-db-password>
DB_NAME=ecommerce_db
DB_SSLMODE=require

JWT_SECRET=<64+ char random string, generated once, never rotated casually>
JWT_EXPIRY_HOURS=72

RAZORPAY_KEY_ID=<live key from Razorpay dashboard>
RAZORPAY_KEY_SECRET=<live secret from Razorpay dashboard>
```

## Pre-launch smoke test (5 minutes)

Once deployed, re-run the critical path from the Postman collection against the **production** `base_url`:
1. `GET /health` → `200`
2. `POST /auth/send-otp` → confirm the `otp` field is **absent** from the response (proves `GIN_MODE=release` took effect) and that you actually receive an SMS
3. `POST /auth/verify-otp` → `200` with token
4. `GET /products` → `200`
5. One real COD checkout end-to-end
6. One real online payment through Razorpay Checkout (test a small ₹1 live transaction if your Razorpay account allows it, or use test mode against a staging deploy first)

If all six pass, you're in reasonable shape to open it up to real users.

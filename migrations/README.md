# migrations/

This project uses golang-migrate for versioned, production-safe schema migrations.

## How it works

- Dev/local (GIN_MODE=debug): the server still runs GORM AutoMigrate automatically on startup (see internal/database/database.go) for fast iteration.
- Production (GIN_MODE=release): AutoMigrate is skipped. Schema changes must be applied explicitly using the migrate CLI, via the versioned .sql files in this folder.

## Files

Each migration has an up (apply) and down (rollback) file:

000001_baseline_schema.up.sql
000001_baseline_schema.down.sql

000001_baseline_schema - initial schema snapshot (14 tables: users, otps, categories, products, inventories, carts, cart_items, addresses, orders, order_items, payments, coupons, order_coupons, notifications), including primary keys, indexes, and foreign key constraints.

## Common commands

Apply all pending migrations:
migrate -path migrations -database "postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable" up

Roll back the last migration:
migrate -path migrations -database "postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable" down 1

Create a new migration:
migrate create -ext sql -dir migrations -seq -digits 6 migration_name

If the installed migrate build errors on create, manually add NNNNNN_name.up.sql and NNNNNN_name.down.sql files following the same naming pattern.

## Notes

- Never edit an already-applied/committed migration file. Create a new migration instead.
- If a migration fails partway, the migrate tool marks the DB as dirty. Fix the issue manually in the DB, then use migrate force <version> to reset the tracked version before retrying.

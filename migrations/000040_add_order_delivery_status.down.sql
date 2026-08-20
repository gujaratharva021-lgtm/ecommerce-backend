DROP INDEX IF EXISTS idx_orders_delivery_status;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_status;

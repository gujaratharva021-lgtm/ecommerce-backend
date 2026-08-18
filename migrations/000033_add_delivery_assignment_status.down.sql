DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_delivery_assignment_status;
DROP INDEX IF EXISTS idx_orders_delivery_partner_id;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_rejection_reason;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_assignment_status;

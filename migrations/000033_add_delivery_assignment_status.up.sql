ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_assignment_status VARCHAR(20);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_rejection_reason TEXT;
CREATE INDEX IF NOT EXISTS idx_orders_delivery_partner_id ON orders(delivery_partner_id);
CREATE INDEX IF NOT EXISTS idx_orders_delivery_assignment_status ON orders(delivery_assignment_status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_status VARCHAR(20);
CREATE INDEX IF NOT EXISTS idx_orders_delivery_status ON orders(delivery_status);

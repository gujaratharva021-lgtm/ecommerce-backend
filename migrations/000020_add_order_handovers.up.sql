CREATE TABLE IF NOT EXISTS order_handovers (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id),
    warehouse_id BIGINT NOT NULL,
    warehouse_staff_id BIGINT NOT NULL,
    delivery_partner_id BIGINT NOT NULL,
    package_count INT NOT NULL DEFAULT 1,
    handed_over_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_order_handovers_warehouse ON order_handovers(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_order_handovers_partner ON order_handovers(delivery_partner_id);

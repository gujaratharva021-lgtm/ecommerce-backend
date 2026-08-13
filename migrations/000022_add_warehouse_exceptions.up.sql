CREATE TABLE IF NOT EXISTS warehouse_exceptions (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    product_id BIGINT REFERENCES products(id),
    warehouse_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    reason TEXT,
    priority VARCHAR(10) DEFAULT 'medium',
    staff_id BIGINT,
    status VARCHAR(20) DEFAULT 'open',
    resolution TEXT,
    resolved_by_id BIGINT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_warehouse_exceptions_order ON warehouse_exceptions(order_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_exceptions_status ON warehouse_exceptions(status);

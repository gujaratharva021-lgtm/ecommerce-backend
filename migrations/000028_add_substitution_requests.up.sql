CREATE TABLE IF NOT EXISTS substitution_requests (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    picking_task_item_id BIGINT NULL REFERENCES picking_task_items(id),
    original_product_id BIGINT NOT NULL REFERENCES products(id),
    substitute_product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    reason TEXT,
    warehouse_id BIGINT NOT NULL REFERENCES warehouses(id),
    requested_by_id BIGINT NOT NULL REFERENCES warehouse_staff(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    decided_by_id BIGINT NULL REFERENCES warehouse_staff(id),
    decision_note TEXT,
    decided_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_substitution_requests_order_id ON substitution_requests(order_id);
CREATE INDEX IF NOT EXISTS idx_substitution_requests_warehouse_id ON substitution_requests(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_substitution_requests_status ON substitution_requests(status);
CREATE INDEX IF NOT EXISTS idx_substitution_requests_picking_task_item_id ON substitution_requests(picking_task_item_id);

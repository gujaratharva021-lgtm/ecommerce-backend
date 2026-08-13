CREATE TABLE IF NOT EXISTS batches (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    warehouse_id BIGINT NOT NULL,
    batch_number VARCHAR(100) NOT NULL,
    manufacture_date TIMESTAMPTZ,
    expiry_date TIMESTAMPTZ NOT NULL,
    quantity INT NOT NULL,
    bin_id BIGINT REFERENCES warehouse_bins(id),
    created_by_staff_id BIGINT NOT NULL,
    receiving_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_batches_product ON batches(product_id);
CREATE INDEX IF NOT EXISTS idx_batches_warehouse ON batches(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_batches_expiry ON batches(expiry_date);

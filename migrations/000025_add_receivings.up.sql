CREATE TABLE IF NOT EXISTS receivings (
    id BIGSERIAL PRIMARY KEY,
    warehouse_id BIGINT NOT NULL,
    supplier_name VARCHAR(255) NOT NULL,
    reference_number VARCHAR(100),
    product_id BIGINT NOT NULL REFERENCES products(id),
    expected_quantity INT NOT NULL,
    received_quantity INT DEFAULT 0,
    damaged_quantity INT DEFAULT 0,
    accepted_quantity INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending',
    bin_id BIGINT REFERENCES warehouse_bins(id),
    created_by_staff_id BIGINT NOT NULL,
    received_by_staff_id BIGINT,
    qc_by_staff_id BIGINT,
    put_away_by_staff_id BIGINT,
    notes TEXT,
    rejection_reason TEXT,
    received_at TIMESTAMPTZ,
    qc_at TIMESTAMPTZ,
    put_away_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_receivings_warehouse ON receivings(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_receivings_product ON receivings(product_id);
CREATE INDEX IF NOT EXISTS idx_receivings_status ON receivings(status);

CREATE TABLE IF NOT EXISTS warehouse_zones (
    id BIGSERIAL PRIMARY KEY,
    warehouse_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(warehouse_id, name)
);

CREATE TABLE IF NOT EXISTS warehouse_racks (
    id BIGSERIAL PRIMARY KEY,
    zone_id BIGINT NOT NULL REFERENCES warehouse_zones(id),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(zone_id, name)
);

CREATE TABLE IF NOT EXISTS warehouse_bins (
    id BIGSERIAL PRIMARY KEY,
    rack_id BIGINT NOT NULL REFERENCES warehouse_racks(id),
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(rack_id, name)
);

ALTER TABLE inventories ADD COLUMN IF NOT EXISTS bin_id BIGINT REFERENCES warehouse_bins(id);
CREATE INDEX IF NOT EXISTS idx_inventories_bin ON inventories(bin_id);

ALTER TABLE warehouses ADD COLUMN IF NOT EXISTS service_area geometry(Polygon, 4326);

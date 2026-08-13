ALTER TABLE inventories DROP COLUMN IF EXISTS bin_id;
DROP TABLE IF EXISTS warehouse_bins;
DROP TABLE IF EXISTS warehouse_racks;
DROP TABLE IF EXISTS warehouse_zones;

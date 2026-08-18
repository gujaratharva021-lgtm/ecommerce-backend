INSERT INTO warehouse_staffs (name, phone, warehouse_id, is_active, created_at, updated_at)
VALUES ('Atharva Gujar Staff', '9876500001', 2, true, now(), now())
RETURNING id, name, phone, warehouse_id;

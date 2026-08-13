ALTER TABLE public.warehouse_staffs DROP CONSTRAINT IF EXISTS chk_warehouse_staffs_role;
ALTER TABLE public.warehouse_staffs DROP COLUMN IF EXISTS role;

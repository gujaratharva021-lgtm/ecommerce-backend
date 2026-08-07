DROP INDEX IF EXISTS public.idx_categories_deleted_at;
ALTER TABLE public.categories DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS public.idx_warehouse_staffs_deleted_at;
ALTER TABLE public.warehouse_staffs DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS public.idx_delivery_partners_deleted_at;
ALTER TABLE public.delivery_partners DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS public.idx_warehouses_deleted_at;
ALTER TABLE public.warehouses DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS public.idx_products_deleted_at;
ALTER TABLE public.products DROP COLUMN IF EXISTS deleted_at;

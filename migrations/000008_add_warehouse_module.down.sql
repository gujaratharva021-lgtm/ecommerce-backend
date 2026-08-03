DROP INDEX IF EXISTS public.idx_product_warehouse;
CREATE UNIQUE INDEX idx_inventories_product_id ON public.inventories USING btree (product_id);
ALTER TABLE public.inventories DROP CONSTRAINT IF EXISTS fk_inventories_warehouse;
ALTER TABLE public.inventories DROP COLUMN IF EXISTS warehouse_id;

DROP TABLE IF EXISTS public.stock_transfers CASCADE;
DROP TABLE IF EXISTS public.warehouse_staffs CASCADE;
DROP TABLE IF EXISTS public.warehouses CASCADE;

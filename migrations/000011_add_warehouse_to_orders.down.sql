DROP INDEX IF EXISTS public.idx_orders_warehouse_id;
ALTER TABLE ONLY public.orders DROP CONSTRAINT IF EXISTS fk_orders_warehouse;
ALTER TABLE public.orders DROP COLUMN IF EXISTS warehouse_id;
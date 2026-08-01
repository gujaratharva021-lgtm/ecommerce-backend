ALTER TABLE public.products ADD COLUMN deleted_at timestamp with time zone;
CREATE INDEX idx_products_deleted_at ON public.products USING btree (deleted_at);

ALTER TABLE public.warehouses ADD COLUMN deleted_at timestamp with time zone;
CREATE INDEX idx_warehouses_deleted_at ON public.warehouses USING btree (deleted_at);

ALTER TABLE public.delivery_partners ADD COLUMN deleted_at timestamp with time zone;
CREATE INDEX idx_delivery_partners_deleted_at ON public.delivery_partners USING btree (deleted_at);

ALTER TABLE public.warehouse_staffs ADD COLUMN deleted_at timestamp with time zone;
CREATE INDEX idx_warehouse_staffs_deleted_at ON public.warehouse_staffs USING btree (deleted_at);

ALTER TABLE public.categories ADD COLUMN deleted_at timestamp with time zone;
CREATE INDEX idx_categories_deleted_at ON public.categories USING btree (deleted_at);

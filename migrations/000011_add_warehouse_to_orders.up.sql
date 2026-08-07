-- Add warehouse_id to orders (order routing / dark-store fulfillment).
-- Nullable: existing orders were placed before warehouse routing existed
-- and have no warehouse to point to. New orders get one assigned at
-- checkout via the nearest-serviceable-warehouse lookup.
ALTER TABLE public.orders ADD COLUMN warehouse_id bigint;
ALTER TABLE ONLY public.orders ADD CONSTRAINT fk_orders_warehouse FOREIGN KEY (warehouse_id) REFERENCES public.warehouses(id);
CREATE INDEX idx_orders_warehouse_id ON public.orders USING btree (warehouse_id);
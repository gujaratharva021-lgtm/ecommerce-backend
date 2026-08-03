-- Warehouses table (multi-warehouse support)
CREATE TABLE public.warehouses (
    id bigint NOT NULL,
    name text NOT NULL,
    city text NOT NULL,
    address text,
    lat numeric,
    lng numeric,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    service_radius_km numeric DEFAULT 5
);
CREATE SEQUENCE public.warehouses_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.warehouses_id_seq OWNED BY public.warehouses.id;
ALTER TABLE ONLY public.warehouses ALTER COLUMN id SET DEFAULT nextval('public.warehouses_id_seq'::regclass);
ALTER TABLE ONLY public.warehouses ADD CONSTRAINT warehouses_pkey PRIMARY KEY (id);
CREATE INDEX idx_warehouses_city ON public.warehouses USING btree (city);

-- Seed a default warehouse so existing inventory rows have somewhere to
-- point once warehouse_id becomes required below. This is a genuine data
-- migration, not just a schema change: production has real inventory rows
-- with no warehouse concept today, and they must not lose their stock.
INSERT INTO public.warehouses (id, name, city, is_active, created_at, updated_at, service_radius_km)
VALUES (1, 'Default Warehouse', 'Unknown', true, now(), now(), 5);
SELECT setval('public.warehouses_id_seq', 1, true);

-- Warehouse staff table
CREATE TABLE public.warehouse_staffs (
    id bigint NOT NULL,
    name text NOT NULL,
    phone text NOT NULL,
    warehouse_id bigint NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE public.warehouse_staffs_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.warehouse_staffs_id_seq OWNED BY public.warehouse_staffs.id;
ALTER TABLE ONLY public.warehouse_staffs ALTER COLUMN id SET DEFAULT nextval('public.warehouse_staffs_id_seq'::regclass);
ALTER TABLE ONLY public.warehouse_staffs ADD CONSTRAINT warehouse_staffs_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX idx_warehouse_staffs_phone ON public.warehouse_staffs USING btree (phone);
ALTER TABLE ONLY public.warehouse_staffs ADD CONSTRAINT fk_warehouse_staffs_warehouse FOREIGN KEY (warehouse_id) REFERENCES public.warehouses(id);

-- Stock transfers table
CREATE TABLE public.stock_transfers (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    from_warehouse_id bigint NOT NULL,
    to_warehouse_id bigint NOT NULL,
    quantity bigint NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    requested_by bigint NOT NULL,
    approved_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE public.stock_transfers_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.stock_transfers_id_seq OWNED BY public.stock_transfers.id;
ALTER TABLE ONLY public.stock_transfers ALTER COLUMN id SET DEFAULT nextval('public.stock_transfers_id_seq'::regclass);
ALTER TABLE ONLY public.stock_transfers ADD CONSTRAINT stock_transfers_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.stock_transfers ADD CONSTRAINT fk_stock_transfers_product FOREIGN KEY (product_id) REFERENCES public.products(id);
ALTER TABLE ONLY public.stock_transfers ADD CONSTRAINT fk_stock_transfers_from_warehouse FOREIGN KEY (from_warehouse_id) REFERENCES public.warehouses(id);
ALTER TABLE ONLY public.stock_transfers ADD CONSTRAINT fk_stock_transfers_to_warehouse FOREIGN KEY (to_warehouse_id) REFERENCES public.warehouses(id);

-- Add warehouse_id to inventories, backfill existing rows to the default
-- warehouse, then enforce NOT NULL and swap the unique constraint from
-- (product_id) alone to (product_id, warehouse_id).
ALTER TABLE public.inventories ADD COLUMN warehouse_id bigint;
UPDATE public.inventories SET warehouse_id = 1;
ALTER TABLE public.inventories ALTER COLUMN warehouse_id SET NOT NULL;
ALTER TABLE ONLY public.inventories ADD CONSTRAINT fk_inventories_warehouse FOREIGN KEY (warehouse_id) REFERENCES public.warehouses(id);
DROP INDEX IF EXISTS public.idx_inventories_product_id;
CREATE UNIQUE INDEX idx_product_warehouse ON public.inventories USING btree (product_id, warehouse_id);

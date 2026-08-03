CREATE TABLE public.delivery_partners (
    id bigint NOT NULL,
    name text NOT NULL,
    phone text NOT NULL,
    vehicle_number text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE public.delivery_partners_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.delivery_partners_id_seq OWNED BY public.delivery_partners.id;
ALTER TABLE ONLY public.delivery_partners ALTER COLUMN id SET DEFAULT nextval('public.delivery_partners_id_seq'::regclass);

ALTER TABLE ONLY public.delivery_partners ADD CONSTRAINT delivery_partners_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX idx_delivery_partners_phone ON public.delivery_partners USING btree (phone);

ALTER TABLE public.orders ADD COLUMN delivery_partner_id bigint;
CREATE INDEX idx_orders_delivery_partner_id ON public.orders USING btree (delivery_partner_id);
ALTER TABLE ONLY public.orders ADD CONSTRAINT fk_orders_delivery_partner FOREIGN KEY (delivery_partner_id) REFERENCES public.delivery_partners(id);

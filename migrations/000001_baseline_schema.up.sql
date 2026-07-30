CREATE TABLE public.addresses (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    label text,
    full_name text NOT NULL,
    phone text NOT NULL,
    line1 text NOT NULL,
    line2 text,
    city text NOT NULL,
    state text NOT NULL,
    pincode text NOT NULL,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.cart_items (
    id bigint NOT NULL,
    cart_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity bigint DEFAULT 1,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.carts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.categories (
    id bigint NOT NULL,
    name text NOT NULL,
    image_url text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.coupons (
    id bigint NOT NULL,
    code text NOT NULL,
    discount_type text,
    discount_value numeric,
    min_order_amount numeric DEFAULT 0,
    max_discount_amount numeric,
    usage_limit bigint DEFAULT 1,
    used_count bigint DEFAULT 0,
    expiry_date timestamp with time zone,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.inventories (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    stock bigint DEFAULT 0,
    in_stock boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.notifications (
    id bigint NOT NULL,
    order_id bigint,
    phone text,
    message text,
    type text,
    status text,
    created_at timestamp with time zone
);

CREATE TABLE public.order_coupons (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    coupon_id bigint NOT NULL,
    discount_amount numeric,
    created_at timestamp with time zone
);

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity bigint NOT NULL,
    price numeric NOT NULL,
    created_at timestamp with time zone
);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total_amount numeric NOT NULL,
    status text DEFAULT 'pending'::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    address_id bigint NOT NULL,
    items_amount numeric NOT NULL,
    delivery_charge numeric DEFAULT 0 NOT NULL,
    payment_method text DEFAULT 'cod'::text,
    payment_status text DEFAULT 'pending'::text
);

CREATE TABLE public.otps (
    id bigint NOT NULL,
    phone text NOT NULL,
    code text NOT NULL,
    expires_at timestamp with time zone,
    verified boolean DEFAULT false,
    created_at timestamp with time zone
);

CREATE TABLE public.payments (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    razorpay_order_id text,
    razorpay_payment_id text,
    razorpay_signature text,
    amount numeric,
    currency text DEFAULT 'INR'::text,
    status text DEFAULT 'created'::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    description text,
    price numeric NOT NULL,
    image_url text,
    category_id bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE TABLE public.users (
    id bigint NOT NULL,
    name text,
    phone text NOT NULL,
    role text DEFAULT 'customer'::text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE public.addresses_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.addresses_id_seq OWNED BY public.addresses.id;
ALTER TABLE ONLY public.addresses ALTER COLUMN id SET DEFAULT nextval('public.addresses_id_seq'::regclass);

CREATE SEQUENCE public.cart_items_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.cart_items_id_seq OWNED BY public.cart_items.id;
ALTER TABLE ONLY public.cart_items ALTER COLUMN id SET DEFAULT nextval('public.cart_items_id_seq'::regclass);

CREATE SEQUENCE public.carts_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.carts_id_seq OWNED BY public.carts.id;
ALTER TABLE ONLY public.carts ALTER COLUMN id SET DEFAULT nextval('public.carts_id_seq'::regclass);

CREATE SEQUENCE public.categories_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.categories_id_seq OWNED BY public.categories.id;
ALTER TABLE ONLY public.categories ALTER COLUMN id SET DEFAULT nextval('public.categories_id_seq'::regclass);

CREATE SEQUENCE public.coupons_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.coupons_id_seq OWNED BY public.coupons.id;
ALTER TABLE ONLY public.coupons ALTER COLUMN id SET DEFAULT nextval('public.coupons_id_seq'::regclass);

CREATE SEQUENCE public.inventories_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.inventories_id_seq OWNED BY public.inventories.id;
ALTER TABLE ONLY public.inventories ALTER COLUMN id SET DEFAULT nextval('public.inventories_id_seq'::regclass);

CREATE SEQUENCE public.notifications_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.notifications_id_seq OWNED BY public.notifications.id;
ALTER TABLE ONLY public.notifications ALTER COLUMN id SET DEFAULT nextval('public.notifications_id_seq'::regclass);

CREATE SEQUENCE public.order_coupons_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.order_coupons_id_seq OWNED BY public.order_coupons.id;
ALTER TABLE ONLY public.order_coupons ALTER COLUMN id SET DEFAULT nextval('public.order_coupons_id_seq'::regclass);

CREATE SEQUENCE public.order_items_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.order_items_id_seq OWNED BY public.order_items.id;
ALTER TABLE ONLY public.order_items ALTER COLUMN id SET DEFAULT nextval('public.order_items_id_seq'::regclass);

CREATE SEQUENCE public.orders_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.orders_id_seq OWNED BY public.orders.id;
ALTER TABLE ONLY public.orders ALTER COLUMN id SET DEFAULT nextval('public.orders_id_seq'::regclass);

CREATE SEQUENCE public.otps_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.otps_id_seq OWNED BY public.otps.id;
ALTER TABLE ONLY public.otps ALTER COLUMN id SET DEFAULT nextval('public.otps_id_seq'::regclass);

CREATE SEQUENCE public.payments_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.payments_id_seq OWNED BY public.payments.id;
ALTER TABLE ONLY public.payments ALTER COLUMN id SET DEFAULT nextval('public.payments_id_seq'::regclass);

CREATE SEQUENCE public.products_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.products_id_seq OWNED BY public.products.id;
ALTER TABLE ONLY public.products ALTER COLUMN id SET DEFAULT nextval('public.products_id_seq'::regclass);

CREATE SEQUENCE public.users_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;
ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.addresses ADD CONSTRAINT addresses_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cart_items ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.carts ADD CONSTRAINT carts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.categories ADD CONSTRAINT categories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.coupons ADD CONSTRAINT coupons_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.inventories ADD CONSTRAINT inventories_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.notifications ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.order_coupons ADD CONSTRAINT order_coupons_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.order_items ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.orders ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.otps ADD CONSTRAINT otps_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.payments ADD CONSTRAINT payments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.products ADD CONSTRAINT products_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE INDEX idx_addresses_user_id ON public.addresses USING btree (user_id);
CREATE INDEX idx_cart_items_cart_id ON public.cart_items USING btree (cart_id);
CREATE UNIQUE INDEX idx_carts_user_id ON public.carts USING btree (user_id);
CREATE UNIQUE INDEX idx_categories_name ON public.categories USING btree (name);
CREATE UNIQUE INDEX idx_coupons_code ON public.coupons USING btree (code);
CREATE UNIQUE INDEX idx_inventories_product_id ON public.inventories USING btree (product_id);
CREATE UNIQUE INDEX idx_order_coupons_order_id ON public.order_coupons USING btree (order_id);
CREATE INDEX idx_order_items_order_id ON public.order_items USING btree (order_id);
CREATE INDEX idx_orders_user_id ON public.orders USING btree (user_id);
CREATE INDEX idx_otps_phone ON public.otps USING btree (phone);
CREATE UNIQUE INDEX idx_payments_order_id ON public.payments USING btree (order_id);
CREATE INDEX idx_products_category_id ON public.products USING btree (category_id);
CREATE INDEX idx_products_name ON public.products USING btree (name);
CREATE INDEX idx_products_price ON public.products USING btree (price);
CREATE UNIQUE INDEX idx_users_phone ON public.users USING btree (phone);

ALTER TABLE ONLY public.addresses ADD CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.cart_items ADD CONSTRAINT fk_cart_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);
ALTER TABLE ONLY public.cart_items ADD CONSTRAINT fk_carts_items FOREIGN KEY (cart_id) REFERENCES public.carts(id);
ALTER TABLE ONLY public.carts ADD CONSTRAINT fk_carts_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.order_coupons ADD CONSTRAINT fk_order_coupons_coupon FOREIGN KEY (coupon_id) REFERENCES public.coupons(id);
ALTER TABLE ONLY public.order_coupons ADD CONSTRAINT fk_order_coupons_order FOREIGN KEY (order_id) REFERENCES public.orders(id);
ALTER TABLE ONLY public.order_items ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES public.products(id);
ALTER TABLE ONLY public.orders ADD CONSTRAINT fk_orders_address FOREIGN KEY (address_id) REFERENCES public.addresses(id);
ALTER TABLE ONLY public.order_items ADD CONSTRAINT fk_orders_items FOREIGN KEY (order_id) REFERENCES public.orders(id);
ALTER TABLE ONLY public.orders ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.payments ADD CONSTRAINT fk_payments_order FOREIGN KEY (order_id) REFERENCES public.orders(id);
ALTER TABLE ONLY public.products ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES public.categories(id);
ALTER TABLE ONLY public.inventories ADD CONSTRAINT fk_products_inventory FOREIGN KEY (product_id) REFERENCES public.products(id);

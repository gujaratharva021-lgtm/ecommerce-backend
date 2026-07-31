CREATE TABLE public.wishlists (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    created_at timestamp with time zone
);

CREATE SEQUENCE public.wishlists_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.wishlists_id_seq OWNED BY public.wishlists.id;
ALTER TABLE ONLY public.wishlists ALTER COLUMN id SET DEFAULT nextval('public.wishlists_id_seq'::regclass);

ALTER TABLE ONLY public.wishlists ADD CONSTRAINT wishlists_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX idx_wishlist_user_product ON public.wishlists USING btree (user_id, product_id);

ALTER TABLE ONLY public.wishlists ADD CONSTRAINT fk_wishlists_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.wishlists ADD CONSTRAINT fk_wishlists_product FOREIGN KEY (product_id) REFERENCES public.products(id);

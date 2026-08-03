CREATE TABLE public.reviews (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    rating bigint NOT NULL,
    comment text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE public.reviews_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.reviews_id_seq OWNED BY public.reviews.id;
ALTER TABLE ONLY public.reviews ALTER COLUMN id SET DEFAULT nextval('public.reviews_id_seq'::regclass);

ALTER TABLE ONLY public.reviews ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX idx_review_user_product ON public.reviews USING btree (user_id, product_id);

ALTER TABLE ONLY public.reviews ADD CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.reviews ADD CONSTRAINT fk_reviews_product FOREIGN KEY (product_id) REFERENCES public.products(id);

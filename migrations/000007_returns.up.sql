CREATE TABLE public.return_requests (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    user_id bigint NOT NULL,
    reason text NOT NULL,
    status text DEFAULT 'pending'::text,
    refund_amount numeric NOT NULL DEFAULT 0,
    processed_by bigint,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);
CREATE SEQUENCE public.return_requests_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.return_requests_id_seq OWNED BY public.return_requests.id;
ALTER TABLE ONLY public.return_requests ALTER COLUMN id SET DEFAULT nextval('public.return_requests_id_seq'::regclass);
ALTER TABLE ONLY public.return_requests ADD CONSTRAINT return_requests_pkey PRIMARY KEY (id);
CREATE INDEX idx_return_requests_order_id ON public.return_requests USING btree (order_id);
CREATE INDEX idx_return_requests_user_id ON public.return_requests USING btree (user_id);
ALTER TABLE ONLY public.return_requests ADD CONSTRAINT fk_return_requests_order FOREIGN KEY (order_id) REFERENCES public.orders(id);

CREATE TABLE public.return_request_items (
    id bigint NOT NULL,
    return_request_id bigint NOT NULL,
    order_item_id bigint NOT NULL,
    quantity bigint NOT NULL,
    refund_amount numeric NOT NULL DEFAULT 0
);
CREATE SEQUENCE public.return_request_items_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.return_request_items_id_seq OWNED BY public.return_request_items.id;
ALTER TABLE ONLY public.return_request_items ALTER COLUMN id SET DEFAULT nextval('public.return_request_items_id_seq'::regclass);
ALTER TABLE ONLY public.return_request_items ADD CONSTRAINT return_request_items_pkey PRIMARY KEY (id);
CREATE INDEX idx_return_request_items_return_request_id ON public.return_request_items USING btree (return_request_id);
CREATE INDEX idx_return_request_items_order_item_id ON public.return_request_items USING btree (order_item_id);
ALTER TABLE ONLY public.return_request_items ADD CONSTRAINT fk_return_requests_items FOREIGN KEY (return_request_id) REFERENCES public.return_requests(id);
ALTER TABLE ONLY public.return_request_items ADD CONSTRAINT fk_return_request_items_order_item FOREIGN KEY (order_item_id) REFERENCES public.order_items(id);

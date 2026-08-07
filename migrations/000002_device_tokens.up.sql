CREATE TABLE public.device_tokens (
    id bigint NOT NULL,
    user_id bigint,
    token text NOT NULL,
    platform text,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE SEQUENCE public.device_tokens_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;
ALTER SEQUENCE public.device_tokens_id_seq OWNED BY public.device_tokens.id;
ALTER TABLE ONLY public.device_tokens ALTER COLUMN id SET DEFAULT nextval('public.device_tokens_id_seq'::regclass);

ALTER TABLE ONLY public.device_tokens ADD CONSTRAINT device_tokens_pkey PRIMARY KEY (id);
CREATE UNIQUE INDEX idx_device_tokens_token ON public.device_tokens USING btree (token);

ALTER TABLE ONLY public.device_tokens ADD CONSTRAINT fk_device_tokens_user FOREIGN KEY (user_id) REFERENCES public.users(id);

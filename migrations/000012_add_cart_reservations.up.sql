-- Short-lived (10-minute) cart holds per (user, product, warehouse), so
-- other shoppers see accurate "in stock" numbers before checkout actually
-- deducts real stock. See internal/services/reservation.go for the logic.
CREATE TABLE public.cart_reservations (
    id BIGSERIAL PRIMARY KEY,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    warehouse_id bigint NOT NULL,
    quantity integer NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);

CREATE UNIQUE INDEX idx_reservation_user_product_wh
    ON public.cart_reservations (user_id, product_id, warehouse_id);

CREATE INDEX idx_reservation_product_wh
    ON public.cart_reservations (product_id, warehouse_id);

CREATE INDEX idx_cart_reservations_expires_at
    ON public.cart_reservations (expires_at);

ALTER TABLE ONLY public.cart_reservations
    ADD CONSTRAINT fk_cart_reservations_user FOREIGN KEY (user_id) REFERENCES public.users(id);
ALTER TABLE ONLY public.cart_reservations
    ADD CONSTRAINT fk_cart_reservations_product FOREIGN KEY (product_id) REFERENCES public.products(id);
ALTER TABLE ONLY public.cart_reservations
    ADD CONSTRAINT fk_cart_reservations_warehouse FOREIGN KEY (warehouse_id) REFERENCES public.warehouses(id);

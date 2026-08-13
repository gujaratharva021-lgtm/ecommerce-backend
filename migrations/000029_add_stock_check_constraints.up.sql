-- Defense-in-depth: the application already prevents negative stock via
-- transactions + row locks + validated deltas, but a DB-level CHECK
-- constraint guarantees it even if a future code path forgets to check,
-- a manual data fix goes wrong, or a bug slips through review.

ALTER TABLE public.inventories
    ADD CONSTRAINT chk_inventories_stock_non_negative
    CHECK (stock >= 0);

ALTER TABLE public.batches
    ADD CONSTRAINT chk_batches_quantity_non_negative
    CHECK (quantity >= 0);

ALTER TABLE public.receivings
    ADD CONSTRAINT chk_receivings_quantities_non_negative
    CHECK (expected_quantity >= 0 AND received_quantity >= 0 AND damaged_quantity >= 0 AND accepted_quantity >= 0);

ALTER TABLE public.receivings
    ADD CONSTRAINT chk_receivings_accepted_le_received
    CHECK (accepted_quantity <= received_quantity);

ALTER TABLE public.stock_transfers
    ADD CONSTRAINT chk_stock_transfers_quantity_positive
    CHECK (quantity > 0);

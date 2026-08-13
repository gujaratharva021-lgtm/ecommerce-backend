ALTER TABLE public.inventories DROP CONSTRAINT IF EXISTS chk_inventories_stock_non_negative;
ALTER TABLE public.batches DROP CONSTRAINT IF EXISTS chk_batches_quantity_non_negative;
ALTER TABLE public.receivings DROP CONSTRAINT IF EXISTS chk_receivings_quantities_non_negative;
ALTER TABLE public.receivings DROP CONSTRAINT IF EXISTS chk_receivings_accepted_le_received;
ALTER TABLE public.stock_transfers DROP CONSTRAINT IF EXISTS chk_stock_transfers_quantity_positive;

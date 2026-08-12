ALTER TABLE public.payments
    DROP COLUMN IF EXISTS gateway,
    DROP COLUMN IF EXISTS refunded_amount;

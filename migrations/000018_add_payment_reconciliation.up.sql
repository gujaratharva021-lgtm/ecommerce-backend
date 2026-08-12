ALTER TABLE public.payments
    ADD COLUMN IF NOT EXISTS gateway text DEFAULT 'razorpay'::text,
    ADD COLUMN IF NOT EXISTS refunded_amount numeric DEFAULT 0;

-- Backfill: existing rows are all online/Razorpay attempts, so gateway is already correct via DEFAULT.

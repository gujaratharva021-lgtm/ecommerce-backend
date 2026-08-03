ALTER TABLE public.orders DROP CONSTRAINT IF EXISTS fk_orders_delivery_partner;
ALTER TABLE public.orders DROP COLUMN IF EXISTS delivery_partner_id;
DROP TABLE IF EXISTS public.delivery_partners CASCADE;

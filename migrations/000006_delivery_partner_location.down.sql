ALTER TABLE public.delivery_partners DROP COLUMN IF EXISTS current_lat;
ALTER TABLE public.delivery_partners DROP COLUMN IF EXISTS current_lng;
ALTER TABLE public.delivery_partners DROP COLUMN IF EXISTS last_location_update;

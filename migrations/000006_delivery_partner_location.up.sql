ALTER TABLE public.delivery_partners ADD COLUMN current_lat double precision;
ALTER TABLE public.delivery_partners ADD COLUMN current_lng double precision;
ALTER TABLE public.delivery_partners ADD COLUMN last_location_update timestamp with time zone;

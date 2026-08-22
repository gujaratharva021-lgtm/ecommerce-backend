ALTER TABLE orders DROP COLUMN IF EXISTS delivery_otp_attempts;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_otp_expires_at;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_otp_hash;

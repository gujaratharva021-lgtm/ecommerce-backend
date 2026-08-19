ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_assignment_expires_at TIMESTAMP NULL;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS delivery_attempted_partner_ids TEXT NOT NULL DEFAULT '''';

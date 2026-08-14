-- Atomic, gap-tolerant sequence for invoice numbers. Replaces the old
-- COUNT(*)+1 scheme, which two concurrent invoice generations for
-- different orders could both read at the same value and then both
-- try to insert the same invoice_number, hitting the unique constraint
-- and failing one of the two orders instead of just proceeding safely.
CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START 1;

ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_amount DOUBLE PRECISION NOT NULL DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_line1 VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_line2 VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_city VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_state VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_pincode VARCHAR(20) NOT NULL DEFAULT '';

ALTER TABLE vendor_bills DROP COLUMN IF EXISTS hold_status;
ALTER TABLE vendor_bills DROP COLUMN IF EXISTS hold_reason;
ALTER TABLE vendor_bills DROP COLUMN IF EXISTS voided_at;
ALTER TABLE vendor_bills DROP COLUMN IF EXISTS void_reason;
ALTER TABLE vendor_bills DROP COLUMN IF EXISTS voided_by_id;
ALTER TABLE bank_transactions DROP COLUMN IF EXISTS voided_at;
ALTER TABLE bank_transactions DROP COLUMN IF EXISTS void_reason;
ALTER TABLE bank_transactions DROP COLUMN IF EXISTS voided_by_id;

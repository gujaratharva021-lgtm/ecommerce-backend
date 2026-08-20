CREATE TABLE IF NOT EXISTS credit_notes (
    id BIGSERIAL PRIMARY KEY,
    credit_note_number VARCHAR(50) UNIQUE NOT NULL,
    invoice_id BIGINT NOT NULL REFERENCES invoices(id),
    order_id BIGINT NOT NULL REFERENCES orders(id),
    return_request_id BIGINT NULL REFERENCES return_requests(id),
    customer_name VARCHAR(255),
    customer_phone VARCHAR(20),
    reason TEXT,
    taxable_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    cgst_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    sgst_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    igst_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    issued_at TIMESTAMPTZ NOT NULL,
    created_by_id BIGINT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_credit_notes_invoice_id ON credit_notes(invoice_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_order_id ON credit_notes(order_id);
CREATE INDEX IF NOT EXISTS idx_credit_notes_return_request_id ON credit_notes(return_request_id);

CREATE TABLE IF NOT EXISTS credit_note_items (
    id BIGSERIAL PRIMARY KEY,
    credit_note_id BIGINT NOT NULL REFERENCES credit_notes(id),
    product_id BIGINT,
    product_name VARCHAR(255),
    quantity INTEGER NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    gst_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    gst_amount DOUBLE PRECISION NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_credit_note_items_credit_note_id ON credit_note_items(credit_note_id);

CREATE TABLE IF NOT EXISTS debit_notes (
    id BIGSERIAL PRIMARY KEY,
    debit_note_number VARCHAR(50) UNIQUE NOT NULL,
    vendor_bill_id BIGINT NOT NULL REFERENCES vendor_bills(id),
    vendor_id BIGINT NOT NULL REFERENCES vendors(id),
    reason TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    gst_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    total_amount DOUBLE PRECISION NOT NULL DEFAULT 0,
    issued_at TIMESTAMPTZ NOT NULL,
    created_by_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_debit_notes_vendor_bill_id ON debit_notes(vendor_bill_id);
CREATE INDEX IF NOT EXISTS idx_debit_notes_vendor_id ON debit_notes(vendor_id);

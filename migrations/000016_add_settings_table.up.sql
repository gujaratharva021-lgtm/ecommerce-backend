CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
    ('free_delivery_threshold', '500'),
    ('flat_delivery_charge', '50'),
    ('min_order_amount', '0'),
    ('cancellation_window_minutes', '10'),
    ('company_name', ''),
    ('support_phone', ''),
    ('support_email', ''),
    ('gst_percentage', '0')
ON CONFLICT (key) DO NOTHING;

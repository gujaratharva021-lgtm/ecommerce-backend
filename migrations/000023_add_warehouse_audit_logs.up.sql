CREATE TABLE IF NOT EXISTS warehouse_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    warehouse_id BIGINT NOT NULL,
    staff_id BIGINT NOT NULL,
    staff_name VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id VARCHAR(50),
    before_value TEXT,
    after_value TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_warehouse_audit_logs_warehouse ON warehouse_audit_logs(warehouse_id);
CREATE INDEX IF NOT EXISTS idx_warehouse_audit_logs_action ON warehouse_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_warehouse_audit_logs_created ON warehouse_audit_logs(created_at);

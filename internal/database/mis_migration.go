package database

func RunMISMigration() {
DB.Exec(`CREATE TABLE IF NOT EXISTS mis_manual_entries (
id BIGSERIAL PRIMARY KEY,
sheet VARCHAR(50) NOT NULL,
week_start DATE NOT NULL,
row_key VARCHAR(255) NOT NULL,
data TEXT NOT NULL DEFAULT '{}',
created_by_id BIGINT NULL REFERENCES users(id),
updated_by_id BIGINT NULL REFERENCES users(id),
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)
DB.Exec(`CREATE INDEX IF NOT EXISTS idx_mis_manual_entries_sheet_week ON mis_manual_entries(sheet, week_start)`)

DB.Exec(`CREATE TABLE IF NOT EXISTS mis_expense_approval (
id BIGSERIAL PRIMARY KEY,
category VARCHAR(100) NOT NULL UNIQUE,
up_to_25k VARCHAR(255),
range_25k_1l VARCHAR(255),
range_1l_5l VARCHAR(255),
above_5l VARCHAR(255),
required_documents TEXT,
approver VARCHAR(255),
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`)

defaults := []string{"Purchase", "Freight", "Warehousing", "Delivery", "Marketing / Promotions", "Wastage / Damage", "Returns", "Other Vendor Expense"}
for _, cat := range defaults {
DB.Exec(`INSERT INTO mis_expense_approval (category) VALUES (?) ON CONFLICT (category) DO NOTHING`, cat)
}
}

package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration019(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
statements := []string{
`CREATE TABLE IF NOT EXISTS picking_tasks (
id BIGSERIAL PRIMARY KEY,
order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id),
warehouse_id BIGINT NOT NULL,
picker_id BIGINT,
status VARCHAR(20) DEFAULT 'pending',
started_at TIMESTAMPTZ,
completed_at TIMESTAMPTZ,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_picking_tasks_warehouse ON picking_tasks(warehouse_id)`,
`CREATE TABLE IF NOT EXISTS picking_task_items (
id BIGSERIAL PRIMARY KEY,
picking_task_id BIGINT NOT NULL REFERENCES picking_tasks(id),
order_item_id BIGINT NOT NULL,
product_id BIGINT NOT NULL REFERENCES products(id),
quantity_needed INT NOT NULL,
quantity_picked INT DEFAULT 0,
status VARCHAR(20) DEFAULT 'pending',
reason VARCHAR(255),
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_picking_task_items_task ON picking_task_items(picking_task_id)`,
`CREATE TABLE IF NOT EXISTS packing_tasks (
id BIGSERIAL PRIMARY KEY,
order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id),
warehouse_id BIGINT NOT NULL,
packer_id BIGINT,
status VARCHAR(20) DEFAULT 'pending',
started_at TIMESTAMPTZ,
completed_at TIMESTAMPTZ,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_packing_tasks_warehouse ON packing_tasks(warehouse_id)`,
}
for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "statement": stmt})
return
}
}
database.DB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (19, false) ON CONFLICT (version) DO UPDATE SET version = 19, dirty = false`)
c.JSON(http.StatusOK, gin.H{"success": true})
}

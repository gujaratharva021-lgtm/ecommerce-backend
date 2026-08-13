package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration020(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
statements := []string{
`CREATE TABLE IF NOT EXISTS order_handovers (
id BIGSERIAL PRIMARY KEY,
order_id BIGINT NOT NULL UNIQUE REFERENCES orders(id),
warehouse_id BIGINT NOT NULL,
warehouse_staff_id BIGINT NOT NULL,
delivery_partner_id BIGINT NOT NULL,
package_count INT NOT NULL DEFAULT 1,
handed_over_at TIMESTAMPTZ NOT NULL,
created_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_order_handovers_warehouse ON order_handovers(warehouse_id)`,
`CREATE INDEX IF NOT EXISTS idx_order_handovers_partner ON order_handovers(delivery_partner_id)`,
}
for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "statement": stmt})
return
}
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration022(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
statements := []string{
`CREATE TABLE IF NOT EXISTS stock_movements (
id BIGSERIAL PRIMARY KEY,
product_id BIGINT NOT NULL,
warehouse_id BIGINT NOT NULL,
previous_qty INT,
change INT,
new_qty INT,
movement_type VARCHAR(20) NOT NULL,
reason VARCHAR(50),
staff_id BIGINT,
reference_id BIGINT,
notes TEXT,
created_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_stock_movements_product ON stock_movements(product_id)`,
`CREATE INDEX IF NOT EXISTS idx_stock_movements_warehouse ON stock_movements(warehouse_id)`,
}
for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "statement": stmt})
return
}
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

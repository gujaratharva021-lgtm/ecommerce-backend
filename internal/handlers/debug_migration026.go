package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration026(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
statements := []string{
`CREATE TABLE IF NOT EXISTS batches (
id BIGSERIAL PRIMARY KEY,
product_id BIGINT NOT NULL REFERENCES products(id),
warehouse_id BIGINT NOT NULL,
batch_number VARCHAR(100) NOT NULL,
manufacture_date TIMESTAMPTZ,
expiry_date TIMESTAMPTZ NOT NULL,
quantity INT NOT NULL,
bin_id BIGINT REFERENCES warehouse_bins(id),
created_by_staff_id BIGINT NOT NULL,
receiving_id BIGINT,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now()
)`,
`CREATE INDEX IF NOT EXISTS idx_batches_product ON batches(product_id)`,
`CREATE INDEX IF NOT EXISTS idx_batches_warehouse ON batches(warehouse_id)`,
`CREATE INDEX IF NOT EXISTS idx_batches_expiry ON batches(expiry_date)`,
}
for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "statement": stmt})
return
}
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

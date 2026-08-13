package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration021(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
statements := []string{
`CREATE TABLE IF NOT EXISTS warehouse_zones (
id BIGSERIAL PRIMARY KEY,
warehouse_id BIGINT NOT NULL,
name VARCHAR(100) NOT NULL,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now(),
UNIQUE(warehouse_id, name)
)`,
`CREATE TABLE IF NOT EXISTS warehouse_racks (
id BIGSERIAL PRIMARY KEY,
zone_id BIGINT NOT NULL REFERENCES warehouse_zones(id),
name VARCHAR(100) NOT NULL,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now(),
UNIQUE(zone_id, name)
)`,
`CREATE TABLE IF NOT EXISTS warehouse_bins (
id BIGSERIAL PRIMARY KEY,
rack_id BIGINT NOT NULL REFERENCES warehouse_racks(id),
name VARCHAR(100) NOT NULL,
created_at TIMESTAMPTZ DEFAULT now(),
updated_at TIMESTAMPTZ DEFAULT now(),
UNIQUE(rack_id, name)
)`,
`ALTER TABLE inventories ADD COLUMN IF NOT EXISTS bin_id BIGINT REFERENCES warehouse_bins(id)`,
`CREATE INDEX IF NOT EXISTS idx_inventories_bin ON inventories(bin_id)`,
}
for _, stmt := range statements {
if err := database.DB.Exec(stmt).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error(), "statement": stmt})
return
}
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

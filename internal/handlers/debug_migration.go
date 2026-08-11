package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunServiceAreaMigration(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
if err := database.DB.Exec(`ALTER TABLE warehouses ADD COLUMN IF NOT EXISTS service_area geometry(Polygon, 4326)`).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
return
}
if err := database.DB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (13, false) ON CONFLICT (version) DO UPDATE SET dirty = false`).Error; err != nil {
c.JSON(http.StatusOK, gin.H{"success": true, "warning": "column added but schema_migrations update failed: " + err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

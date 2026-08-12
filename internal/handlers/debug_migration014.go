package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugRunMigration014(c *gin.Context) {
if c.Query("key") != os.Getenv("DEBUG_MIGRATION_KEY") {
c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
return
}
if err := database.DB.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_blocked BOOLEAN NOT NULL DEFAULT FALSE`).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
return
}
if err := database.DB.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES (14, false) ON CONFLICT (version) DO UPDATE SET version = 14, dirty = false`).Error; err != nil {
c.JSON(http.StatusOK, gin.H{"success": true, "warning": "column added but schema_migrations update failed: " + err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

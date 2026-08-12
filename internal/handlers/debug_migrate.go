package handlers

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// DebugRunMigration is a temporary, key-protected endpoint to apply pending
// schema changes in production without a local DB connection. Remove after use.
func DebugRunMigration(c *gin.Context) {
key := c.Query("key")
expected := os.Getenv("DEBUG_MIGRATION_KEY")
if expected == "" || key != expected {
c.JSON(http.StatusForbidden, gin.H{"error": "invalid or missing key"})
return
}

results := gin.H{}

if err := database.DB.AutoMigrate(&models.Setting{}); err != nil {
results["settings_table"] = "FAILED: " + err.Error()
} else {
results["settings_table"] = "OK"
}

if err := database.DB.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_role VARCHAR(50) NOT NULL DEFAULT ''").Error; err != nil {
results["admin_role_column"] = "FAILED: " + err.Error()
} else {
results["admin_role_column"] = "OK"
}

defaultSettings := map[string]string{
"free_delivery_threshold":     "500",
"flat_delivery_charge":        "50",
"min_order_amount":            "0",
"cancellation_window_minutes": "10",
"company_name":                "",
"support_phone":               "",
"support_email":               "",
"gst_percentage":              "0",
}
seeded := []string{}
for key, value := range defaultSettings {
var existing models.Setting
if err := database.DB.Where("key = ?", key).First(&existing).Error; err != nil {
database.DB.Create(&models.Setting{Key: key, Value: value})
seeded = append(seeded, key)
}
}
results["seeded_settings"] = seeded

c.JSON(http.StatusOK, results)
}

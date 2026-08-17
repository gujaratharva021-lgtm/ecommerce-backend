package database

import "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"

// defaultSettings mirrors the values that used to be hardcoded in
// order_handler.go, so seeding this table doesn't change existing behavior
// until an admin explicitly edits a setting via the admin panel.
var defaultSettings = map[string]string{
"free_delivery_threshold":     "500",
"flat_delivery_charge":        "50",
"platform_fee":                "5",
"min_order_amount":            "0",
"cancellation_window_minutes": "10",
"company_name":                "",
"support_phone":               "",
"support_email":               "",
"gst_percentage":              "0",
}

// seedDefaultSettings inserts any default setting keys that don't already
// exist. Existing values are never overwritten, so admin edits persist
// across deploys/restarts.
func seedDefaultSettings() {
for key, value := range defaultSettings {
var existing models.Setting
if err := DB.Where("key = ?", key).First(&existing).Error; err != nil {
DB.Create(&models.Setting{Key: key, Value: value})
}
}
}

package utils

import (
"strconv"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetSettingFloat reads a numeric business setting from the settings table
// (admin-configurable via GET/PUT /admin/settings). Falls back to
// defaultVal if the key is missing or not parseable as a float.
func GetSettingFloat(key string, defaultVal float64) float64 {
var s models.Setting
if err := database.DB.Where("key = ?", key).First(&s).Error; err != nil {
return defaultVal
}
v, err := strconv.ParseFloat(s.Value, 64)
if err != nil {
return defaultVal
}
return v
}

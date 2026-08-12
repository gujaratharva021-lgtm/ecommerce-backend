package models

import "time"

// Setting is a simple key-value store for admin-configurable business
// settings (delivery fee, min order, cancellation window, company info, etc).
// Values are stored as strings and parsed by the caller to the expected type.
type Setting struct {
Key       string    `gorm:"primaryKey" json:"key"`
Value     string    `json:"value"`
UpdatedAt time.Time `json:"updated_at"`
}

// SettingUpdateRequest is the body for PUT /admin/settings
type SettingUpdateRequest struct {
Settings map[string]string `json:"settings" binding:"required"`
}

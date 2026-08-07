package models

import "time"

// AuditLog records an admin action for accountability/traceability.
// Examples: admin deletes a product, changes an order status, creates a
// coupon. AdminID/AdminPhone identify who did it; Action is a short verb
// phrase; EntityType/EntityID identify what was affected.
type AuditLog struct {
ID         uint      `gorm:"primaryKey" json:"id"`
AdminID    uint      `gorm:"not null;index" json:"admin_id"`
AdminPhone string    `json:"admin_phone"`
Action     string    `gorm:"not null" json:"action"` // e.g. "delete_product", "update_order_status"
EntityType string    `gorm:"index" json:"entity_type"` // e.g. "product", "order", "category"
EntityID   string    `json:"entity_id"`
Details    string    `json:"details"` // free-text summary, e.g. "status: pending -> confirmed"
CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

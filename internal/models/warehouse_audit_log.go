package models

import "time"

// WarehouseAuditLog records warehouse-staff actions for accountability -
// separate from the admin-only AuditLog. Normal warehouse staff can read
// this via GET but there is no PUT/DELETE endpoint, so it cannot be edited.
type WarehouseAuditLog struct {
ID          uint      `gorm:"primaryKey" json:"id"`
WarehouseID uint      `gorm:"not null;index" json:"warehouse_id"`
StaffID     uint      `gorm:"not null;index" json:"staff_id"`
StaffName   string    `json:"staff_name"`
Action      string    `gorm:"not null;index" json:"action"`
EntityType  string    `gorm:"index" json:"entity_type"`
EntityID    string    `json:"entity_id"`
BeforeValue string    `json:"before_value,omitempty"`
AfterValue  string    `json:"after_value,omitempty"`
CreatedAt   time.Time `gorm:"index" json:"created_at"`
}

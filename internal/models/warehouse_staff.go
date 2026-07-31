package models

import (
"time"

"gorm.io/gorm"
)

// WarehouseStaff is a person who works at a specific warehouse — manages
// stock and packs orders there. Logs in via phone + OTP, same flow as
// delivery partners.
type WarehouseStaff struct {
ID          uint           `gorm:"primaryKey" json:"id"`
Name        string         `gorm:"not null" json:"name"`
Phone       string         `gorm:"not null;uniqueIndex" json:"phone"`
WarehouseID uint           `gorm:"not null" json:"warehouse_id"`
Warehouse   Warehouse      `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
IsActive    bool           `gorm:"default:true" json:"is_active"`
CreatedAt   time.Time      `json:"created_at"`
UpdatedAt   time.Time      `json:"updated_at"`
DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// WarehouseStaffRequest is the body for POST/PUT /admin/warehouse-staff
type WarehouseStaffRequest struct {
Name        string `json:"name" binding:"required"`
Phone       string `json:"phone" binding:"required,len=10,numeric"`
WarehouseID uint   `json:"warehouse_id" binding:"required"`
IsActive    *bool  `json:"is_active"`
}
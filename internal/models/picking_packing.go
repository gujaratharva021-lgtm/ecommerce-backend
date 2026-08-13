package models

import "time"

// Picking item statuses
const (
PickItemPending     = "pending"
PickItemPicked      = "picked"
PickItemUnavailable = "unavailable"
PickItemShort       = "short" // partially picked - less than required quantity
)

// PickingTask represents the picking work for one order. Created when the
// warehouse accepts an order and moves it to the picking queue.
type PickingTask struct {
ID          uint       `gorm:"primaryKey" json:"id"`
OrderID     uint       `gorm:"not null;uniqueIndex" json:"order_id"`
Order       Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
WarehouseID uint       `gorm:"not null;index" json:"warehouse_id"`
PickerID    *uint      `gorm:"index" json:"picker_id,omitempty"` // WarehouseStaff.ID
Status      string     `gorm:"default:pending" json:"status"`   // pending/in_progress/completed
StartedAt   *time.Time `json:"started_at,omitempty"`
CompletedAt *time.Time `json:"completed_at,omitempty"`
CreatedAt   time.Time  `json:"created_at"`
UpdatedAt   time.Time  `json:"updated_at"`
Items       []PickingTaskItem `gorm:"foreignKey:PickingTaskID" json:"items,omitempty"`
}

// PickingTaskItem tracks the picking outcome for one order item (one product line).
type PickingTaskItem struct {
ID              uint      `gorm:"primaryKey" json:"id"`
PickingTaskID   uint      `gorm:"not null;index" json:"picking_task_id"`
OrderItemID     uint      `gorm:"not null" json:"order_item_id"`
ProductID       uint      `gorm:"not null" json:"product_id"`
Product         Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
QuantityNeeded  int       `gorm:"not null" json:"quantity_needed"`
QuantityPicked  int       `gorm:"default:0" json:"quantity_picked"`
Status          string    `gorm:"default:pending" json:"status"` // pending/picked/unavailable/short
Reason          string    `json:"reason,omitempty"`              // set when unavailable/short
CreatedAt       time.Time `json:"created_at"`
UpdatedAt       time.Time `json:"updated_at"`
}

// PackingTask represents the packing work for one order. Created when
// picking is completed.
type PackingTask struct {
ID          uint       `gorm:"primaryKey" json:"id"`
OrderID     uint       `gorm:"not null;uniqueIndex" json:"order_id"`
Order       Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
WarehouseID uint       `gorm:"not null;index" json:"warehouse_id"`
PackerID    *uint      `gorm:"index" json:"packer_id,omitempty"` // WarehouseStaff.ID
Status      string     `gorm:"default:pending" json:"status"`    // pending/in_progress/completed
StartedAt   *time.Time `json:"started_at,omitempty"`
CompletedAt *time.Time `json:"completed_at,omitempty"`
CreatedAt   time.Time  `json:"created_at"`
UpdatedAt   time.Time  `json:"updated_at"`
}

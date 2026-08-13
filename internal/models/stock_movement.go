package models

import "time"

// Stock movement types
const (
MovementReceive    = "receive"
MovementSale        = "sale"
MovementAdjustment  = "adjustment"
MovementTransfer    = "transfer"
MovementDamaged     = "damaged"
MovementExpired     = "expired"
MovementReturn      = "return"
MovementCorrection  = "correction"
)

// Stock adjustment reasons
const (
AdjustReasonDamaged        = "damaged"
AdjustReasonExpired        = "expired"
AdjustReasonCountingError  = "counting_error"
AdjustReasonLost           = "lost"
AdjustReasonFound          = "found"
AdjustReasonManualCorrection = "manual_correction"
AdjustReasonOther          = "other"
)

// StockMovement is an immutable audit log entry for every inventory quantity
// change at a warehouse - receiving, sales/order allocation, adjustments,
// transfers, damage, expiry, returns, corrections. Nothing should change
// Inventory.Stock without also writing one of these.
type StockMovement struct {
ID           uint      `gorm:"primaryKey" json:"id"`
ProductID    uint      `gorm:"not null;index" json:"product_id"`
Product      Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
WarehouseID  uint      `gorm:"not null;index" json:"warehouse_id"`
PreviousQty  int       `json:"previous_qty"`
Change       int       `json:"change"` // signed - positive = increase, negative = decrease
NewQty       int       `json:"new_qty"`
MovementType string    `gorm:"not null" json:"movement_type"`
Reason       string    `json:"reason,omitempty"`
StaffID      *uint     `json:"staff_id,omitempty"` // WarehouseStaff.ID, null for system-generated (order allocation)
ReferenceID  *uint     `json:"reference_id,omitempty"` // e.g. order_id for sale, transfer_id for transfer
Notes        string    `json:"notes,omitempty"`
CreatedAt    time.Time `json:"created_at"`
}

// StockAdjustmentRequest is the body for POST /warehouse/inventory/:product_id/adjust
type StockAdjustmentRequest struct {
NewQuantity int    `json:"new_quantity" binding:"gte=0"`
Reason      string `json:"reason" binding:"required,oneof=damaged expired counting_error lost found manual_correction other"`
Notes       string `json:"notes"`
}

package models

import "time"

const (
StockTransferPending   = "pending"
StockTransferInTransit = "in_transit"
StockTransferReceived  = "received"
StockTransferRejected  = "rejected"
StockTransferCancelled = "cancelled"
)

// StockTransfer moves stock from one warehouse to another. A warehouse
// staff member requests it; an admin approves (deducting from the source
// immediately, marking it in_transit) or rejects it. Once the destination
// warehouse staff confirms receipt, the stock is added there and the
// transfer is marked received.
type StockTransfer struct {
ID              uint      `gorm:"primaryKey" json:"id"`
ProductID       uint      `gorm:"not null;index" json:"product_id"`
Product         Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
FromWarehouseID uint      `gorm:"not null;index" json:"from_warehouse_id"`
FromWarehouse   Warehouse `gorm:"foreignKey:FromWarehouseID" json:"from_warehouse,omitempty"`
ToWarehouseID   uint      `gorm:"not null;index" json:"to_warehouse_id"`
ToWarehouse     Warehouse `gorm:"foreignKey:ToWarehouseID" json:"to_warehouse,omitempty"`
Quantity        int       `gorm:"not null" json:"quantity"`
Status          string    `gorm:"not null;default:pending;index" json:"status"`
RequestedBy     uint      `gorm:"not null" json:"requested_by"`
ApprovedBy      *uint     `json:"approved_by,omitempty"`
CreatedAt       time.Time `json:"created_at"`
UpdatedAt       time.Time `json:"updated_at"`
}

// StockTransferRequest is the body for POST /warehouse/stock-transfers (warehouse staff only)
type StockTransferRequest struct {
ProductID     uint `json:"product_id" binding:"required"`
ToWarehouseID uint `json:"to_warehouse_id" binding:"required"`
Quantity      int  `json:"quantity" binding:"required,gt=0"`
}

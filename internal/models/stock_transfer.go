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
// BatchID optionally pins the transfer to a specific source batch -
// otherwise the earliest-expiry batch (FEFO) is picked automatically
// at approve time, if the product is batch-tracked at all.
BatchID         *uint     `json:"batch_id,omitempty"`
// DamagedQuantity is set at receive time for stock lost/damaged in
// transit. Partial transfers are not otherwise supported - the full
// approved Quantity must arrive; any shortfall must be accounted for
// explicitly here, not silently dropped.
DamagedQuantity int       `gorm:"not null;default:0" json:"damaged_quantity"`
// ReceivedBinID is the destination bin the receiving staff assigns
// the stock to, mirroring PutAwayReceivingRequest.BinID.
ReceivedBinID   *uint     `json:"received_bin_id,omitempty"`
Status          string    `gorm:"not null;default:pending;index" json:"status"`
RequestedBy     uint      `gorm:"not null" json:"requested_by"`
ApprovedBy      *uint     `json:"approved_by,omitempty"`
CreatedAt       time.Time `json:"created_at"`
UpdatedAt       time.Time `json:"updated_at"`
}

// StockTransferRequest is the body for POST /warehouse/stock-transfers (warehouse staff only)
type StockTransferRequest struct {
ProductID     uint  `json:"product_id" binding:"required"`
ToWarehouseID uint  `json:"to_warehouse_id" binding:"required"`
Quantity      int   `json:"quantity" binding:"required,gt=0"`
// BatchID optionally pins the transfer to a specific source batch.
// If omitted and the product is batch-tracked, the earliest-expiry
// batch (FEFO) is picked automatically at approve time.
BatchID       *uint `json:"batch_id"`
}

// ReceiveStockTransferRequest is the body for PUT /warehouse/stock-transfers/:id/receive.
// Both fields are optional: BinID assigns the destination bin (mirrors
// PutAwayReceivingRequest); DamagedQuantity accounts for stock lost or
// damaged in transit. Partial transfers are otherwise not supported -
// the full approved Quantity must be receivable; any shortfall must be
// declared here explicitly, not silently dropped.
type ReceiveStockTransferRequest struct {
BinID           *uint `json:"bin_id"`
DamagedQuantity int   `json:"damaged_quantity" binding:"gte=0"`
}

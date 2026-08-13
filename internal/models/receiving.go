package models

import "time"

// Receiving statuses
const (
ReceivingStatusPending  = "pending"
ReceivingStatusReceived = "received"
ReceivingStatusAccepted = "accepted"
ReceivingStatusRejected = "rejected"
ReceivingStatusPutAway  = "put_away"
)

// Receiving is a single inbound stock record for one product at one
// warehouse. SupplierName/ReferenceNumber are free-text so this can be
// used standalone today and later linked to a real PurchaseOrder model
// without changing this shape.
type Receiving struct {
ID                uint       `gorm:"primaryKey" json:"id"`
WarehouseID       uint       `gorm:"not null;index" json:"warehouse_id"`
SupplierName      string     `gorm:"not null" json:"supplier_name"`
ReferenceNumber   string     `json:"reference_number"`
ProductID         uint       `gorm:"not null;index" json:"product_id"`
Product           Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
ExpectedQuantity  int        `gorm:"not null" json:"expected_quantity"`
ReceivedQuantity  int        `json:"received_quantity"`
DamagedQuantity   int        `json:"damaged_quantity"`
AcceptedQuantity  int        `json:"accepted_quantity"`
Status            string     `gorm:"default:pending;index" json:"status"`
BinID             *uint      `json:"bin_id,omitempty"`
Bin               *WarehouseBin `gorm:"foreignKey:BinID" json:"bin,omitempty"`
CreatedByStaffID  uint       `gorm:"not null" json:"created_by_staff_id"`
ReceivedByStaffID *uint      `json:"received_by_staff_id,omitempty"`
QCByStaffID       *uint      `json:"qc_by_staff_id,omitempty"`
PutAwayByStaffID  *uint      `json:"put_away_by_staff_id,omitempty"`
Notes             string     `json:"notes,omitempty"`
RejectionReason   string     `json:"rejection_reason,omitempty"`
ReceivedAt        *time.Time `json:"received_at,omitempty"`
QCAt              *time.Time `json:"qc_at,omitempty"`
PutAwayAt         *time.Time `json:"put_away_at,omitempty"`
CreatedAt         time.Time  `json:"created_at"`
UpdatedAt         time.Time  `json:"updated_at"`
}

// CreateReceivingRequest is the body for POST /warehouse/receiving
type CreateReceivingRequest struct {
SupplierName     string `json:"supplier_name" binding:"required"`
ReferenceNumber  string `json:"reference_number"`
ProductID        uint   `json:"product_id" binding:"required"`
ExpectedQuantity int    `json:"expected_quantity" binding:"required,gt=0"`
Notes            string `json:"notes"`
}

// MarkReceivedRequest is the body for PUT /warehouse/receiving/:id/receive
type MarkReceivedRequest struct {
ReceivedQuantity int    `json:"received_quantity" binding:"gte=0"`
DamagedQuantity  int    `json:"damaged_quantity" binding:"gte=0"`
Notes            string `json:"notes"`
}

// QCReceivingRequest is the body for PUT /warehouse/receiving/:id/qc
type QCReceivingRequest struct {
Action           string `json:"action" binding:"required,oneof=accept reject"`
AcceptedQuantity int    `json:"accepted_quantity" binding:"gte=0"`
RejectionReason  string `json:"rejection_reason"`
}

// PutAwayReceivingRequest is the body for PUT /warehouse/receiving/:id/putaway
type PutAwayReceivingRequest struct {
BinID *uint `json:"bin_id"`
}

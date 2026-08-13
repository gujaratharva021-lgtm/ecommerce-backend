package models

import "time"

// Exception types
const (
ExceptionUnavailable         = "unavailable"
ExceptionShortQuantity       = "short_quantity"
ExceptionDamaged             = "damaged"
ExceptionWrongProduct        = "wrong_product"
ExceptionBarcodeMismatch     = "barcode_mismatch"
ExceptionPickingFailure      = "picking_failure"
ExceptionPackingFailure      = "packing_failure"
ExceptionOrderCancellation   = "order_cancellation"
ExceptionDeliveryUnavailable = "delivery_partner_unavailable"
ExceptionOrderDelayed        = "order_delayed"
)

// Exception statuses
const (
ExceptionStatusOpen          = "open"
ExceptionStatusInvestigating = "investigating"
ExceptionStatusResolved      = "resolved"
ExceptionStatusClosed        = "closed"
)

// Exception priorities
const (
ExceptionPriorityLow    = "low"
ExceptionPriorityMedium = "medium"
ExceptionPriorityHigh   = "high"
)

// WarehouseException is a dedicated record for anything that went wrong
// during fulfillment - unavailable/short/damaged items, barcode mismatches,
// picking/packing failures, etc. Created automatically where possible (e.g.
// when a picker marks an item unavailable) so staff don't have to
// double-enter what they already reported inline.
type WarehouseException struct {
ID           uint       `gorm:"primaryKey" json:"id"`
OrderID      uint       `gorm:"not null;index" json:"order_id"`
Order        Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
ProductID    *uint      `json:"product_id,omitempty"`
Product      *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
WarehouseID  uint       `gorm:"not null;index" json:"warehouse_id"`
Type         string     `gorm:"not null" json:"type"`
Reason       string     `json:"reason"`
Priority     string     `gorm:"default:medium" json:"priority"`
StaffID      *uint      `json:"staff_id,omitempty"`
Status       string     `gorm:"default:open;index" json:"status"`
Resolution   string     `json:"resolution,omitempty"`
ResolvedByID *uint      `json:"resolved_by_id,omitempty"`
ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
CreatedAt    time.Time  `json:"created_at"`
UpdatedAt    time.Time  `json:"updated_at"`
}

// UpdateExceptionRequest is the body for PUT /warehouse/exceptions/:id
type UpdateExceptionRequest struct {
Status     string `json:"status" binding:"required,oneof=open investigating resolved closed"`
Resolution string `json:"resolution"`
}

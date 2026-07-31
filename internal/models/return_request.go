package models

import "time"

// Return request status
const (
	ReturnStatusPending  = "pending"
	ReturnStatusApproved = "approved"
	ReturnStatusRejected = "rejected"
)

// ReturnRequest is a customer-initiated request to return a delivered
// order. Approval triggers a stock restore and a wallet refund for the
// order's total amount — no partial/item-level returns yet, the whole
// order goes back.
type ReturnRequest struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OrderID      uint      `gorm:"not null;index" json:"order_id"`
	Order        Order     `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	Reason       string    `gorm:"not null" json:"reason"`
	Status       string    `gorm:"default:pending" json:"status"`
	RefundAmount float64   `gorm:"not null;default:0" json:"refund_amount"`
	ProcessedBy  *uint     `json:"processed_by,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReturnRequestBody is the body for POST /orders/:id/return
type ReturnRequestBody struct {
	Reason string `json:"reason" binding:"required"`
}

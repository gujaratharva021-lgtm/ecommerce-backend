package models

import "time"

// Payment lifecycle for a single Razorpay order attempt.
const (
	PaymentStatusCreated = "created"
	PaymentStatusPaid    = "paid"
	PaymentStatusFailed  = "failed"
)

// Payment tracks the Razorpay side of an order's online payment.
// One row per Order (re-creating a payment order overwrites the same row,
// so a customer can retry after a failed attempt without piling up rows).
type Payment struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	OrderID           uint      `gorm:"uniqueIndex;not null" json:"order_id"`
	Order             Order     `gorm:"foreignKey:OrderID" json:"-"`
	RazorpayOrderID   string    `json:"razorpay_order_id"`
	RazorpayPaymentID string    `json:"razorpay_payment_id,omitempty"`
	RazorpaySignature string    `json:"-"`
	Amount            float64   `json:"amount"`
	Currency          string    `gorm:"default:INR" json:"currency"`
	Status            string    `gorm:"default:created" json:"status"` // created/paid/failed
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CreatePaymentOrderResponse is returned by POST /orders/:id/payment so the
// frontend (web/Flutter Razorpay SDK) has everything needed to open Checkout.
type CreatePaymentOrderResponse struct {
	RazorpayOrderID string `json:"razorpay_order_id"`
	Amount          int64  `json:"amount"` // in paise, as Razorpay Checkout expects
	Currency        string `json:"currency"`
	KeyID           string `json:"key_id"`
	OrderID         uint   `json:"order_id"`
}

// VerifyPaymentRequest is the body for POST /orders/:id/payment/verify —
// the three fields Razorpay Checkout hands back to the frontend after a
// successful payment, forwarded here for signature verification.
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string `json:"razorpay_signature" binding:"required"`
}

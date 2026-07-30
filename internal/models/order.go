package models

import "time"

// Valid order status transitions: pending -> confirmed -> shipped -> delivered
// A pending or confirmed order can also move to cancelled.
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Payment method chosen at checkout, and the resulting payment lifecycle.
const (
	PaymentMethodCOD    = "cod"
	PaymentMethodOnline = "online"

	OrderPaymentStatusPending = "pending"
	OrderPaymentStatusPaid    = "paid"
	OrderPaymentStatusFailed  = "failed"
)

type Order struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	UserID         uint        `gorm:"not null;index" json:"user_id"`
	User           User        `gorm:"foreignKey:UserID" json:"-"`
	AddressID      uint        `gorm:"not null" json:"address_id"`
	Address        Address     `gorm:"foreignKey:AddressID" json:"address,omitempty"`
	ItemsAmount    float64     `gorm:"not null" json:"items_amount"`
	DeliveryCharge float64     `gorm:"not null;default:0" json:"delivery_charge"`
	TotalAmount    float64     `gorm:"not null" json:"total_amount"`
	Status         string      `gorm:"default:pending" json:"status"`         // pending/confirmed/shipped/delivered/cancelled
	PaymentMethod  string      `gorm:"default:cod" json:"payment_method"`     // cod/online
	PaymentStatus  string      `gorm:"default:pending" json:"payment_status"` // pending/paid/failed
	Items          []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Price     float64   `gorm:"not null" json:"price"` // price at time of order
	CreatedAt time.Time `json:"created_at"`
}

// CheckoutRequest is the body for POST /orders/checkout.
// AddressID is optional — if omitted, the user's default address is used.
// PaymentMethod is optional — defaults to "cod" if omitted; "online" starts
// the Razorpay flow (see POST /orders/:id/payment).
type CheckoutRequest struct {
	AddressID     uint   `json:"address_id"`
	PaymentMethod string `json:"payment_method" binding:"omitempty,oneof=cod online"`
	CouponCode    string `json:"coupon_code"`
}

// OrderStatusUpdateRequest is the body for PUT /admin/orders/:id/status (admin only).
type OrderStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=confirmed shipped delivered cancelled"`
}

// OrderListResponse wraps paginated order results.
type OrderListResponse struct {
	Orders     []Order `json:"orders"`
	Page       int     `json:"page"`
	Limit      int     `json:"limit"`
	Total      int64   `json:"total"`
	TotalPages int     `json:"total_pages"`
}

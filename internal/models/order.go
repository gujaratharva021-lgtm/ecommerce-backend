package models

import "time"

// Valid order status transitions: pending -> confirmed -> shipped -> delivered
// A pending or confirmed order can also move to cancelled.
const (
    OrderStatusPending         = "pending"
    OrderStatusConfirmed       = "confirmed"
    OrderStatusPicking         = "picking"
    OrderStatusPicked          = "picked"
    OrderStatusPacking         = "packing"
    OrderStatusPacked          = "packed"
    OrderStatusReadyForDispatch = "ready_for_dispatch"
    OrderStatusHandedOver      = "handed_over"
    OrderStatusShipped         = "shipped"
    OrderStatusDelivered       = "delivered"
    OrderStatusReturned        = "returned"
    OrderStatusCancelled       = "cancelled"
)

// Payment method chosen at checkout, and the resulting payment lifecycle.
const (
    PaymentMethodCOD    = "cod"
    PaymentMethodOnline = "online"

    OrderPaymentStatusPending = "pending"
    OrderPaymentStatusPaid    = "paid"
    OrderPaymentStatusFailed  = "failed"
)

// Delivery assignment lifecycle. This tracks the state of the *courier's*
// response to being handed an order, independent of the order's own
// fulfillment Status above. It only ever applies while DeliveryPartnerID
// is set:
//
//    ASSIGNED - a partner has just been offered the delivery (admin action,
//               auto-assign, or automatic reassignment) and hasn't responded
//               yet. Only the ASSIGNED state is open to Accept/Reject, and it
//               carries an acceptance deadline in DeliveryAssignmentExpiresAt.
//    ACCEPTED - the assigned partner accepted the delivery. Terminal (happy
//               path) - no further automatic reassignment.
//    REJECTED - the assigned partner declined. Automatically triggers an
//               attempt to offer the order to the next eligible partner who
//               hasn't already been tried (see services.TryAssignNextPartner).
//    EXPIRED  - the partner never responded within the acceptance window.
//               Set by the periodic expiry sweep (services.ExpireStaleAssignments),
//               which also triggers the same automatic reassignment as REJECTED.
const (
    DeliveryAssignmentStatusAssigned = "assigned"
    DeliveryAssignmentStatusAccepted = "accepted"
    DeliveryAssignmentStatusRejected = "rejected"
    DeliveryAssignmentStatusExpired  = "expired"
)

type Order struct {
    ID                uint             `gorm:"primaryKey" json:"id"`
    UserID            uint             `gorm:"not null;index" json:"user_id"`
    User              User             `gorm:"foreignKey:UserID" json:"-"`
    AddressID         uint             `gorm:"not null" json:"address_id"`
    Address           Address          `gorm:"foreignKey:AddressID" json:"address,omitempty"`
    WarehouseID       *uint            `gorm:"index;index:idx_orders_warehouse_status,priority:1" json:"warehouse_id,omitempty"`
    Warehouse         *Warehouse       `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
    ItemsAmount       float64          `gorm:"not null" json:"items_amount"`
    DeliveryCharge    float64          `gorm:"not null;default:0" json:"delivery_charge"`
    PlatformFee       float64          `gorm:"not null;default:0" json:"platform_fee"`
    WalletAmountUsed  float64          `gorm:"not null;default:0" json:"wallet_amount_used"`
    TotalAmount       float64          `gorm:"not null" json:"total_amount"`
    Status            string           `gorm:"default:pending;index:idx_orders_warehouse_status,priority:2" json:"status"`         // pending/confirmed/shipped/delivered/cancelled
    PaymentMethod     string           `gorm:"default:cod" json:"payment_method"`     // cod/online
    PaymentStatus     string           `gorm:"default:pending" json:"payment_status"` // pending/paid/failed
    DeliveryPartnerID *uint            `gorm:"index" json:"delivery_partner_id,omitempty"`
    DeliveryPartner   *DeliveryPartner `gorm:"foreignKey:DeliveryPartnerID" json:"delivery_partner,omitempty"`
    // DeliveryAssignmentStatus is one of the DeliveryAssignmentStatus*
    // constants above, or empty/nil when no partner has ever been assigned.
    DeliveryAssignmentStatus *string `gorm:"index;size:20" json:"delivery_assignment_status,omitempty"`
    DeliveryRejectionReason  *string `json:"delivery_rejection_reason,omitempty"`
    // DeliveryAssignmentExpiresAt is when the current ASSIGNED offer expires
    // if the partner doesn't respond in time. Nil when there's no pending
    // offer (before first assignment, or after accept/reject/expiry).
    DeliveryAssignmentExpiresAt *time.Time `gorm:"index" json:"delivery_assignment_expires_at,omitempty"`
    // DeliveryAttemptedPartnerIDs is a comma-separated list of every
    // partner ID already offered this order (assigned, then
    // rejected/expired), so automatic reassignment never offers the same
    // order to the same partner twice.
    DeliveryAttemptedPartnerIDs string       `gorm:"default:''" json:"-"`
    Items                       []OrderItem  `gorm:"foreignKey:OrderID" json:"items,omitempty"`
    CreatedAt                   time.Time    `json:"created_at"`
    UpdatedAt                   time.Time    `json:"updated_at"`
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
// AddressID is optional - if omitted, the user's default address is used.
// PaymentMethod is optional - defaults to "cod" if omitted; "online" starts
// the Razorpay flow (see POST /orders/:id/payment).
type CheckoutRequest struct {
    AddressID     uint   `json:"address_id"`
    PaymentMethod string `json:"payment_method" binding:"omitempty,oneof=cod online"`
    CouponCode    string `json:"coupon_code"`
    UseWallet     bool   `json:"use_wallet"`
}

// OrderStatusUpdateRequest is the body for PUT /admin/orders/:id/status (admin only).
type OrderStatusUpdateRequest struct {
    Status string `json:"status" binding:"required,oneof=confirmed shipped delivered cancelled"`
}

// RejectAssignmentRequest is the body for PUT /delivery/orders/:id/reject
// (delivery partner only). Reason is optional.
type RejectAssignmentRequest struct {
    Reason string `json:"reason"`
}

// OrderListResponse wraps paginated order results.
type OrderListResponse struct {
    Orders     []Order `json:"orders"`
    Page       int     `json:"page"`
    Limit      int     `json:"limit"`
    Total      int64   `json:"total"`
    TotalPages int     `json:"total_pages"`
}

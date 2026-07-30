package models

import "time"

// Coupon discount types.
const (
	CouponTypeFlat       = "flat"
	CouponTypePercentage = "percentage"
)

// Coupon is a discount code that can be applied at checkout. UsedCount is
// incremented every time it's successfully applied to an order; once it
// reaches UsageLimit the coupon is rejected even if still active/unexpired.
type Coupon struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Code              string    `gorm:"uniqueIndex;not null" json:"code"`
	DiscountType      string    `json:"discount_type"` // "flat" or "percentage"
	DiscountValue     float64   `json:"discount_value"`
	MinOrderAmount    float64   `gorm:"default:0" json:"min_order_amount"`
	MaxDiscountAmount *float64  `json:"max_discount_amount,omitempty"` // optional cap, mainly for percentage coupons
	UsageLimit        int       `gorm:"default:1" json:"usage_limit"`
	UsedCount         int       `gorm:"default:0" json:"used_count"`
	ExpiryDate        time.Time `json:"expiry_date"`
	IsActive          bool      `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// OrderCoupon records which coupon was applied to a given order and how much
// discount it gave. One row per order — an order can only use one coupon.
type OrderCoupon struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrderID        uint      `gorm:"uniqueIndex;not null" json:"order_id"`
	Order          Order     `gorm:"foreignKey:OrderID" json:"-"`
	CouponID       uint      `gorm:"not null" json:"coupon_id"`
	Coupon         Coupon    `gorm:"foreignKey:CouponID" json:"coupon,omitempty"`
	DiscountAmount float64   `json:"discount_amount"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateCouponRequest is the admin request body for POST /admin/coupons.
type CreateCouponRequest struct {
	Code              string   `json:"code" binding:"required"`
	DiscountType      string   `json:"discount_type" binding:"required,oneof=flat percentage"`
	DiscountValue     float64  `json:"discount_value" binding:"required,gt=0"`
	MinOrderAmount    float64  `json:"min_order_amount"`
	MaxDiscountAmount *float64 `json:"max_discount_amount"`
	UsageLimit        int      `json:"usage_limit"`
	ExpiryDate        string   `json:"expiry_date" binding:"required"` // format: "2006-01-02"
}

// ValidateCouponRequest is the body for POST /coupons/validate, used by the
// checkout screen to preview a discount before placing the order.
type ValidateCouponRequest struct {
	Code        string  `json:"code" binding:"required"`
	OrderAmount float64 `json:"order_amount" binding:"required,gt=0"`
}

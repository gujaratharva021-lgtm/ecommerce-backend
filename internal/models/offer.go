package models

import "time"

// Offer represents a promotional campaign shown on the storefront (e.g. a
// homepage banner or category promo), independent of coupon codes. Unlike
// Coupon, an offer doesn't require the customer to enter a code - it's
// informational/display-only unless tied to a discount that's applied
// automatically elsewhere.
type Offer struct {
ID           uint      `gorm:"primaryKey" json:"id"`
Title        string    `gorm:"not null" json:"title"`
Description  string    `json:"description"`
ImageURL     string    `json:"image_url"`
DiscountText string    `json:"discount_text"` // free-text display, e.g. "Up to 30% off"
CategoryID   *uint     `json:"category_id,omitempty"`
StartDate    time.Time `json:"start_date"`
EndDate      time.Time `json:"end_date"`
IsActive     bool      `gorm:"default:true" json:"is_active"`
CreatedAt    time.Time `json:"created_at"`
UpdatedAt    time.Time `json:"updated_at"`
}

// CreateOfferRequest is the admin request body for POST /admin/offers.
type CreateOfferRequest struct {
Title        string `json:"title" binding:"required"`
Description  string `json:"description"`
ImageURL     string `json:"image_url"`
DiscountText string `json:"discount_text"`
CategoryID   *uint  `json:"category_id"`
StartDate    string `json:"start_date" binding:"required"` // "2006-01-02"
EndDate      string `json:"end_date" binding:"required"`
}

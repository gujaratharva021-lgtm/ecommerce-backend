package models

import "time"

// Banner represents a homepage carousel/hero image shown on the storefront.
// LinkType/LinkValue let admin point a banner to a product, category, or an
// arbitrary external/internal URL - frontend decides how to route based on
// LinkType.
type Banner struct {
ID          uint      `gorm:"primaryKey" json:"id"`
ImageURL    string    `gorm:"not null" json:"image_url"`
Title       string    `json:"title"`
LinkType    string    `json:"link_type"` // "product" | "category" | "url" | "none"
LinkValue   string    `json:"link_value"` // product id / category id / raw url, per LinkType
DisplayOrder int      `gorm:"default:0" json:"display_order"`
IsActive    bool      `gorm:"default:true" json:"is_active"`
CreatedAt   time.Time `json:"created_at"`
UpdatedAt   time.Time `json:"updated_at"`
}

// CreateBannerRequest is the admin request body for POST /admin/banners.
type CreateBannerRequest struct {
ImageURL     string `json:"image_url" binding:"required"`
Title        string `json:"title"`
LinkType     string `json:"link_type"`
LinkValue    string `json:"link_value"`
DisplayOrder int    `json:"display_order"`
}

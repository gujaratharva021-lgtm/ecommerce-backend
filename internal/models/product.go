package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"not null;index" json:"name"`
	Description string     `json:"description"`
	Price       float64    `gorm:"not null;index" json:"price"`
// CostPrice is what we paid to acquire the product (per unit). Used for
// COGS / gross profit calculations in the Finance panel. 0 means not yet
// set - callers should treat 0 as "unknown cost", not "free product".
CostPrice   float64    `gorm:"not null;default:0" json:"cost_price"`
	GSTPercent  float64    `gorm:"not null;default:0" json:"gst_percent"`
	HSNCode     string     `json:"hsn_code,omitempty"`
	ImageURL    string     `json:"image_url"`
Barcode     string     `gorm:"index" json:"barcode,omitempty"`
	CategoryID  uint       `gorm:"index" json:"category_id"`
	Category    Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Inventories []Inventory `gorm:"foreignKey:ProductID" json:"inventories,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProductRequest is the body for POST/PUT /admin/products (admin only).
// Stock is only used on create, to seed the product's Inventory row ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬ÃƒÂ¢Ã¢â€šÂ¬Ã‚Â
// use PUT /admin/products/:id/inventory to adjust stock afterwards.
type ProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	GSTPercent  float64 `json:"gst_percent" binding:"gte=0,lte=100"`
CostPrice   float64 `json:"cost_price" binding:"gte=0"`
	HSNCode     string  `json:"hsn_code"`
	ImageURL    string  `json:"image_url"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	Stock       int     `json:"stock" binding:"gte=0"`
}

// ProductListQuery binds query params for GET /products (filter, sort, search, paginate).
type ProductListQuery struct {
	Search     string  `form:"search"`
	CategoryID uint    `form:"category_id"`
	MinPrice   float64 `form:"min_price"`
	MaxPrice   float64 `form:"max_price"`
	InStock    *bool   `form:"in_stock"`
	Sort       string  `form:"sort"` // price_asc, price_desc, name_asc, name_desc, newest
	Page       int     `form:"page,default=1"`
	Limit      int     `form:"limit,default=20"`
}

// ProductListResponse wraps paginated product results.
type ProductListResponse struct {
	Products   []Product `json:"products"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	Total      int64     `json:"total"`
	TotalPages int       `json:"total_pages"`
}




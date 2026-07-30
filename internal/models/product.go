package models

import "time"

type Product struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"not null;index" json:"name"`
	Description string     `json:"description"`
	Price       float64    `gorm:"not null;index" json:"price"`
	ImageURL    string     `json:"image_url"`
	CategoryID  uint       `gorm:"index" json:"category_id"`
	Category    Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Inventory   *Inventory `gorm:"foreignKey:ProductID" json:"inventory,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProductRequest is the body for POST/PUT /admin/products (admin only).
// Stock is only used on create, to seed the product's Inventory row —
// use PUT /admin/products/:id/inventory to adjust stock afterwards.
type ProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
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

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

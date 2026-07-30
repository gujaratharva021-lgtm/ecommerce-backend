package models

import "time"

type Inventory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"uniqueIndex;not null" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Stock     int       `gorm:"default:0" json:"stock"`
	InStock   bool      `gorm:"default:true" json:"in_stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InventoryUpdateRequest is the body for PUT /admin/products/:id/inventory (admin only).
// Stock is an absolute value (not a delta) — it replaces the current stock count.
type InventoryUpdateRequest struct {
	Stock int `json:"stock" binding:"required,gte=0"`
}

package models

import "time"

// Batch tracks a specific manufacturing lot of a product at a warehouse,
// for products where batch/expiry tracking makes business sense
// (perishables, food, medicine). Products without expiry concerns simply
// never get Batch rows and continue to use Inventory.Stock as-is.
type Batch struct {
ID             uint       `gorm:"primaryKey" json:"id"`
ProductID      uint       `gorm:"not null;index" json:"product_id"`
Product        Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
WarehouseID    uint       `gorm:"not null;index" json:"warehouse_id"`
BatchNumber    string     `gorm:"not null" json:"batch_number"`
ManufactureDate *time.Time `json:"manufacture_date,omitempty"`
ExpiryDate     time.Time  `gorm:"not null;index" json:"expiry_date"`
Quantity       int        `gorm:"not null" json:"quantity"`
BinID          *uint      `json:"bin_id,omitempty"`
Bin            *WarehouseBin `gorm:"foreignKey:BinID" json:"bin,omitempty"`
CreatedByStaffID uint     `gorm:"not null" json:"created_by_staff_id"`
ReceivingID    *uint      `json:"receiving_id,omitempty"`
CreatedAt      time.Time  `json:"created_at"`
UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateBatchRequest is the body for POST /warehouse/batches
type CreateBatchRequest struct {
ProductID       uint       `json:"product_id" binding:"required"`
BatchNumber     string     `json:"batch_number" binding:"required"`
ManufactureDate *time.Time `json:"manufacture_date"`
ExpiryDate      time.Time  `json:"expiry_date" binding:"required"`
Quantity        int        `json:"quantity" binding:"required,gt=0"`
BinID           *uint      `json:"bin_id"`
}

// AdjustBatchQuantityRequest is the body for PUT /warehouse/batches/:id/quantity
// Used when a batch's quantity is consumed (e.g. FEFO pick) or corrected.
type AdjustBatchQuantityRequest struct {
Quantity int    `json:"quantity" binding:"required,gte=0"`
Reason   string `json:"reason" binding:"required"`
}

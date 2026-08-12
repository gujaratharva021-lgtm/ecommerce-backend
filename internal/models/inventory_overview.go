package models

// InventoryOverviewQuery is the query for GET /admin/inventory
type InventoryOverviewQuery struct {
WarehouseID uint   `form:"warehouse_id"`
CategoryID  uint   `form:"category_id"`
StockStatus string `form:"stock_status"` // "low", "out", or empty for all
Page        int    `form:"page,default=1"`
Limit       int    `form:"limit,default=20"`
}

// InventoryRow is one product+warehouse inventory line.
type InventoryRow struct {
ProductID     uint   `json:"product_id"`
ProductName   string `json:"product_name"`
CategoryName  string `json:"category_name"`
WarehouseID   uint   `json:"warehouse_id"`
WarehouseName string `json:"warehouse_name"`
Stock         int    `json:"stock"`
Reserved      int    `json:"reserved"`
Available     int    `json:"available"`
InStock       bool   `json:"in_stock"`
}

// InventoryOverviewResponse is the response for GET /admin/inventory
type InventoryOverviewResponse struct {
TotalSKUs      int64          `json:"total_skus"`
TotalAvailable int64          `json:"total_available_stock"`
TotalReserved  int64          `json:"total_reserved_stock"`
LowStockCount  int64          `json:"low_stock_count"`
OutOfStockCount int64         `json:"out_of_stock_count"`
DamagedStock   int64          `json:"damaged_stock"`
ExpiredStock   int64          `json:"expired_stock"`
Rows           []InventoryRow `json:"rows"`
Page           int            `json:"page"`
Limit          int            `json:"limit"`
Total          int64          `json:"total"`
TotalPages     int            `json:"total_pages"`
}

package models

// WarehouseInventoryQuery is the query for GET /warehouse/inventory (warehouse staff only).
type WarehouseInventoryQuery struct {
Search      string `form:"search"`
CategoryID  uint   `form:"category_id"`
StockStatus string `form:"stock_status"` // "in_stock", "low", "out", "damaged", "expired", or empty for all
ZoneID      uint   `form:"zone_id"`
RackID      uint   `form:"rack_id"`
BinID       uint   `form:"bin_id"`
Page        int    `form:"page,default=1"`
Limit       int    `form:"limit,default=20"`
}

// WarehouseInventoryRow is one product's inventory line at the caller's warehouse.
// Damaged/Expired are derived, not stored buckets: Damaged reflects a recent
// (last 30 days) damaged-reason stock adjustment, Expired reflects Batch rows
// past their expiry date that still carry quantity.
type WarehouseInventoryRow struct {
ProductID      uint    `json:"product_id"`
ProductName    string  `json:"product_name"`
Barcode        string  `json:"barcode,omitempty"`
ImageURL       string  `json:"image_url,omitempty"`
CategoryID     uint    `json:"category_id"`
CategoryName   string  `json:"category_name"`
Stock          int     `json:"stock"`
Reserved       int     `json:"reserved"`
Available      int     `json:"available"`
InStock        bool    `json:"in_stock"`
StockStatus    string  `json:"stock_status"` // in_stock / low / out
BinID          *uint   `json:"bin_id,omitempty"`
BinName        string  `json:"bin_name,omitempty"`
RackName       string  `json:"rack_name,omitempty"`
ZoneName       string  `json:"zone_name,omitempty"`
ExpiredQty     int     `json:"expired_qty"`
LastDamagedAt  *string `json:"last_damaged_at,omitempty"`
LastDamagedQty int     `json:"last_damaged_qty,omitempty"`
}

// WarehouseInventoryResponse wraps the paginated result plus warehouse-wide
// summary counts (computed across the whole warehouse, not just this page).
type WarehouseInventoryResponse struct {
Rows              []WarehouseInventoryRow `json:"rows"`
Page              int                     `json:"page"`
Limit             int                     `json:"limit"`
Total             int64                   `json:"total"`
TotalPages        int                     `json:"total_pages"`
InStockCount      int64                   `json:"in_stock_count"`
LowStockCount     int64                   `json:"low_stock_count"`
OutOfStockCount   int64                   `json:"out_of_stock_count"`
DamagedCount      int64                   `json:"damaged_count"`
ExpiredCount      int64                   `json:"expired_count"`
LowStockThreshold int                     `json:"low_stock_threshold"`
}

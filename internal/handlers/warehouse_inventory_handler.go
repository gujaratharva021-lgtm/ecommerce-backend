package handlers

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetWarehouseInventory godoc
// GET /api/v1/warehouse/inventory?search=&category_id=&stock_status=&zone_id=&rack_id=&bin_id=&page=&limit= (warehouse staff only)
// Scoped to the caller's own warehouse. stock_status accepts in_stock/low/out
// for live stock-count filtering, or damaged/expired for the derived views
// (recent damaged-reason adjustments / unexpired-but-past-date batch rows).
func GetWarehouseInventory(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

var query models.WarehouseInventoryQuery
if err := c.ShouldBindQuery(&query); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if query.Page < 1 {
query.Page = 1
}
if query.Limit < 1 || query.Limit > 100 {
query.Limit = 20
}

// Product IDs currently flagged damaged (adjustment reason=damaged in last 30 days)
// or expired (a Batch past its expiry date that still has quantity) at this warehouse.
damagedCutoff := time.Now().AddDate(0, 0, -30)
var damagedIDs []uint
database.DB.Model(&models.StockMovement{}).
Where("warehouse_id = ? AND reason = ? AND created_at >= ?", warehouseID, models.AdjustReasonDamaged, damagedCutoff).
Distinct("product_id").Pluck("product_id", &damagedIDs)

var expiredIDs []uint
database.DB.Model(&models.Batch{}).
Where("warehouse_id = ? AND expiry_date < ? AND quantity > 0", warehouseID, time.Now()).
Distinct("product_id").Pluck("product_id", &expiredIDs)

db := database.DB.Model(&models.Inventory{}).
Joins("JOIN products ON products.id = inventories.product_id").
Where("inventories.warehouse_id = ?", warehouseID)

if query.Search != "" {
like := "%" + query.Search + "%"
db = db.Where("products.name ILIKE ? OR products.barcode ILIKE ?", like, like)
}
if query.CategoryID > 0 {
db = db.Where("products.category_id = ?", query.CategoryID)
}
if query.BinID > 0 {
db = db.Where("inventories.bin_id = ?", query.BinID)
} else if query.RackID > 0 {
db = db.Joins("JOIN warehouse_bins ON warehouse_bins.id = inventories.bin_id").
Where("warehouse_bins.rack_id = ?", query.RackID)
} else if query.ZoneID > 0 {
db = db.Joins("JOIN warehouse_bins ON warehouse_bins.id = inventories.bin_id").
Joins("JOIN warehouse_racks ON warehouse_racks.id = warehouse_bins.rack_id").
Where("warehouse_racks.zone_id = ?", query.ZoneID)
}

switch query.StockStatus {
case "out":
db = db.Where("inventories.stock <= 0")
case "low":
db = db.Where("inventories.stock > 0 AND inventories.stock < ?", lowStockThreshold)
case "in_stock":
db = db.Where("inventories.stock >= ?", lowStockThreshold)
case "damaged":
filterIDs := damagedIDs
if len(filterIDs) == 0 {
filterIDs = []uint{0}
}
db = db.Where("inventories.product_id IN ?", filterIDs)
case "expired":
filterIDs := expiredIDs
if len(filterIDs) == 0 {
filterIDs = []uint{0}
}
db = db.Where("inventories.product_id IN ?", filterIDs)
}

var total int64
if err := db.Count(&total).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count inventory"})
return
}

offset := (query.Page - 1) * query.Limit
var invs []models.Inventory
if err := db.Preload("Product").Preload("Product.Category").Preload("Bin.Rack.Zone").
Order("inventories.stock ASC").Offset(offset).Limit(query.Limit).Find(&invs).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory"})
return
}

damagedSet := make(map[uint]bool, len(damagedIDs))
for _, id := range damagedIDs {
damagedSet[id] = true
}
expiredSet := make(map[uint]bool, len(expiredIDs))
for _, id := range expiredIDs {
expiredSet[id] = true
}

rows := make([]models.WarehouseInventoryRow, 0, len(invs))
for _, inv := range invs {
var reserved int
database.DB.Model(&models.CartReservation{}).
Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", inv.ProductID, inv.WarehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&reserved)

status := "in_stock"
if inv.Stock <= 0 {
status = "out"
} else if inv.Stock < lowStockThreshold {
status = "low"
}

row := models.WarehouseInventoryRow{
ProductID:   inv.ProductID,
ProductName: inv.Product.Name,
Barcode:     inv.Product.Barcode,
ImageURL:    inv.Product.ImageURL,
CategoryID:  inv.Product.CategoryID,
CategoryName: inv.Product.Category.Name,
Stock:       inv.Stock,
Reserved:    reserved,
Available:   inv.Stock - reserved,
InStock:     inv.InStock,
StockStatus: status,
BinID:       inv.BinID,
}
if inv.Bin != nil {
row.BinName = inv.Bin.Name
row.RackName = inv.Bin.Rack.Name
row.ZoneName = inv.Bin.Rack.Zone.Name
}

if expiredSet[inv.ProductID] {
var expiredQty int
database.DB.Model(&models.Batch{}).
Where("product_id = ? AND warehouse_id = ? AND expiry_date < ? AND quantity > 0", inv.ProductID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&expiredQty)
row.ExpiredQty = expiredQty
}

if damagedSet[inv.ProductID] {
var lastDamaged models.StockMovement
if err := database.DB.Where("product_id = ? AND warehouse_id = ? AND reason = ?", inv.ProductID, warehouseID, models.AdjustReasonDamaged).
Order("created_at DESC").First(&lastDamaged).Error; err == nil {
ts := lastDamaged.CreatedAt.Format(time.RFC3339)
row.LastDamagedAt = &ts
qty := lastDamaged.Change
if qty < 0 {
qty = -qty
}
row.LastDamagedQty = qty
}
}

rows = append(rows, row)
}

// Warehouse-wide summary counts (not just this page).
var inStockCount, lowStockCount, outOfStockCount int64
database.DB.Model(&models.Inventory{}).Where("warehouse_id = ? AND stock >= ?", warehouseID, lowStockThreshold).Count(&inStockCount)
database.DB.Model(&models.Inventory{}).Where("warehouse_id = ? AND stock > 0 AND stock < ?", warehouseID, lowStockThreshold).Count(&lowStockCount)
database.DB.Model(&models.Inventory{}).Where("warehouse_id = ? AND stock <= 0", warehouseID).Count(&outOfStockCount)

c.JSON(http.StatusOK, models.WarehouseInventoryResponse{
Rows:              rows,
Page:              query.Page,
Limit:             query.Limit,
Total:             total,
TotalPages:        int((total + int64(query.Limit) - 1) / int64(query.Limit)),
InStockCount:      inStockCount,
LowStockCount:     lowStockCount,
OutOfStockCount:   outOfStockCount,
DamagedCount:      int64(len(damagedIDs)),
ExpiredCount:      int64(len(expiredIDs)),
LowStockThreshold: lowStockThreshold,
})
}

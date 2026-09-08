package services

import (
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// DefaultLowStockThreshold is the final fallback when neither the specific
// Inventory row nor its Warehouse define an override.
const DefaultLowStockThreshold = 10

// ResolveLowStockThreshold implements the fallback chain:
// Inventory.LowStockThreshold (product+warehouse specific)
//   -> Warehouse.LowStockThreshold (warehouse default)
//   -> DefaultLowStockThreshold (global fallback)
//
// Callers typically have the Inventory row loaded but not always its
// Warehouse association preloaded - pass nil for warehouseThreshold if
// unavailable, and this still resolves correctly to the global default.
func ResolveLowStockThreshold(inventoryThreshold *int, warehouseThreshold *int) int {
if inventoryThreshold != nil {
return *inventoryThreshold
}
if warehouseThreshold != nil {
return *warehouseThreshold
}
return DefaultLowStockThreshold
}

// ResolveLowStockThresholdFor is a convenience wrapper that reads the
// threshold fields directly off a loaded Inventory + Warehouse pair.
func ResolveLowStockThresholdFor(inv models.Inventory, wh models.Warehouse) int {
return ResolveLowStockThreshold(inv.LowStockThreshold, wh.LowStockThreshold)
}

// CountLowAndOutOfStock computes low-stock and out-of-stock counts for a
// warehouse using the per-row threshold fallback chain (Inventory override
// -> Warehouse default -> global default), on AVAILABLE stock (physical
// stock minus active cart reservations minus expired-batch quantity) - not
// raw Inventory.Stock. Used by both the warehouse dashboard and the
// warehouse inventory listing so the two views can never disagree.
func CountLowAndOutOfStock(warehouseID uint) (lowStock int64, outOfStock int64, err error) {
var warehouse models.Warehouse
if err = database.DB.First(&warehouse, warehouseID).Error; err != nil {
return 0, 0, err
}

var invs []models.Inventory
if err = database.DB.Where("warehouse_id = ?", warehouseID).Find(&invs).Error; err != nil {
return 0, 0, err
}

for _, inv := range invs {
var reserved int
database.DB.Model(&models.CartReservation{}).
Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", inv.ProductID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&reserved)

var expiredQty int
database.DB.Model(&models.Batch{}).
Where("product_id = ? AND warehouse_id = ? AND expiry_date < ? AND quantity > 0", inv.ProductID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&expiredQty)

available := inv.Stock - reserved - expiredQty
if available < 0 {
available = 0
}
threshold := ResolveLowStockThresholdFor(inv, warehouse)

if available <= 0 {
outOfStock++
} else if available < threshold {
lowStock++
}
}
return lowStock, outOfStock, nil
}

// CheckAndNotifyLowStock evaluates a single product+warehouse's current
// available stock against its resolved threshold and fires the
// appropriate WhNotifyOutOfStock/WhNotifyLowStock alert if crossed. Meant
// to be called post-transaction (e.g. after checkout commits), typically
// via `go services.CheckAndNotifyLowStock(...)`, so a slow/failing
// notification send never blocks or fails the triggering request.
func CheckAndNotifyLowStock(productID, warehouseID uint) {
var inv models.Inventory
if err := database.DB.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&inv).Error; err != nil {
return
}
var warehouse models.Warehouse
if err := database.DB.First(&warehouse, warehouseID).Error; err != nil {
return
}

var reserved int
database.DB.Model(&models.CartReservation{}).
Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", productID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&reserved)

var expiredQty int
database.DB.Model(&models.Batch{}).
Where("product_id = ? AND warehouse_id = ? AND expiry_date < ? AND quantity > 0", productID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&expiredQty)

available := inv.Stock - reserved - expiredQty
if available < 0 {
available = 0
}
threshold := ResolveLowStockThresholdFor(inv, warehouse)

if available <= 0 {
NotifyWarehouse(warehouseID, models.WhNotifyOutOfStock,
"Product out of stock",
"Product is now out of stock at your warehouse.",
nil, &productID)
} else if available < threshold {
NotifyWarehouse(warehouseID, models.WhNotifyLowStock,
"Low stock warning",
"Product is running low on available stock.",
nil, &productID)
}
}
package services

import (
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// DefaultLowStockThreshold is the final fallback when neither the specific
// Inventory row nor its Warehouse define an override.
const DefaultLowStockThreshold = 10

// Stock alert states persisted on Inventory.StockAlertState - see that
// field's doc comment for why this is DB-authoritative rather than an
// in-process flag.
const (
StockAlertStateNormal    = "normal"
StockAlertStateLow       = "low"
StockAlertStateOutOfStock = "out_of_stock"
)

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

// computeAvailable returns Inventory.Stock minus active cart reservations
// minus expired-batch quantity for a product+warehouse, floored at 0.
// Shared by CountLowAndOutOfStock and CheckAndNotifyLowStock so the two
// views can never disagree on what "available" means.
func computeAvailable(inv models.Inventory) int {
var reserved int
database.DB.Model(&models.CartReservation{}).
Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", inv.ProductID, inv.WarehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&reserved)

var expiredQty int
database.DB.Model(&models.Batch{}).
Where("product_id = ? AND warehouse_id = ? AND expiry_date < ? AND quantity > 0", inv.ProductID, inv.WarehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").Scan(&expiredQty)

available := inv.Stock - reserved - expiredQty
if available < 0 {
available = 0
}
return available
}

// ClassifyStockAlertState maps available stock against a threshold to one
// of the three persisted alert states.
func ClassifyStockAlertState(available, threshold int) string {
if available <= 0 {
return StockAlertStateOutOfStock
}
if available <= threshold {
return StockAlertStateLow
}
return StockAlertStateNormal
}

// CountLowAndOutOfStock computes low-stock and out-of-stock counts for a
// warehouse using the per-row threshold fallback chain (Inventory override
// -> Warehouse default -> global default), on AVAILABLE stock (physical
// stock minus active cart reservations minus expired-batch quantity) - not
// raw Inventory.Stock. Used by both the warehouse dashboard and the
// warehouse inventory listing so the two views can never disagree.
//
// This is a pure read - it does not touch StockAlertState or send any
// notifications. See CheckAndNotifyLowStock for the write/notify path.
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
available := computeAvailable(inv)
threshold := ResolveLowStockThresholdFor(inv, warehouse)
switch ClassifyStockAlertState(available, threshold) {
case StockAlertStateOutOfStock:
outOfStock++
case StockAlertStateLow:
lowStock++
}
}
return lowStock, outOfStock, nil
}

// CheckAndNotifyLowStock evaluates a single product+warehouse's current
// available stock against its resolved threshold, and fires a warehouse
// alert ONLY when that evaluation actually crosses into a new
// Inventory.StockAlertState - never merely because the product is still
// below threshold from a previous check. This is what prevents every
// subsequent sale of an already-low product (or every subsequent manual
// stock adjustment - see handlers.AdjustStock, which now also calls this
// function instead of notifying directly) from re-firing the same alert.
//
// Transition table (old -> new: notify?):
//
//normal -> low            : yes (low_stock)
//low -> low                : no
//low -> out_of_stock      : yes (out_of_stock)
//out_of_stock -> out_of_stock : no
//low -> normal / out_of_stock -> normal : no (recovery is not alert-worthy)
//normal -> low (again, after a restock cleared state back to normal)  : yes
//
// The state transition itself is the concurrency gate: the UPDATE is
// conditioned on the row's CURRENT StockAlertState still matching what was
// just read (`stock_alert_state = ?`), so if two goroutines/instances race
// the same transition (e.g. two concurrent checkouts both taking a product
// from normal to low), only the first UPDATE actually changes a row -
// RowsAffected reports 0 for the loser, which is treated as "someone else
// already handled this transition, don't notify again". This is safe
// across multiple goroutines and multiple backend instances, unlike an
// in-process flag or a read-then-write without a WHERE-matched UPDATE.
//
// Meant to be called post-transaction (e.g. after checkout commits or an
// adjustment saves), typically via `go services.CheckAndNotifyLowStock(...)`,
// so a slow/failing notification send never blocks or fails the triggering
// request.
func CheckAndNotifyLowStock(productID, warehouseID uint) {
var inv models.Inventory
if err := database.DB.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&inv).Error; err != nil {
return
}
var warehouse models.Warehouse
if err := database.DB.First(&warehouse, warehouseID).Error; err != nil {
return
}

available := computeAvailable(inv)
threshold := ResolveLowStockThresholdFor(inv, warehouse)
newState := ClassifyStockAlertState(available, threshold)

oldState := inv.StockAlertState
if oldState == "" {
oldState = StockAlertStateNormal
}
if newState == oldState {
return
}

// The WHERE clause must also match a NULL stock_alert_state whenever
// oldState is "normal" (the in-memory default applied above) - a
// freshly-inserted or freshly-migrated Inventory row has NULL here,
// not the literal string "normal", and Postgres NULL = \'normal\'
// never matches. Without this OR, the very first normal -> low/out_of_stock
// transition on a brand-new row always lost the race-check below
// (RowsAffected == 0) and silently swallowed the alert (Bug#25).
query := database.DB.Model(&models.Inventory{}).Where("id = ?", inv.ID)
if oldState == StockAlertStateNormal {
query = query.Where("(stock_alert_state = ? OR stock_alert_state IS NULL)", oldState)
} else {
query = query.Where("stock_alert_state = ?", oldState)
}
result := query.Update("stock_alert_state", newState)
if result.Error != nil || result.RowsAffected == 0 {
// Either a DB error, or we lost the race to another concurrent
// caller that already made this same transition - either way,
// don't notify.
return
}

// Only low_stock and out_of_stock crossings are alert-worthy;
// recovering to normal is deliberately silent (see doc comment above).
switch newState {
case StockAlertStateOutOfStock:
NotifyWarehouse(warehouseID, models.WhNotifyOutOfStock,
"Product out of stock",
"Product is now out of stock at your warehouse.",
nil, &productID)
case StockAlertStateLow:
NotifyWarehouse(warehouseID, models.WhNotifyLowStock,
"Low stock warning",
"Product is running low on available stock.",
nil, &productID)
}
}
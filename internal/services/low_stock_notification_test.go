package services

import (
"testing"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// resetLowStockNotifTestTables truncates every table this file touches.
// Kept separate from resetBatchTestTables/resetDeliveryAssignmentTables
// since this file additionally needs warehouse_notifications, which
// neither of the others truncate.
func resetLowStockNotifTestTables(t *testing.T) {
t.Helper()
if err := database.DB.Exec("TRUNCATE TABLE warehouse_notifications, batches, inventories, products, warehouses, categories, cart_reservations RESTART IDENTITY CASCADE").Error; err != nil {
t.Fatalf("failed to reset tables: %v", err)
}
}

// countStockNotifications returns how many low_stock/out_of_stock
// WarehouseNotification rows exist for a product+warehouse.
func countStockNotifications(t *testing.T, warehouseID, productID uint, notifType string) int64 {
t.Helper()
var count int64
if err := database.DB.Model(&models.WarehouseNotification{}).
Where("warehouse_id = ? AND product_id = ? AND type = ?", warehouseID, productID, notifType).
Count(&count).Error; err != nil {
t.Fatalf("failed to count notifications: %v", err)
}
return count
}

// ---------------------------------------------------------------------------
// 1. normal -> low fires exactly one notification; repeated sale while
//    still low fires zero additional notifications.
// ---------------------------------------------------------------------------

func TestCheckAndNotifyLowStock_NormalToLow_FiresOnceThenStaysSilent(t *testing.T) {
resetLowStockNotifTestTables(t)
cat := seedBatchTestCategory(t)
product := seedBatchTestProduct(t, cat.ID)
warehouse := seedBatchTestWarehouse(t)
// Stock of 5, default threshold 10 -> classifies as low.
seedBatchTestInventory(t, product.ID, warehouse.ID, 5)

CheckAndNotifyLowStock(product.ID, warehouse.ID)
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyLowStock); got != 1 {
t.Errorf("expected exactly 1 low_stock notification after normal->low, got %d", got)
}

// Simulate another sale that keeps it low (still > 0, still < threshold).
database.DB.Model(&models.Inventory{}).
Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).
Update("stock", 3)
CheckAndNotifyLowStock(product.ID, warehouse.ID)
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyLowStock); got != 1 {
t.Errorf("expected still exactly 1 low_stock notification after low->low, got %d", got)
}
}

// ---------------------------------------------------------------------------
// 2. low -> out_of_stock fires exactly one OOS notification; repeated sale
//    while already OOS fires zero additional notifications.
// ---------------------------------------------------------------------------

func TestCheckAndNotifyLowStock_LowToOutOfStock_FiresOnceThenStaysSilent(t *testing.T) {
resetLowStockNotifTestTables(t)
cat := seedBatchTestCategory(t)
product := seedBatchTestProduct(t, cat.ID)
warehouse := seedBatchTestWarehouse(t)
seedBatchTestInventory(t, product.ID, warehouse.ID, 5)

CheckAndNotifyLowStock(product.ID, warehouse.ID) // normal -> low

database.DB.Model(&models.Inventory{}).
Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).
Update("stock", 0)
CheckAndNotifyLowStock(product.ID, warehouse.ID) // low -> out_of_stock
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyOutOfStock); got != 1 {
t.Errorf("expected exactly 1 out_of_stock notification after low->out_of_stock, got %d", got)
}

// Calling again while still at 0 stock must not re-notify.
CheckAndNotifyLowStock(product.ID, warehouse.ID)
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyOutOfStock); got != 1 {
t.Errorf("expected still exactly 1 out_of_stock notification after OOS->OOS, got %d", got)
}
}

// ---------------------------------------------------------------------------
// 3. Restock clears state back to normal silently (no notification for
//    recovery), and a later dip back into low fires a fresh notification.
// ---------------------------------------------------------------------------

func TestCheckAndNotifyLowStock_RestockClearsState_ThenLowAgainNotifies(t *testing.T) {
resetLowStockNotifTestTables(t)
cat := seedBatchTestCategory(t)
product := seedBatchTestProduct(t, cat.ID)
warehouse := seedBatchTestWarehouse(t)
seedBatchTestInventory(t, product.ID, warehouse.ID, 5)

CheckAndNotifyLowStock(product.ID, warehouse.ID) // normal -> low (1 notif)

// Restock well above threshold - recovery must be silent.
database.DB.Model(&models.Inventory{}).
Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).
Update("stock", 50)
CheckAndNotifyLowStock(product.ID, warehouse.ID)
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyLowStock); got != 1 {
t.Errorf("expected still exactly 1 low_stock notification after recovery to normal (recovery is silent), got %d", got)
}
var inv models.Inventory
database.DB.Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).First(&inv)
if inv.StockAlertState != StockAlertStateNormal {
t.Errorf("expected stock_alert_state to clear to 'normal' after restock, got %q", inv.StockAlertState)
}

// Dip low again - since state was cleared, this must fire a NEW notification.
database.DB.Model(&models.Inventory{}).
Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).
Update("stock", 4)
CheckAndNotifyLowStock(product.ID, warehouse.ID)
if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyLowStock); got != 2 {
t.Errorf("expected 2 low_stock notifications after normal->low a second time post-restock, got %d", got)
}
}

// ---------------------------------------------------------------------------
// 4. Concurrent checkouts crossing the same threshold at once must still
//    fire exactly one notification - the atomic conditional UPDATE is the
//    concurrency gate, not an in-process flag.
// ---------------------------------------------------------------------------

func TestCheckAndNotifyLowStock_ConcurrentCrossing_FiresExactlyOnce(t *testing.T) {
resetLowStockNotifTestTables(t)
cat := seedBatchTestCategory(t)
product := seedBatchTestProduct(t, cat.ID)
warehouse := seedBatchTestWarehouse(t)
seedBatchTestInventory(t, product.ID, warehouse.ID, 5) // already low

const attempts = 10
done := make(chan struct{}, attempts)
for i := 0; i < attempts; i++ {
go func() {
CheckAndNotifyLowStock(product.ID, warehouse.ID)
done <- struct{}{}
}()
}
for i := 0; i < attempts; i++ {
<-done
}

if got := countStockNotifications(t, warehouse.ID, product.ID, models.WhNotifyLowStock); got != 1 {
t.Errorf("expected exactly 1 low_stock notification despite %d concurrent callers, got %d", attempts, got)
}
}

// ---------------------------------------------------------------------------
// 5. Warehouse isolation: the same product going low at one warehouse must
//    not fire (or be counted against) a notification at a different one.
// ---------------------------------------------------------------------------

func TestCheckAndNotifyLowStock_WarehouseIsolation(t *testing.T) {
resetLowStockNotifTestTables(t)
cat := seedBatchTestCategory(t)
product := seedBatchTestProduct(t, cat.ID)
whA := seedBatchTestWarehouse(t)
whB := seedBatchTestWarehouse(t)
seedBatchTestInventory(t, product.ID, whA.ID, 5)  // low
seedBatchTestInventory(t, product.ID, whB.ID, 50) // normal

CheckAndNotifyLowStock(product.ID, whA.ID)
CheckAndNotifyLowStock(product.ID, whB.ID)

if got := countStockNotifications(t, whA.ID, product.ID, models.WhNotifyLowStock); got != 1 {
t.Errorf("expected 1 low_stock notification at warehouse A, got %d", got)
}
if got := countStockNotifications(t, whB.ID, product.ID, models.WhNotifyLowStock); got != 0 {
t.Errorf("expected 0 low_stock notifications at warehouse B (stock is normal there), got %d", got)
}
}

var _ = time.Now // keep time import if unused elsewhere in package build
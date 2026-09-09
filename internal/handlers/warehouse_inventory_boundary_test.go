package handlers

import (
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// newWarehouseInventoryTestRouter wires up GET /warehouse/inventory with a
// stub middleware injecting warehouse_id directly, since GetWarehouseInventory
// only reads that one context value (see routes.go: no extra role gate on
// this GET route beyond WarehouseStaffOnly + InjectWarehouseScope, which
// this stub replaces for test purposes).
func newWarehouseInventoryTestRouter(warehouseID uint) *gin.Engine {
r := gin.New()
r.GET("/api/v1/warehouse/inventory", func(c *gin.Context) {
c.Set("warehouse_id", warehouseID)
GetWarehouseInventory(c)
})
return r
}

func doWarehouseInventoryRequest(r *gin.Engine, path string) *httptest.ResponseRecorder {
req := httptest.NewRequest(http.MethodGet, path, nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
return w
}

// TestGetWarehouseInventory_ExactThresholdBoundary_ClassifiesAsLow is the
// regression test for the boundary-condition drift between
// services.ClassifyStockAlertState (available <= threshold -> low, used by
// notifications and dashboard counts) and this handler's previously
// independent "available < threshold" comparison, which disagreed at
// available == threshold exactly (classified as in_stock instead of low).
//
// With the default threshold of 10 (services.DefaultLowStockThreshold) and
// available stock of exactly 10, this row must now be classified "low" -
// consistent with what a low_stock notification and the dashboard
// low-stock count would already report for the same state.
func TestGetWarehouseInventory_ExactThresholdBoundary_ClassifiesAsLow(t *testing.T) {
resetProductListTestTables(t)
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "boundary product")
warehouse := seedPLWarehouse(t, "Boundary WH", 0, 0)
// No LowStockThreshold override on Inventory or Warehouse, so this
// resolves to the global default of 10 - exactly matching stock.
seedPLInventory(t, product.ID, warehouse.ID, 10, true)

r := newWarehouseInventoryTestRouter(warehouse.ID)
w := doWarehouseInventoryRequest(r, "/api/v1/warehouse/inventory")
if w.Code != http.StatusOK {
t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
}

var resp models.WarehouseInventoryResponse
if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if len(resp.Rows) != 1 {
t.Fatalf("expected 1 row, got %d", len(resp.Rows))
}
if resp.Rows[0].StockStatus != "low" {
t.Errorf("expected stock_status='low' at exactly the threshold (available=10, threshold=10), got %q", resp.Rows[0].StockStatus)
}
if resp.LowStockCount != 1 {
t.Errorf("expected low_stock_count=1, got %d", resp.LowStockCount)
}
if resp.OutOfStockCount != 0 {
t.Errorf("expected out_of_stock_count=0, got %d", resp.OutOfStockCount)
}
}

// TestGetWarehouseInventory_StockStatusFilter_MatchesBoundaryRow checks
// that ?stock_status=low actually returns the exact-boundary row too -
// this is the concrete symptom a warehouse manager would have hit: filtering
// for low stock and missing a product that was already low per notifications
// and dashboard counts.
func TestGetWarehouseInventory_StockStatusFilter_MatchesBoundaryRow(t *testing.T) {
resetProductListTestTables(t)
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "boundary product")
warehouse := seedPLWarehouse(t, "Boundary WH", 0, 0)
seedPLInventory(t, product.ID, warehouse.ID, 10, true)

r := newWarehouseInventoryTestRouter(warehouse.ID)
w := doWarehouseInventoryRequest(r, "/api/v1/warehouse/inventory?stock_status=low")
if w.Code != http.StatusOK {
t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
}

var resp models.WarehouseInventoryResponse
if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Total != 1 {
t.Errorf("expected the exact-threshold row to match ?stock_status=low, got total=%d", resp.Total)
}
}

// TestGetWarehouseInventory_OneUnderThreshold_StillLow is a sanity check
// that the ordinary (non-boundary) low case still works after routing
// through the shared ClassifyStockAlertState function.
func TestGetWarehouseInventory_OneUnderThreshold_StillLow(t *testing.T) {
resetProductListTestTables(t)
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "just under")
warehouse := seedPLWarehouse(t, "Under WH", 0, 0)
seedPLInventory(t, product.ID, warehouse.ID, 9, true)

r := newWarehouseInventoryTestRouter(warehouse.ID)
w := doWarehouseInventoryRequest(r, "/api/v1/warehouse/inventory")

var resp models.WarehouseInventoryResponse
if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Rows[0].StockStatus != "low" {
t.Errorf("expected stock_status='low' for available=9 < threshold=10, got %q", resp.Rows[0].StockStatus)
}
}

// TestGetWarehouseInventory_OneOverThreshold_IsInStock is the corresponding
// sanity check just above the boundary.
func TestGetWarehouseInventory_OneOverThreshold_IsInStock(t *testing.T) {
resetProductListTestTables(t)
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "just over")
warehouse := seedPLWarehouse(t, "Over WH", 0, 0)
seedPLInventory(t, product.ID, warehouse.ID, 11, true)

r := newWarehouseInventoryTestRouter(warehouse.ID)
w := doWarehouseInventoryRequest(r, "/api/v1/warehouse/inventory")

var resp models.WarehouseInventoryResponse
if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Rows[0].StockStatus != "in_stock" {
t.Errorf("expected stock_status='in_stock' for available=11 > threshold=10, got %q", resp.Rows[0].StockStatus)
}
}
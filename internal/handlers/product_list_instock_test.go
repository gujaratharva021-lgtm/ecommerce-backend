package handlers

import (
"encoding/json"
"net/http"
"net/http/httptest"
"testing"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// newProductListTestRouter wires up just GET /products, unauthenticated,
// same as the real route (see routes.go).
func newProductListTestRouter() *gin.Engine {
r := gin.New()
r.GET("/api/v1/products", GetProducts)
return r
}

// resetProductListTestTables truncates every table this file touches.
func resetProductListTestTables(t *testing.T) {
t.Helper()
if err := database.DB.Exec("TRUNCATE TABLE inventories, products, warehouses, categories RESTART IDENTITY CASCADE").Error; err != nil {
t.Fatalf("failed to reset tables: %v", err)
}
}

// seedPLCategory/seedPLProduct/seedPLWarehouse/seedPLInventory: minimal
// seed helpers local to this file (prefixed PL for "product list" to avoid
// colliding with any handler-package helpers of the same shape).
func seedPLCategory(t *testing.T) models.Category {
t.Helper()
cat := models.Category{Name: "Test Category"}
if err := database.DB.Create(&cat).Error; err != nil {
t.Fatalf("failed to seed category: %v", err)
}
return cat
}

func seedPLProduct(t *testing.T, categoryID uint, name string) models.Product {
t.Helper()
product := models.Product{
Name:       name,
Price:      100,
CategoryID: categoryID,
}
if err := database.DB.Create(&product).Error; err != nil {
t.Fatalf("failed to seed product: %v", err)
}
return product
}

// seedPLWarehouse creates an active, open warehouse at the given
// coordinates - both IsActive and Status must be explicitly set since
// FindNearestWarehouse filters on "is_active = ? AND status = ?" and a
// raw Create is not guaranteed to apply gorm struct-tag defaults.
func seedPLWarehouse(t *testing.T, name string, lat, lng float64) models.Warehouse {
t.Helper()
wh := models.Warehouse{
Name:     name,
City:     "Test City",
Lat:      lat,
Lng:      lng,
IsActive: true,
Status:   "open",
}
if err := database.DB.Create(&wh).Error; err != nil {
t.Fatalf("failed to seed warehouse: %v", err)
}
return wh
}

func seedPLInventory(t *testing.T, productID, warehouseID uint, stock int, inStock bool) models.Inventory {
t.Helper()
inv := models.Inventory{
ProductID:   productID,
WarehouseID: warehouseID,
Stock:       stock,
InStock:     inStock,
}
if err := database.DB.Create(&inv).Error; err != nil {
t.Fatalf("failed to seed inventory: %v", err)
}
// GORM skips zero-value fields against a gorm:"default:..." tag on
// Create - InStock:false is a zero value, so it silently gets
// overridden by the schema default (true) instead of actually being
// stored as false. Force it explicitly with a follow-up Update so
// tests can seed a genuinely out-of-stock row.
if err := database.DB.Model(&models.Inventory{}).Where("id = ?", inv.ID).Update("in_stock", inStock).Error; err != nil {
t.Fatalf("failed to force in_stock on seeded inventory: %v", err)
}
inv.InStock = inStock
return inv
}

func doProductListRequest(r *gin.Engine, path string) *httptest.ResponseRecorder {
req := httptest.NewRequest(http.MethodGet, path, nil)
w := httptest.NewRecorder()
r.ServeHTTP(w, req)
return w
}

// parseJSONBody decodes a recorded response body into target.
func parseJSONBody(t *testing.T, w *httptest.ResponseRecorder, target interface{}) error {
	t.Helper()
	return json.Unmarshal(w.Body.Bytes(), target)
}

// ---------------------------------------------------------------------------
// Gap B: in_stock filter must scope to the customer's resolved nearest
// warehouse, not "in stock at ANY warehouse".
// ---------------------------------------------------------------------------

// TestGetProducts_InStockTrue_ExcludesProductOutOfStockAtNearestWarehouse
// is the core regression test for Gap B: a product out of stock at the
// customer's nearest warehouse but in stock at a farther warehouse must
// NOT pass an in_stock=true filter, even though it would still appear
// (correctly) via nearest_stock/nearest_in_stock in an unfiltered list.
func TestGetProducts_InStockTrue_ExcludesProductOutOfStockAtNearestWarehouse(t *testing.T) {
resetProductListTestTables(t)
r := newProductListTestRouter()
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "sambar masala")

// Near warehouse (0,0) - out of stock. Far warehouse (10,10) - in stock.
near := seedPLWarehouse(t, "Near WH", 0, 0)
far := seedPLWarehouse(t, "Far WH", 10, 10)
seedPLInventory(t, product.ID, near.ID, 0, false)
seedPLInventory(t, product.ID, far.ID, 50, true)

w := doProductListRequest(r, "/api/v1/products?search=sambar&lat=0&lng=0&in_stock=true")
if w.Code != http.StatusOK {
t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
}

var resp models.ProductListResponse
if err := parseJSONBody(t, w, &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Total != 0 {
t.Errorf("expected 0 products (out of stock at the resolved nearest warehouse), got total=%d", resp.Total)
}
}

// TestGetProducts_InStockFalse_IncludesProductOutOfStockAtNearestWarehouse
// is the corresponding positive case: the same product must be INCLUDED
// under in_stock=false, consistent with nearest_in_stock=false.
func TestGetProducts_InStockFalse_IncludesProductOutOfStockAtNearestWarehouse(t *testing.T) {
resetProductListTestTables(t)
r := newProductListTestRouter()
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "sambar masala")

near := seedPLWarehouse(t, "Near WH", 0, 0)
far := seedPLWarehouse(t, "Far WH", 10, 10)
seedPLInventory(t, product.ID, near.ID, 0, false)
seedPLInventory(t, product.ID, far.ID, 50, true)

w := doProductListRequest(r, "/api/v1/products?search=sambar&lat=0&lng=0&in_stock=false")
if w.Code != http.StatusOK {
t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
}

var resp models.ProductListResponse
if err := parseJSONBody(t, w, &resp); err != nil {
t.Fatalf("failed to parse response: %v", err)
}
if resp.Total != 1 {
t.Fatalf("expected 1 product (out of stock at the resolved nearest warehouse), got total=%d", resp.Total)
}
if resp.Products[0].NearestInStock == nil || *resp.Products[0].NearestInStock != false {
t.Errorf("expected nearest_in_stock=false, got %v", resp.Products[0].NearestInStock)
}
}

// TestGetProducts_InStockFilter_WarehouseIsolation checks that a product
// in stock at warehouse A but out of stock at warehouse B correctly
// passes in_stock=true for a customer nearest to A, and correctly fails
// it for a customer nearest to B - the filter must track whichever
// warehouse the CALLER resolves to, not a fixed one.
func TestGetProducts_InStockFilter_WarehouseIsolation(t *testing.T) {
resetProductListTestTables(t)
r := newProductListTestRouter()
cat := seedPLCategory(t)
product := seedPLProduct(t, cat.ID, "sambar masala")

whA := seedPLWarehouse(t, "Warehouse A", 0, 0)
whB := seedPLWarehouse(t, "Warehouse B", 50, 50)
seedPLInventory(t, product.ID, whA.ID, 20, true)
seedPLInventory(t, product.ID, whB.ID, 0, false)

// Customer near A: in_stock=true must include it.
wA := doProductListRequest(r, "/api/v1/products?search=sambar&lat=0&lng=0&in_stock=true")
var respA models.ProductListResponse
if err := parseJSONBody(t, wA, &respA); err != nil {
t.Fatalf("failed to parse response A: %v", err)
}
if respA.Total != 1 {
t.Errorf("expected customer near warehouse A to see the product under in_stock=true, got total=%d", respA.Total)
}

// Customer near B: in_stock=true must exclude it.
wB := doProductListRequest(r, "/api/v1/products?search=sambar&lat=50&lng=50&in_stock=true")
var respB models.ProductListResponse
if err := parseJSONBody(t, wB, &respB); err != nil {
t.Fatalf("failed to parse response B: %v", err)
}
if respB.Total != 0 {
t.Errorf("expected customer near warehouse B to NOT see the product under in_stock=true, got total=%d", respB.Total)
}
}
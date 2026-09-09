package services

import (
	"testing"
	"time"

	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
)

// resetBatchTestTables truncates every table this test file touches (and
// restarts identity counters) so each test starts from a clean, predictable
// slate regardless of what earlier tests inserted. Kept separate from
// resetDeliveryAssignmentTables (delivery_assignment_test.go) since the two
// files exercise disjoint table sets within the same package/TestMain.
func resetBatchTestTables(t *testing.T) {
	t.Helper()
	if err := database.DB.Exec("TRUNCATE TABLE batches, inventories, products, warehouses, categories RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset tables: %v", err)
	}
}

// seedBatchTestCategory creates a category row - Product.CategoryID has no
// enforced FK in this schema, but seeding one keeps fixtures realistic and
// matches the pattern used by other test files in this repo.
func seedBatchTestCategory(t *testing.T) models.Category {
	t.Helper()
	cat := models.Category{Name: "Test Category"}
	if err := database.DB.Create(&cat).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}
	return cat
}

// seedBatchTestProduct creates a minimal product for batch-deduction tests.
func seedBatchTestProduct(t *testing.T, categoryID uint) models.Product {
	t.Helper()
	product := models.Product{
		Name:       "Test Product",
		Price:      100,
		CategoryID: categoryID,
	}
	if err := database.DB.Create(&product).Error; err != nil {
		t.Fatalf("failed to seed product: %v", err)
	}
	return product
}

// seedBatchTestWarehouse creates a minimal warehouse for batch-deduction tests.
func seedBatchTestWarehouse(t *testing.T) models.Warehouse {
	t.Helper()
	wh := models.Warehouse{Name: "Test Warehouse", City: "Testville", IsActive: true}
	if err := database.DB.Create(&wh).Error; err != nil {
		t.Fatalf("failed to seed warehouse: %v", err)
	}
	return wh
}

// seedBatchTestInventory creates the Inventory row backing a product at a
// warehouse. DeductFromBatchesFEFO itself never touches Inventory.Stock
// (that's the caller's job, e.g. order_handler.go's checkout loop) - this
// exists purely so fixtures resemble a real product/warehouse pairing.
func seedBatchTestInventory(t *testing.T, productID, warehouseID uint, stock int) models.Inventory {
	t.Helper()
	inv := models.Inventory{
		ProductID:   productID,
		WarehouseID: warehouseID,
		Stock:       stock,
		InStock:     stock > 0,
	}
	if err := database.DB.Create(&inv).Error; err != nil {
		t.Fatalf("failed to seed inventory: %v", err)
	}
	return inv
}

// seedBatch creates a Batch row with the given expiry offset (days from
// now; negative = already expired) and quantity.
func seedBatch(t *testing.T, productID, warehouseID uint, batchNumber string, expiryDaysFromNow, quantity int) models.Batch {
	t.Helper()
	batch := models.Batch{
		ProductID:        productID,
		WarehouseID:      warehouseID,
		BatchNumber:      batchNumber,
		ExpiryDate:       time.Now().AddDate(0, 0, expiryDaysFromNow),
		Quantity:         quantity,
		CreatedByStaffID: 1,
	}
	if err := database.DB.Create(&batch).Error; err != nil {
		t.Fatalf("failed to seed batch %s: %v", batchNumber, err)
	}
	return batch
}

// reloadBatch re-fetches a batch's current row from the DB by ID.
func reloadBatch(t *testing.T, id uint) models.Batch {
	t.Helper()
	var b models.Batch
	if err := database.DB.First(&b, id).Error; err != nil {
		t.Fatalf("failed to reload batch %d: %v", id, err)
	}
	return b
}

// runDeduct is a small wrapper so every test calls DeductFromBatchesFEFO
// inside its own transaction, matching how the real callers (order
// checkout, substitution approval) always invoke it inside an existing tx.
func runDeduct(t *testing.T, productID, warehouseID uint, qty int) error {
	t.Helper()
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return DeductFromBatchesFEFO(tx, productID, warehouseID, qty)
	})
}

// ---------------------------------------------------------------------------
// 1. Multi-batch split
// ---------------------------------------------------------------------------

// TestDeductFromBatchesFEFO_MultiBatchSplit is the core FEFO regression: an
// order quantity larger than the earliest-expiry batch alone must spill
// over into the next-earliest batch, in expiry order - not whichever batch
// was created/found first, and not split evenly.
func TestDeductFromBatchesFEFO_MultiBatchSplit(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 8)

	// EARLY expires first but only holds 3; LATE has plenty but should
	// only be touched for the remainder.
	early := seedBatch(t, product.ID, warehouse.ID, "EARLY", 5, 3)
	late := seedBatch(t, product.ID, warehouse.ID, "LATE", 30, 5)

	if err := runDeduct(t, product.ID, warehouse.ID, 4); err != nil {
		t.Fatalf("expected deduction to succeed, got %v", err)
	}

	freshEarly := reloadBatch(t, early.ID)
	freshLate := reloadBatch(t, late.ID)

	if freshEarly.Quantity != 0 {
		t.Errorf("expected EARLY batch fully drained to 0, got %d", freshEarly.Quantity)
	}
	if freshLate.Quantity != 4 {
		t.Errorf("expected LATE batch to cover only the 1-unit remainder (5-1=4), got %d", freshLate.Quantity)
	}
}

// ---------------------------------------------------------------------------
// 2. Exact-boundary drain
// ---------------------------------------------------------------------------

// TestDeductFromBatchesFEFO_ExactBoundaryDrain checks that consuming exactly
// a batch's remaining quantity leaves it at precisely 0 - never negative,
// and without erroring.
func TestDeductFromBatchesFEFO_ExactBoundaryDrain(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 4)

	batch := seedBatch(t, product.ID, warehouse.ID, "EXACT", 10, 4)

	if err := runDeduct(t, product.ID, warehouse.ID, 4); err != nil {
		t.Fatalf("expected exact-boundary deduction to succeed, got %v", err)
	}

	fresh := reloadBatch(t, batch.ID)
	if fresh.Quantity != 0 {
		t.Errorf("expected batch drained to exactly 0, got %d", fresh.Quantity)
	}
	if fresh.Quantity < 0 {
		t.Fatalf("batch quantity must never go negative, got %d", fresh.Quantity)
	}
}

// ---------------------------------------------------------------------------
// 3. Non-batch-tracked product -> no-op
// ---------------------------------------------------------------------------

// TestDeductFromBatchesFEFO_NoBatchTrackedProductIsNoOp checks that a
// product with zero Batch rows (non-perishable / not opted into batch
// tracking) never blocks or errors the deduction - this must remain a
// silent no-op per the function's documented best-effort semantics, since
// Inventory.Stock (already deducted by the caller before this runs) is the
// authoritative figure and a sale must never be reversed just because
// batch bookkeeping has nothing to draw from.
func TestDeductFromBatchesFEFO_NoBatchTrackedProductIsNoOp(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 10)
	// Deliberately no Batch rows created for this product/warehouse.

	if err := runDeduct(t, product.ID, warehouse.ID, 2); err != nil {
		t.Fatalf("expected no-op success for a non-batch-tracked product, got error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// 4. Expired batch is never selected
// ---------------------------------------------------------------------------

// TestDeductFromBatchesFEFO_ExpiredBatchNeverSelected is the guard against a
// future refactor reintroducing expired-stock consumption: even though an
// expired batch is technically the "earliest expiry" (it's in the past),
// FEFO selection must skip it entirely and draw only from a live batch -
// checkout already excludes expired quantity from the availability check
// (GetExpiredBatchQty), so this deduction path must honor the same rule and
// never touch expired batch rows.
func TestDeductFromBatchesFEFO_ExpiredBatchNeverSelected(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 8)

	expired := seedBatch(t, product.ID, warehouse.ID, "EXPIRED", -2, 5)
	live := seedBatch(t, product.ID, warehouse.ID, "LIVE", 10, 5)

	if err := runDeduct(t, product.ID, warehouse.ID, 3); err != nil {
		t.Fatalf("expected deduction to succeed, got %v", err)
	}

	freshExpired := reloadBatch(t, expired.ID)
	freshLive := reloadBatch(t, live.ID)

	if freshExpired.Quantity != 5 {
		t.Errorf("expired batch must never be touched, expected quantity unchanged at 5, got %d", freshExpired.Quantity)
	}
	if freshLive.Quantity != 2 {
		t.Errorf("expected the live batch to absorb the full deduction (5-3=2), got %d", freshLive.Quantity)
	}
}

package services

import (
	"testing"
	"time"

	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// intPtr is a small helper since Go has no literal syntax for a pointer to
// an int constant.
func intPtr(v int) *int { return &v }

// ---------------------------------------------------------------------------
// 1. Threshold fallback chain (matrix items #1, #2, #3)
// ---------------------------------------------------------------------------

// TestResolveLowStockThreshold_GlobalFallback checks that with neither an
// inventory-specific override nor a warehouse default set, the global
// DefaultLowStockThreshold (10) applies. Matrix item #3.
func TestResolveLowStockThreshold_GlobalFallback(t *testing.T) {
	got := ResolveLowStockThreshold(nil, nil)
	if got != DefaultLowStockThreshold {
		t.Errorf("expected global fallback %d, got %d", DefaultLowStockThreshold, got)
	}
}

// TestResolveLowStockThreshold_WarehouseDefaultApplies checks that a
// warehouse-level default is used when no product+warehouse-specific
// override exists. Matrix item #1.
func TestResolveLowStockThreshold_WarehouseDefaultApplies(t *testing.T) {
	got := ResolveLowStockThreshold(nil, intPtr(25))
	if got != 25 {
		t.Errorf("expected warehouse default 25 to apply, got %d", got)
	}
}

// TestResolveLowStockThreshold_InventoryOverrideWins checks that a
// product+warehouse-specific override always wins over the warehouse
// default, even when both are set. Matrix item #2.
func TestResolveLowStockThreshold_InventoryOverrideWins(t *testing.T) {
	got := ResolveLowStockThreshold(intPtr(3), intPtr(25))
	if got != 3 {
		t.Errorf("expected inventory-specific override 3 to win over warehouse default 25, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// 4. Reserved stock reduces available (matrix item #4)
// ---------------------------------------------------------------------------

// TestCountLowAndOutOfStock_ReservedStockReducesAvailable checks that an
// active cart reservation is subtracted from physical stock before
// comparing against the threshold - a product that looks fine on raw Stock
// alone can still count as low/out-of-stock once reservations are
// accounted for.
func TestCountLowAndOutOfStock_ReservedStockReducesAvailable(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	// Physical stock of 10, well above the global threshold of 10 on its
	// own... but 9 units are reserved by another user, leaving only 1
	// available - which is below the default threshold of 10.
	seedBatchTestInventory(t, product.ID, warehouse.ID, 10)

	user := seedReservationTestUser(t, "9400000001")
	reservation := models.CartReservation{
		UserID:      user.ID,
		ProductID:   product.ID,
		WarehouseID: warehouse.ID,
		Quantity:    9,
		ExpiresAt:   time.Now().Add(10 * time.Minute),
	}
	if err := database.DB.Create(&reservation).Error; err != nil {
		t.Fatalf("failed to seed reservation: %v", err)
	}

	low, oos, err := CountLowAndOutOfStock(warehouse.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock failed: %v", err)
	}
	if oos != 0 {
		t.Errorf("expected 0 out-of-stock (1 unit still available), got %d", oos)
	}
	if low != 1 {
		t.Errorf("expected 1 low-stock product (available=1 < threshold=10 despite raw stock=10), got %d", low)
	}
}

// ---------------------------------------------------------------------------
// 5. Expired batch quantity excluded from availability (matrix item #5)
// ---------------------------------------------------------------------------

// TestCountLowAndOutOfStock_ExpiredBatchExcludedFromAvailable checks that
// stock sitting in an expired batch cannot count as available, even though
// it is still physically present in Inventory.Stock - it must push the
// product toward (or into) low-stock/out-of-stock the same as if the units
// simply weren't there.
func TestCountLowAndOutOfStock_ExpiredBatchExcludedFromAvailable(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	// Raw stock of 10 (above threshold on paper), but all 10 units sit in
	// an already-expired batch - available must be treated as 0.
	seedBatchTestInventory(t, product.ID, warehouse.ID, 10)
	seedBatch(t, product.ID, warehouse.ID, "ALL-EXPIRED", -1, 10)

	low, oos, err := CountLowAndOutOfStock(warehouse.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock failed: %v", err)
	}
	if oos != 1 {
		t.Errorf("expected the product to count as out-of-stock (all stock expired), got oos=%d low=%d", oos, low)
	}
}

// TestCountLowAndOutOfStock_PartiallyExpiredBatchReducesAvailable is the
// partial counterpart: only some of the stock is expired, so the product
// should land as low-stock (not fully out-of-stock) once the expired
// portion is excluded.
func TestCountLowAndOutOfStock_PartiallyExpiredBatchReducesAvailable(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	// Raw stock 10; 8 of those units are in an expired batch, leaving only
	// 2 genuinely available - below the default threshold of 10, but not 0.
	seedBatchTestInventory(t, product.ID, warehouse.ID, 10)
	seedBatch(t, product.ID, warehouse.ID, "EXPIRED-PORTION", -1, 8)

	low, oos, err := CountLowAndOutOfStock(warehouse.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock failed: %v", err)
	}
	if oos != 0 {
		t.Errorf("expected 0 out-of-stock (2 units still genuinely available), got %d", oos)
	}
	if low != 1 {
		t.Errorf("expected 1 low-stock product (available=2 < threshold=10), got %d", low)
	}
}

// ---------------------------------------------------------------------------
// 9. Warehouse isolation
// ---------------------------------------------------------------------------

// TestCountLowAndOutOfStock_WarehouseIsolation checks that stock levels and
// thresholds for the same product at two different warehouses never leak
// into each other - a critically low product at warehouse A must not
// affect warehouse B's counts, and a warehouse-specific threshold override
// at A must not apply to B.
func TestCountLowAndOutOfStock_WarehouseIsolation(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)

	// Warehouse A: out of stock, and a tight custom threshold that would
	// not matter here anyway (0 available is 0 available under any threshold).
	whA := seedBatchTestWarehouse(t)
	if err := database.DB.Model(&whA).Update("low_stock_threshold", 2).Error; err != nil {
		t.Fatalf("failed to set warehouse A threshold: %v", err)
	}
	seedBatchTestInventory(t, product.ID, whA.ID, 0)

	// Warehouse B: healthy stock, well above the global default threshold,
	// and no override - must resolve to the global default (10), not leak
	// warehouse A's override of 2.
	whB := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, whB.ID, 50)

	lowA, oosA, err := CountLowAndOutOfStock(whA.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock(A) failed: %v", err)
	}
	if oosA != 1 {
		t.Errorf("expected warehouse A to show 1 out-of-stock product, got %d", oosA)
	}
	// The product is fully out-of-stock at A, not partially low - the
	// two counts are mutually exclusive per product, so lowA must be 0.
	if lowA != 0 {
		t.Errorf("expected warehouse A to show 0 low-stock (product is fully out-of-stock, not low), got %d", lowA)
	}

	lowB, oosB, err := CountLowAndOutOfStock(whB.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock(B) failed: %v", err)
	}
	if oosB != 0 || lowB != 0 {
		t.Errorf("expected warehouse B to be unaffected by warehouse A's stock/threshold, got low=%d oos=%d", lowB, oosB)
	}
}

// ---------------------------------------------------------------------------
// 10. Non-batch-tracked product still gets normal low-stock treatment
// ---------------------------------------------------------------------------

// TestCountLowAndOutOfStock_NonBatchProductStillWorks checks that a product
// with zero Batch rows (not opted into batch/expiry tracking) still gets
// correctly evaluated against its threshold using plain Inventory.Stock -
// the expired-batch exclusion must be a no-op adjustment (subtracting 0),
// not something that breaks or skips non-batch-tracked products entirely.
func TestCountLowAndOutOfStock_NonBatchProductStillWorks(t *testing.T) {
	resetBatchTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	// No Batch rows created at all for this product.
	seedBatchTestInventory(t, product.ID, warehouse.ID, 5) // below default threshold of 10

	low, oos, err := CountLowAndOutOfStock(warehouse.ID)
	if err != nil {
		t.Fatalf("CountLowAndOutOfStock failed: %v", err)
	}
	if oos != 0 {
		t.Errorf("expected 0 out-of-stock (5 units genuinely available), got %d", oos)
	}
	if low != 1 {
		t.Errorf("expected 1 low-stock product (available=5 < threshold=10) even with no batch tracking, got %d", low)
	}
}

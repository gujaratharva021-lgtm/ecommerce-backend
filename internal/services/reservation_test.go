package services

import (
	"sync"
	"testing"
	"time"

	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
)

// resetReservationTestTables truncates every table this test file touches
// (and restarts identity counters), kept separate from the other reset
// helpers in this package since it covers a disjoint table set
// (cart_reservations layered on top of the product/warehouse/inventory
// tables shared with stock_transfer_batch_test.go).
func resetReservationTestTables(t *testing.T) {
	t.Helper()
	if err := database.DB.Exec("TRUNCATE TABLE cart_reservations, batches, inventories, products, warehouses, categories, users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset tables: %v", err)
	}
}

// runReserve wraps ReserveStock in its own transaction, mirroring how the
// real caller (AddToCart) always invokes it inside database.DB.Transaction.
func runReserve(userID, productID, warehouseID uint, quantity int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		return ReserveStock(tx, userID, productID, warehouseID, quantity)
	})
}

// seedReservationTestUser creates a real User row - CartReservation.UserID
// carries an enforced FK constraint (fk_cart_reservations_user), so an
// arbitrary/unseeded ID fails the insert rather than the reservation logic
// itself.
func seedReservationTestUser(t *testing.T, phone string) models.User {
	t.Helper()
	user := models.User{Name: "Test User", Phone: phone}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user %s: %v", phone, err)
	}
	return user
}

// TestReserveStock_ConcurrentUsers_NeverOversellLastUnit is the core
// concurrency regression: two different users racing to reserve the same
// single remaining unit must never both succeed. ReserveStock's own doc
// comment states the caller must lock the Inventory row (SELECT ... FOR
// UPDATE) before counting other users' reservations - this test proves
// that locking actually serializes the two attempts rather than letting
// both read the same "available" snapshot.
//
// Reservations are keyed per (user, product, warehouse) - see
// CartReservation's unique index - so this must use two distinct users;
// the same user calling twice would just upsert their own hold rather than
// compete for stock.
func TestReserveStock_ConcurrentUsers_NeverOversellLastUnit(t *testing.T) {
	resetReservationTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 1) // exactly 1 unit available

	userA := seedReservationTestUser(t, "9100000101")
	userB := seedReservationTestUser(t, "9100000102")

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		errs[0] = runReserve(userA.ID, product.ID, warehouse.ID, 1)
	}()
	go func() {
		defer wg.Done()
		errs[1] = runReserve(userB.ID, product.ID, warehouse.ID, 1)
	}()
	wg.Wait()

	successCount := 0
	insufficientCount := 0
	for _, err := range errs {
		switch err {
		case nil:
			successCount++
		case ErrInsufficientStock:
			insufficientCount++
		default:
			t.Errorf("unexpected error from concurrent reserve: %v", err)
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful reservation for the last unit, got %d", successCount)
	}
	if insufficientCount != 1 {
		t.Errorf("expected exactly 1 rejected reservation (ErrInsufficientStock), got %d", insufficientCount)
	}

	// Confirm the DB agrees: exactly one reservation row, for 1 unit, and
	// total reserved never exceeds physical stock.
	var totalReserved int64
	if err := database.DB.Model(&models.CartReservation{}).
		Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", product.ID, warehouse.ID, time.Now()).
		Select("COALESCE(SUM(quantity), 0)").Scan(&totalReserved).Error; err != nil {
		t.Fatalf("failed to sum reservations: %v", err)
	}
	if totalReserved != 1 {
		t.Errorf("expected total reserved quantity to be exactly 1 (no overselling), got %d", totalReserved)
	}

	var reservationCount int64
	database.DB.Model(&models.CartReservation{}).
		Where("product_id = ? AND warehouse_id = ?", product.ID, warehouse.ID).
		Count(&reservationCount)
	if reservationCount != 1 {
		t.Errorf("expected exactly 1 reservation row, got %d", reservationCount)
	}
}

// TestReserveStock_ConcurrentUsers_HigherDemandThanStock scales the same
// race up: 5 users, each wanting 1 unit, but only 3 units available.
// Exactly 3 must succeed and 2 must be rejected - proving the serialization
// holds under contention wider than a single unit, not just the trivial
// 1-vs-1 case above.
func TestReserveStock_ConcurrentUsers_HigherDemandThanStock(t *testing.T) {
	resetReservationTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 3)

	const numUsers = 5
	userIDs := make([]uint, numUsers)
	for i := 0; i < numUsers; i++ {
		u := seedReservationTestUser(t, "920000020"+string(rune('0'+i)))
		userIDs[i] = u.ID
	}

	var wg sync.WaitGroup
	errs := make([]error, numUsers)
	wg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(idx int) {
			defer wg.Done()
			errs[idx] = runReserve(userIDs[idx], product.ID, warehouse.ID, 1)
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		} else if err != ErrInsufficientStock {
			t.Errorf("unexpected error from concurrent reserve: %v", err)
		}
	}
	if successCount != 3 {
		t.Errorf("expected exactly 3 successful reservations (matching available stock), got %d", successCount)
	}

	var totalReserved int64
	database.DB.Model(&models.CartReservation{}).
		Where("product_id = ? AND warehouse_id = ? AND expires_at > ?", product.ID, warehouse.ID, time.Now()).
		Select("COALESCE(SUM(quantity), 0)").Scan(&totalReserved)
	if totalReserved != 3 {
		t.Errorf("expected total reserved quantity to be exactly 3 (no overselling), got %d", totalReserved)
	}
}

// TestReserveStock_ExpiredReservationFreesStock checks that a reservation
// past its TTL is treated as if it never existed - expireStaleReservations
// must clean it up lazily on the very next reserve attempt, not require the
// separate periodic sweep to have run first.
func TestReserveStock_ExpiredReservationFreesStock(t *testing.T) {
	resetReservationTestTables(t)
	cat := seedBatchTestCategory(t)
	product := seedBatchTestProduct(t, cat.ID)
	warehouse := seedBatchTestWarehouse(t)
	seedBatchTestInventory(t, product.ID, warehouse.ID, 1)

	userA := seedReservationTestUser(t, "9300000301")
	userB := seedReservationTestUser(t, "9300000302")

	if err := runReserve(userA.ID, product.ID, warehouse.ID, 1); err != nil {
		t.Fatalf("setup: expected first reservation to succeed, got %v", err)
	}

	// Force userA's reservation to already be expired.
	if err := database.DB.Model(&models.CartReservation{}).
		Where("user_id = ? AND product_id = ? AND warehouse_id = ?", userA.ID, product.ID, warehouse.ID).
		Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("failed to force-expire reservation: %v", err)
	}

	// userB should now see the unit as available again.
	if err := runReserve(userB.ID, product.ID, warehouse.ID, 1); err != nil {
		t.Fatalf("expected userB to succeed once userA's hold expired, got %v", err)
	}

	var stillExists int64
	database.DB.Model(&models.CartReservation{}).
		Where("user_id = ?", userA.ID).Count(&stillExists)
	if stillExists != 0 {
		t.Errorf("expected userA's expired reservation row to be cleaned up, but it still exists")
	}
}

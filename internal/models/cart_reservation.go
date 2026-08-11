package models

import "time"

// CartReservation is a short-lived hold on stock for one (user, product,
// warehouse) combination, created when an item is added to (or updated in)
// a cart. It exists to give an accurate "in stock" picture to *other*
// shoppers browsing the same product — it does NOT replace the real,
// transactional stock deduction that happens at Checkout (which still uses
// SELECT ... FOR UPDATE on the Inventory row and is the actual source of
// truth). If the hold expires before checkout, the stock silently becomes
// available to others again; nothing needs to "undo" it, since a reserved
// unit was never actually deducted from Inventory.Stock.
//
// One row per (UserID, ProductID, WarehouseID) — updating quantity/expiry
// in place (upsert) rather than creating duplicates, so a user changing
// the quantity of the same cart item doesn't pile up rows.
type CartReservation struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_reservation_user_product_wh" json:"user_id"`
	ProductID   uint      `gorm:"not null;uniqueIndex:idx_reservation_user_product_wh;index:idx_reservation_product_wh" json:"product_id"`
	WarehouseID uint      `gorm:"not null;uniqueIndex:idx_reservation_user_product_wh;index:idx_reservation_product_wh" json:"warehouse_id"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

package database

import "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"

// GetTotalStock returns the combined available stock for a product across
// all warehouses. Only warehouse-inventory rows marked in_stock count toward
// the total — a warehouse that has zeroed out its stock or flagged a product
// unavailable does not contribute.
//
// NOTE: this is a Phase 3 simplification (single combined availability
// number). Routing an order to a specific nearest warehouse comes in a
// later phase; for now, checkout/cart treat "in stock somewhere" as
// purchasable.
func GetTotalStock(productID uint) (int, error) {
var total int
err := DB.Model(&models.Inventory{}).
Where("product_id = ? AND in_stock = ?", productID, true).
Select("COALESCE(SUM(stock), 0)").
Scan(&total).Error
return total, err
}
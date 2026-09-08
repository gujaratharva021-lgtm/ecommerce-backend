package services

import (
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
)

// GetExpiredBatchQty returns the total quantity sitting in expired batches
// (expiry_date < now, quantity > 0) for a product at a warehouse. This
// represents units that are physically counted in Inventory.Stock but must
// not be sold or reserved, since Batch tracking is a subset of total stock
// (non-batch-tracked stock is unaffected and always sellable) rather than
// the authoritative stock figure - see Batch model comment.
//
// Returns 0 (not an error) when the product has no batches at all, or none
// expired - the common case for non-perishable / non-batch-tracked products.
func GetExpiredBatchQty(tx *gorm.DB, productID, warehouseID uint) (int, error) {
var expiredQty int
err := tx.Model(&models.Batch{}).
Where("product_id = ? AND warehouse_id = ? AND expiry_date < ? AND quantity > 0", productID, warehouseID, time.Now()).
Select("COALESCE(SUM(quantity), 0)").
Scan(&expiredQty).Error
return expiredQty, err
}
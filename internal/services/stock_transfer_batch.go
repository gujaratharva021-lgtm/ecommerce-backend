package services

import (
"errors"
"fmt"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// ErrBatchInsufficientQuantity is returned when the specified/auto-picked
// batch does not have enough quantity to cover the requested amount.
var ErrBatchInsufficientQuantity = errors.New("selected batch does not have enough quantity for this transfer")

// DeductFromBatchFEFO reduces a batch's tracked quantity by qty, inside the
// caller's transaction, under a row lock. If batchID is nil, it picks the
// earliest-expiry batch with quantity > 0 for the given product/warehouse
// (First-Expiry-First-Out) automatically. If the product has no batches at
// all (non-perishable / not batch-tracked), this is a no-op - batch
// tracking is opt-in per Batch model semantics, so absence of batches is
// not an error.
func DeductFromBatchFEFO(tx *gorm.DB, productID, warehouseID uint, batchID *uint, qty int) error {
var batch models.Batch

q := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("product_id = ? AND warehouse_id = ? AND quantity > 0", productID, warehouseID)

if batchID != nil {
q = q.Where("id = ?", *batchID)
} else {
q = q.Order("expiry_date ASC")
}

if err := q.First(&batch).Error; err != nil {
if err == gorm.ErrRecordNotFound {
// No batch-tracked stock for this product at this warehouse -
// treat as not batch-tracked and skip silently.
return nil
}
return err
}

if batch.Quantity < qty {
return ErrBatchInsufficientQuantity
}

batch.Quantity -= qty
if err := tx.Save(&batch).Error; err != nil {
return err
}
return nil
}

// CreateReceivedBatch creates a new Batch row at the destination warehouse
// carrying over the batch number and expiry date from the source batch, so
// expiry tracking survives a stock transfer instead of being lost. Used by
// ReceiveStockTransfer. No-op (returns nil, nil) if sourceBatchID is nil.
func CreateReceivedBatch(tx *gorm.DB, sourceBatchID *uint, productID, warehouseID uint, quantity int, binID *uint, staffID uint) error {
if sourceBatchID == nil {
return nil
}
var source models.Batch
if err := tx.First(&source, *sourceBatchID).Error; err != nil {
return fmt.Errorf("source batch not found for transfer: %w", err)
}
newBatch := models.Batch{
ProductID:        productID,
WarehouseID:      warehouseID,
BatchNumber:      source.BatchNumber,
ManufactureDate:  source.ManufactureDate,
ExpiryDate:       source.ExpiryDate,
Quantity:         quantity,
BinID:            binID,
CreatedByStaffID: staffID,
}
return tx.Create(&newBatch).Error
}
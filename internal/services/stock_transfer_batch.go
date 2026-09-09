package services

import (
"errors"
"fmt"
"time"

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
func DeductFromBatchFEFO(tx *gorm.DB, productID, warehouseID uint, batchID *uint, qty int) (*uint, error) {
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
return nil, nil
}
return nil, err
}

if batch.Quantity < qty {
return nil, ErrBatchInsufficientQuantity
}

batch.Quantity -= qty
if err := tx.Save(&batch).Error; err != nil {
return nil, err
}
selectedID := batch.ID
return &selectedID, nil
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

// DeductFromBatchesFEFO consumes qty units across one or more batches for a
// product/warehouse, in FEFO order (earliest expiry first), inside the
// caller's transaction under row locks. Unlike DeductFromBatchFEFO (which
// targets exactly one batch and errors if that batch alone can't cover qty
// - fine for a transfer, where the requester explicitly picked a batch),
// this is for order checkout: the customer doesn't pick a batch, and stock
// legitimately spans multiple batches (e.g. two 5-unit batches covering an
// 8-unit order).
//
// Only expired batches (see GetExpiredBatchQty) are excluded from
// selection - checkout has already verified overall non-expired
// availability before calling this. If Inventory.Stock is not fully
// covered by batch rows (product only partially batch-tracked, or a
// batch was deleted/consumed out of band), this deducts what it can find
// and returns without error - Batch tracking is best-effort bookkeeping
// on top of the authoritative Inventory.Stock figure, never a hard block
// on a sale that Inventory.Stock has already approved.
func DeductFromBatchesFEFO(tx *gorm.DB, productID, warehouseID uint, qty int) error {
remaining := qty
for remaining > 0 {
var batch models.Batch
err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("product_id = ? AND warehouse_id = ? AND quantity > 0 AND expiry_date >= ?",
productID, warehouseID, time.Now()).
Order("expiry_date ASC").
First(&batch).Error
if err != nil {
if err == gorm.ErrRecordNotFound {
// No more non-expired batch-tracked stock to draw from -
// stop silently, per the best-effort semantics above.
return nil
}
return err
}

take := batch.Quantity
if take > remaining {
take = remaining
}
batch.Quantity -= take
if err := tx.Save(&batch).Error; err != nil {
return err
}
remaining -= take
}
return nil
}
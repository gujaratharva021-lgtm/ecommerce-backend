package services

import (
"errors"
"fmt"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// cancellableFulfillmentStatuses are the order statuses from which a
// fulfillment-stage cancellation (admin or inventory manager) is allowed.
// Deliberately excludes handed_over/shipped/delivered - once the rider has
// the package, this is no longer a warehouse-side cancellation.
var cancellableFulfillmentStatuses = map[string]bool{
models.OrderStatusConfirmed:        true,
models.OrderStatusPicking:          true,
models.OrderStatusPicked:           true,
models.OrderStatusPacking:          true,
models.OrderStatusPacked:           true,
models.OrderStatusReadyForDispatch: true,
}

// CancelOrderInFulfillment cancels an order that has progressed past the
// plain customer-cancellable window (pending/confirmed only, see
// handlers.CancelOrder) but hasn't yet been handed over to a delivery
// partner. Callable by admin or an inventory manager - actorType records
// which, for audit purposes.
//
// Restores inventory, reverses wallet/coupon usage exactly like the
// customer-facing cancel path, additionally unassigns any delivery
// partner already attached (auto-assign can fire as early as order
// confirmation), and cancels any in-flight picking/packing task so the
// warehouse UI doesn't keep showing dead work.
func CancelOrderInFulfillment(orderID uint, actorType string, actorID uint, actorName, reason string) error {
if reason == "" {
return errors.New("a cancellation reason is required")
}

var order models.Order
var payment models.Payment
var gatewayRefundAmount float64
var previouslyAssignedPartnerID *uint

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Preload("Items.Product").First(&order, orderID).Error; err != nil {
return errors.New("order not found")
}
if !cancellableFulfillmentStatuses[order.Status] {
return fmt.Errorf("cannot cancel order from status %q", order.Status)
}

// Restore inventory for each item, same warehouse-aware logic as
// the customer-facing CancelOrder path.
for _, item := range order.Items {
var inventory models.Inventory
q := tx.Where("product_id = ?", item.ProductID)
if order.WarehouseID != nil {
q = q.Where("warehouse_id = ?", *order.WarehouseID)
}
if err := q.Order("id").First(&inventory).Error; err == nil {
previousQty := inventory.Stock
inventory.Stock += item.Quantity
inventory.InStock = true
if err := tx.Save(&inventory).Error; err != nil {
return err
}
movement := models.StockMovement{
ProductID:    item.ProductID,
WarehouseID:  inventory.WarehouseID,
PreviousQty:  previousQty,
Change:       item.Quantity,
NewQty:       inventory.Stock,
MovementType: models.MovementReturn,
Reason:       "Order cancelled in fulfillment: " + reason,
ReferenceID:  &order.ID,
}
if err := tx.Create(&movement).Error; err != nil {
return err
}
}
}

if order.WalletAmountUsed > 0 {
refID := order.ID
if err := utils.CreditWallet(tx, order.UserID, order.WalletAmountUsed, models.WalletReasonOrderRefund, "order", &refID, "Refund for cancelled order"); err != nil {
return err
}
}

var orderCoupon models.OrderCoupon
if err := tx.Where("order_id = ?", order.ID).First(&orderCoupon).Error; err == nil {
if err := tx.Model(&models.Coupon{}).Where("id = ? AND used_count > 0", orderCoupon.CouponID).
UpdateColumn("used_count", gorm.Expr("used_count - 1")).Error; err != nil {
return err
}
}

if order.PaymentMethod == models.PaymentMethodOnline {
if err := tx.Where("order_id = ?", order.ID).First(&payment).Error; err == nil {
if payment.Status == models.PaymentStatusPaid {
gatewayRefundAmount = payment.Amount - payment.RefundedAmount
if gatewayRefundAmount > 0 {
refID := order.ID
if err := utils.CreditWallet(tx, order.UserID, gatewayRefundAmount, models.WalletReasonOrderRefund, "order", &refID, "Gateway refund for cancelled order"); err != nil {
return err
}
payment.RefundedAmount += gatewayRefundAmount
payment.Status = models.PaymentStatusRefunded
if err := tx.Save(&payment).Error; err != nil {
return err
}
}
}
}
}

// Cancel any in-flight picking/packing task so the warehouse UI
// doesn't keep showing dead work for a cancelled order.
if err := tx.Model(&models.PickingTask{}).Where("order_id = ? AND status != ?", order.ID, "completed").
Update("status", "cancelled").Error; err != nil {
return err
}
if err := tx.Model(&models.PackingTask{}).Where("order_id = ? AND status != ?", order.ID, "completed").
Update("status", "cancelled").Error; err != nil {
return err
}

// Unassign any delivery partner already attached - auto-assign
// can fire as early as order confirmation, well before this
// fulfillment-stage cancel window closes.
if order.DeliveryPartnerID != nil {
previouslyAssignedPartnerID = order.DeliveryPartnerID
if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
"delivery_partner_id":            nil,
"delivery_assignment_status":     nil,
"delivery_assignment_expires_at": nil,
"delivery_status":                nil,
"assigned_at":                    nil,
}).Error; err != nil {
return err
}
}

now := time.Now()
return tx.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
"status":              models.OrderStatusCancelled,
"cancellation_reason": reason,
"cancelled_by_type":   actorType,
"cancelled_by_id":     actorID,
"cancelled_at":        now,
}).Error
})

if txErr != nil {
return txErr
}

if previouslyAssignedPartnerID != nil {
CreateDeliveryNotification(*previouslyAssignedPartnerID, "Order cancelled",
fmt.Sprintf("Order #%d was cancelled by the warehouse and is no longer assigned to you.", order.ID),
"order_cancelled", &order.ID)
}
if order.WarehouseID != nil {
NotifyWarehouse(*order.WarehouseID, "order_cancelled",
"Order #"+fmt.Sprint(order.ID)+" cancelled",
fmt.Sprintf("Cancelled by %s: %s", actorType, reason), &order.ID, nil)
}

return nil
}
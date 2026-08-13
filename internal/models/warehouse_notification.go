package models

import "time"

// Warehouse notification types - operational events staff should see.
const (
WhNotifyNewOrder         = "new_order"
WhNotifyUrgentOrder      = "urgent_order"
WhNotifyOrderCancelled   = "order_cancelled"
WhNotifyLowStock         = "low_stock"
WhNotifyOutOfStock       = "out_of_stock"
WhNotifyExpiryAlert      = "expiry_alert"
WhNotifyStockTransfer    = "stock_transfer"
WhNotifyReceiving        = "receiving"
WhNotifyHandoverRequired = "handover_required"
WhNotifyExceptionCreated = "exception_created"
)

// WarehouseNotification is an in-app, warehouse-scoped operational alert -
// distinct from models.Notification, which is the customer-facing SMS log.
// Reuses the same "write a record, list it" shape rather than inventing a
// new delivery mechanism; a push/SMS channel can be layered on top later by
// having NotifyWarehouse also call utils.SendNotification if ever needed.
type WarehouseNotification struct {
ID          uint       `gorm:"primaryKey" json:"id"`
WarehouseID uint       `gorm:"not null;index" json:"warehouse_id"`
Type        string     `gorm:"not null;index" json:"type"`
Title       string     `gorm:"not null" json:"title"`
Message     string     `json:"message"`
OrderID     *uint      `json:"order_id,omitempty"`
ProductID   *uint      `json:"product_id,omitempty"`
IsRead      bool       `gorm:"default:false;index" json:"is_read"`
ReadAt      *time.Time `json:"read_at,omitempty"`
CreatedAt   time.Time  `gorm:"index" json:"created_at"`
}

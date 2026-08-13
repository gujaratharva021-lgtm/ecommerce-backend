package services

import (
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// NotifyWarehouse writes one in-app operational notification for every
// staff member at warehouseID to see on their next fetch. Fire-and-forget,
// same as LogWarehouseAction - never block the action that triggered it.
func NotifyWarehouse(warehouseID uint, notifType, title, message string, orderID, productID *uint) {
n := models.WarehouseNotification{
WarehouseID: warehouseID,
Type:        notifType,
Title:       title,
Message:     message,
OrderID:     orderID,
ProductID:   productID,
}
database.DB.Create(&n)
}

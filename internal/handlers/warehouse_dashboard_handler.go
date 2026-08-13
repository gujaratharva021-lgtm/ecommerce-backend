package handlers

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetWarehouseDashboard godoc
// GET /api/v1/warehouse/dashboard (warehouse staff only)
func GetWarehouseDashboard(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
todayStart := time.Now().Truncate(24 * time.Hour)

var todayOrders, newOrders, picking, packed, readyForDispatch, completed, cancelled int64
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND created_at >= ?", warehouseID, todayStart).Count(&todayOrders)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status = ?", warehouseID, models.OrderStatusConfirmed).Count(&newOrders)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status IN ?", warehouseID, []string{models.OrderStatusPicking, models.OrderStatusPicked}).Count(&picking)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status IN ?", warehouseID, []string{models.OrderStatusPacking, models.OrderStatusPacked}).Count(&packed)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status = ?", warehouseID, models.OrderStatusReadyForDispatch).Count(&readyForDispatch)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status = ? AND created_at >= ?", warehouseID, models.OrderStatusDelivered, todayStart).Count(&completed)
database.DB.Model(&models.Order{}).Where("warehouse_id = ? AND status = ? AND created_at >= ?", warehouseID, models.OrderStatusCancelled, todayStart).Count(&cancelled)

var lowStock, outOfStock int64
database.DB.Model(&models.Inventory{}).Where("warehouse_id = ? AND stock > 0 AND stock < ?", warehouseID, lowStockThreshold).Count(&lowStock)
database.DB.Model(&models.Inventory{}).Where("warehouse_id = ? AND stock <= 0", warehouseID).Count(&outOfStock)

var pendingTransfers int64
database.DB.Model(&models.StockTransfer{}).
Where("(from_warehouse_id = ? OR to_warehouse_id = ?) AND status = ?", warehouseID, warehouseID, "pending").
Count(&pendingTransfers)

var activeStaff int64
database.DB.Model(&models.WarehouseStaff{}).Where("warehouse_id = ? AND is_active = ?", warehouseID, true).Count(&activeStaff)

// Average picking time (minutes) for tasks completed today.
var avgPickingSeconds float64
database.DB.Model(&models.PickingTask{}).
Where("warehouse_id = ? AND status = ? AND completed_at IS NOT NULL AND started_at IS NOT NULL AND completed_at >= ?", warehouseID, "completed", todayStart).
Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0)").
Scan(&avgPickingSeconds)

var avgPackingSeconds float64
database.DB.Model(&models.PackingTask{}).
Where("warehouse_id = ? AND status = ? AND completed_at IS NOT NULL AND started_at IS NOT NULL AND completed_at >= ?", warehouseID, "completed", todayStart).
Select("COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - started_at))), 0)").
Scan(&avgPackingSeconds)

var fulfillmentRate float64
if todayOrders > 0 {
fulfillmentRate = float64(completed) / float64(todayOrders) * 100
}

c.JSON(http.StatusOK, gin.H{
"today_orders":         todayOrders,
"new_orders":           newOrders,
"picking":              picking,
"packed":               packed,
"ready_for_dispatch":   readyForDispatch,
"completed_today":      completed,
"cancelled_today":      cancelled,
"low_stock_products":   lowStock,
"out_of_stock_products": outOfStock,
"pending_stock_transfers": pendingTransfers,
"active_staff":         activeStaff,
"avg_picking_minutes":  avgPickingSeconds / 60,
"avg_packing_minutes":  avgPackingSeconds / 60,
"fulfillment_rate":     fulfillmentRate,
})
}

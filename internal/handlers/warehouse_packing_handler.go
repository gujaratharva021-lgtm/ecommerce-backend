package handlers

import (
"fmt"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// GetPackingTask godoc
// GET /api/v1/warehouse/packing/:order_id (warehouse staff only)
func GetPackingTask(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("order_id")

var task models.PackingTask
if err := database.DB.Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).
Preload("Order.Address").First(&task).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Packing task not found for your warehouse"})
return
}

// Include the picking outcome so the packer can see what was actually
// picked (including any short/unavailable items) before packing.
var pickingTask models.PickingTask
database.DB.Where("order_id = ?", orderID).Preload("Items.Product").First(&pickingTask)

c.JSON(http.StatusOK, gin.H{
"packing_task":  task,
"picked_items":  pickingTask.Items,
})
}

// StartPacking godoc
// PUT /api/v1/warehouse/packing/:order_id/start (warehouse staff only)
func StartPacking(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
orderID := c.Param("order_id")

var task models.PackingTask
if err := database.DB.Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).First(&task).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Packing task not found for your warehouse"})
return
}
if task.Status == "completed" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Packing already completed for this order"})
return
}
if task.PackerID != nil && *task.PackerID != staffID {
c.JSON(http.StatusConflict, gin.H{"error": "This order is already being packed by another staff member"})
return
}

now := time.Now()
task.PackerID = &staffID
task.Status = "in_progress"
if task.StartedAt == nil {
task.StartedAt = &now
}
database.DB.Save(&task)

c.JSON(http.StatusOK, task)
}

// CompletePacking godoc
// PUT /api/v1/warehouse/packing/:order_id/complete (warehouse staff only)
// Marks the order packed and ready for dispatch. Prevents double-packing by
// checking task.Status == "completed" up front (idempotency guard).
func CompletePacking(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
orderID := c.Param("order_id")

var task models.PackingTask
if err := database.DB.Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).First(&task).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Packing task not found for your warehouse"})
return
}
if task.Status == "completed" {
c.JSON(http.StatusConflict, gin.H{"error": "This order has already been packed"})
return
}
if task.Status != "in_progress" {
c.JSON(http.StatusBadRequest, gin.H{"error": "Packing must be started before it can be completed"})
return
}

now := time.Now()
task.Status = "completed"
task.CompletedAt = &now
database.DB.Save(&task)

database.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("status", models.OrderStatusReadyForDispatch)

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "complete_packing", "order", orderID,
"status=packing", "status=ready_for_dispatch")

c.JSON(http.StatusOK, gin.H{"success": true, "packing_task": task, "order_status": models.OrderStatusReadyForDispatch})
}

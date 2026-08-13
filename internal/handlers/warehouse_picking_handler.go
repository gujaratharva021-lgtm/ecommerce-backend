package handlers

import (
"fmt"
	"errors"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetPickingTask godoc
// GET /api/v1/warehouse/picking/:order_id (warehouse staff only)
func GetPickingTask(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("order_id")

var task models.PickingTask
if err := database.DB.Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).
Preload("Items.Product").Preload("Order.Address").First(&task).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Picking task not found for your warehouse"})
return
}

// Attach bin location for each item so the picker can see where to go.
productIDs := make([]uint, 0, len(task.Items))
for _, item := range task.Items {
productIDs = append(productIDs, item.ProductID)
}
var invRows []models.Inventory
database.DB.Where("warehouse_id = ? AND product_id IN ?", warehouseID, productIDs).
Preload("Bin.Rack.Zone").Find(&invRows)
locationByProduct := make(map[uint]models.Inventory)
for _, inv := range invRows {
locationByProduct[inv.ProductID] = inv
}

type PickingTaskItemWithLocation struct {
models.PickingTaskItem
Location *models.Inventory `json:"location,omitempty"`
}
itemsWithLocation := make([]PickingTaskItemWithLocation, 0, len(task.Items))
for _, item := range task.Items {
entry := PickingTaskItemWithLocation{PickingTaskItem: item}
if inv, ok := locationByProduct[item.ProductID]; ok && inv.Bin != nil {
entry.Location = &inv
}
itemsWithLocation = append(itemsWithLocation, entry)
}

c.JSON(http.StatusOK, gin.H{
"id":           task.ID,
"order_id":     task.OrderID,
"order":        task.Order,
"warehouse_id": task.WarehouseID,
"picker_id":    task.PickerID,
"status":       task.Status,
"started_at":   task.StartedAt,
"completed_at": task.CompletedAt,
"created_at":   task.CreatedAt,
"updated_at":   task.UpdatedAt,
"items":        itemsWithLocation,
})
}

// StartPicking godoc
// PUT /api/v1/warehouse/picking/:order_id/start (warehouse staff only)
func StartPicking(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
orderID := c.Param("order_id")

var task models.PickingTask
statusCode := http.StatusInternalServerError

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).
First(&task).Error; err != nil {
statusCode = http.StatusNotFound
return errors.New("Picking task not found for your warehouse")
}
if task.Status == "completed" {
statusCode = http.StatusBadRequest
return errors.New("Picking already completed for this order")
}
if task.PickerID != nil && *task.PickerID != staffID {
statusCode = http.StatusConflict
return errors.New("This order is already being picked by another staff member")
}

now := time.Now()
task.PickerID = &staffID
task.Status = "in_progress"
if task.StartedAt == nil {
task.StartedAt = &now
}
return tx.Save(&task).Error
})

if txErr != nil {
c.JSON(statusCode, gin.H{"error": txErr.Error()})
return
}

c.JSON(http.StatusOK, task)
}

// PickItemRequest is the body for PUT /warehouse/picking/items/:item_id
type PickItemRequest struct {
Status         string `json:"status" binding:"required,oneof=picked unavailable short"`
QuantityPicked int    `json:"quantity_picked"`
Reason         string `json:"reason"`
}

// MarkPickItem godoc
// PUT /api/v1/warehouse/picking/items/:item_id (warehouse staff only)
// Marks one order-item line as picked/unavailable/short. Barcode verification
// happens client-side (scan resolves to a product_id/SKU which must match
// this item's product before the app calls this endpoint) - the backend's
// job is to record the outcome, not to re-verify hardware input.
func MarkPickItem(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
itemID := c.Param("item_id")

var req PickItemRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var item models.PickingTaskItem
if err := database.DB.First(&item, itemID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Picking item not found"})
return
}

// Verify this item's picking task belongs to the caller's warehouse.
var task models.PickingTask
if err := database.DB.Where("id = ? AND warehouse_id = ?", item.PickingTaskID, warehouseID).First(&task).Error; err != nil {
c.JSON(http.StatusForbidden, gin.H{"error": "This item does not belong to your warehouse"})
return
}

switch req.Status {
case models.PickItemPicked:
item.QuantityPicked = item.QuantityNeeded
case models.PickItemShort:
if req.QuantityPicked <= 0 || req.QuantityPicked >= item.QuantityNeeded {
c.JSON(http.StatusBadRequest, gin.H{"error": "quantity_picked must be between 1 and quantity_needed-1 for a short pick"})
return
}
item.QuantityPicked = req.QuantityPicked
case models.PickItemUnavailable:
item.QuantityPicked = 0
}
item.Status = req.Status
item.Reason = req.Reason
database.DB.Save(&item)

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "mark_pick_item", "picking_task_item", itemID,
"status=pending", fmt.Sprintf("status=%s qty_picked=%d", item.Status, item.QuantityPicked))

// Auto-create a WarehouseException for unavailable/short picks so staff
// don't have to double-enter what they already reported inline here.
if req.Status == models.PickItemUnavailable || req.Status == models.PickItemShort {
productID := item.ProductID
exceptionType := models.ExceptionUnavailable
priority := models.ExceptionPriorityMedium
if req.Status == models.PickItemShort {
exceptionType = models.ExceptionShortQuantity
priority = models.ExceptionPriorityLow
}
exception := models.WarehouseException{
OrderID:     task.OrderID,
ProductID:   &productID,
WarehouseID: warehouseID,
Type:        exceptionType,
Reason:      req.Reason,
Priority:    priority,
StaffID:     &staffID,
Status:      models.ExceptionStatusOpen,
}
database.DB.Create(&exception)
}

c.JSON(http.StatusOK, item)
}

// CompletePicking godoc
// PUT /api/v1/warehouse/picking/:order_id/complete (warehouse staff only)
// Finalizes picking and creates the PackingTask. Requires every item to
// have been marked (picked/unavailable/short) - no item left pending.
func CompletePicking(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("order_id")
staffID := c.MustGet("staff_id").(uint)

var task models.PickingTask
var packTask models.PackingTask
statusCode := http.StatusInternalServerError

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).
Preload("Items").First(&task).Error; err != nil {
statusCode = http.StatusNotFound
return errors.New("Picking task not found for your warehouse")
}
if task.Status == "completed" {
statusCode = http.StatusBadRequest
return errors.New("Picking already completed")
}

for _, item := range task.Items {
if item.Status == models.PickItemPending {
statusCode = http.StatusBadRequest
return errors.New("Cannot complete picking - not all items have been marked")
}
}

now := time.Now()
task.Status = "completed"
task.CompletedAt = &now
if err := tx.Save(&task).Error; err != nil {
return err
}

packTask = models.PackingTask{
OrderID:     task.OrderID,
WarehouseID: warehouseID,
Status:      "pending",
}
if err := tx.Create(&packTask).Error; err != nil {
return err
}

return tx.Model(&models.Order{}).Where("id = ?", task.OrderID).Update("status", models.OrderStatusPicked).Error
})

if txErr != nil {
c.JSON(statusCode, gin.H{"error": txErr.Error()})
return
}

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "complete_picking", "order", orderID,
"status=picking", "status=picked")

c.JSON(http.StatusOK, gin.H{"success": true, "picking_task": task, "packing_task": packTask})
}

package handlers

import (
"fmt"
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// CancelOrderInFulfillmentRequest is the body for both
// POST /admin/orders/:id/cancel and POST /warehouse/orders/:id/cancel.
type CancelOrderInFulfillmentRequest struct {
Reason string `json:"reason" binding:"required"`
}

// AdminCancelOrder godoc
// POST /api/v1/admin/orders/:id/cancel (admin only)
// Cancels an order that has progressed past the plain customer-cancellable
// window (confirmed through ready_for_dispatch), restoring stock and
// reversing wallet/coupon/payment same as CancelOrderInFulfillment.
func AdminCancelOrder(c *gin.Context) {
adminID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var req CancelOrderInFulfillmentRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var order models.Order
if err := database.DB.First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
return
}

if err := services.CancelOrderInFulfillment(order.ID, "admin", adminID, "", req.Reason); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"success": true, "order_id": order.ID, "status": models.OrderStatusCancelled})
}

// WarehouseCancelOrder godoc
// POST /api/v1/warehouse/orders/:id/cancel (InventoryManagerOnly)
// Same as AdminCancelOrder, but scoped to the caller's own warehouse and
// restricted to inventory-manager-tier warehouse staff via route
// middleware, mirroring the pattern used by ReassignPicking/ReassignPacking.
func WarehouseCancelOrder(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
staffName, _ := c.Get("staff_name")
orderID := c.Param("id")

var req CancelOrderInFulfillmentRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var order models.Order
if err := database.DB.Where("warehouse_id = ?", warehouseID).First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
return
}

if err := services.CancelOrderInFulfillment(order.ID, "inventory_manager", staffID, fmt.Sprint(staffName), req.Reason); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"success": true, "order_id": order.ID, "status": models.OrderStatusCancelled})
}
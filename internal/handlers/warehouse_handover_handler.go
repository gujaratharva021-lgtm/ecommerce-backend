package handlers

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
)

// HandoverOrderRequest is the body for PUT /warehouse/orders/:id/handover
type HandoverOrderRequest struct {
PackageCount int `json:"package_count" binding:"required,min=1"`
}

// HandoverOrder godoc
// PUT /api/v1/warehouse/orders/:id/handover (warehouse staff only)
// Verifies a ready_for_dispatch order has an assigned delivery partner,
// records the handover, and moves the order to handed_over. This is what
// unblocks the delivery partner's own "mark shipped" action.
func HandoverOrder(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
orderID := c.Param("id")

var req HandoverOrderRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var order models.Order
if err := database.DB.Where("warehouse_id = ?", warehouseID).First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
return
}

if order.Status != models.OrderStatusReadyForDispatch {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only ready_for_dispatch orders can be handed over, current status: " + order.Status})
return
}

if order.DeliveryPartnerID == nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "No delivery partner assigned to this order yet"})
return
}

var existing models.OrderHandover
if err := database.DB.Where("order_id = ?", order.ID).First(&existing).Error; err == nil {
c.JSON(http.StatusConflict, gin.H{"error": "This order has already been handed over"})
return
}

now := time.Now()
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
handover := models.OrderHandover{
OrderID:           order.ID,
WarehouseID:       warehouseID,
WarehouseStaffID:  staffID,
DeliveryPartnerID: *order.DeliveryPartnerID,
PackageCount:      req.PackageCount,
HandedOverAt:      now,
}
if err := tx.Create(&handover).Error; err != nil {
return err
}
order.Status = models.OrderStatusHandedOver
return tx.Save(&order).Error
})

if txErr != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record handover: " + txErr.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"success": true, "order_id": order.ID, "status": order.Status})
}

// GetHandover godoc
// GET /api/v1/warehouse/orders/:id/handover (warehouse staff only)
func GetHandover(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("id")

var handover models.OrderHandover
if err := database.DB.Where("order_id = ? AND warehouse_id = ?", orderID, warehouseID).First(&handover).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No handover record for this order"})
return
}
c.JSON(http.StatusOK, handover)
}

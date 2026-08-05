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

// ---------------------------------------------------------------------------
// Delivery Partners (admin only)
// ---------------------------------------------------------------------------

// CreateDeliveryPartner godoc
// POST /api/v1/admin/delivery-partners (admin only)
func CreateDeliveryPartner(c *gin.Context) {
var req models.DeliveryPartnerRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

partner := models.DeliveryPartner{
Name:          req.Name,
Phone:         req.Phone,
VehicleNumber: req.VehicleNumber,
IsActive:      true,
}
if req.IsActive != nil {
partner.IsActive = *req.IsActive
}

if err := database.DB.Create(&partner).Error; err != nil {
c.JSON(http.StatusConflict, gin.H{"error": "Delivery partner already exists or could not be created"})
return
}

c.JSON(http.StatusCreated, gin.H{"delivery_partner": partner})
}

// GetDeliveryPartners godoc
// GET /api/v1/admin/delivery-partners (admin only)
func GetDeliveryPartners(c *gin.Context) {
var partners []models.DeliveryPartner
if err := database.DB.Order("created_at DESC").Find(&partners).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load delivery partners"})
return
}
c.JSON(http.StatusOK, gin.H{"delivery_partners": partners})
}

// UpdateDeliveryPartner godoc
// PUT /api/v1/admin/delivery-partners/:id (admin only)
func UpdateDeliveryPartner(c *gin.Context) {
id := c.Param("id")

var partner models.DeliveryPartner
if err := database.DB.First(&partner, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Delivery partner not found"})
return
}

var req models.DeliveryPartnerRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

partner.Name = req.Name
partner.Phone = req.Phone
partner.VehicleNumber = req.VehicleNumber
if req.IsActive != nil {
partner.IsActive = *req.IsActive
}

if err := database.DB.Save(&partner).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update delivery partner"})
return
}

c.JSON(http.StatusOK, gin.H{"delivery_partner": partner})
}

// DeleteDeliveryPartner godoc
// DELETE /api/v1/admin/delivery-partners/:id (admin only)
func DeleteDeliveryPartner(c *gin.Context) {
id := c.Param("id")

result := database.DB.Delete(&models.DeliveryPartner{}, id)
if result.Error != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete delivery partner"})
return
}
if result.RowsAffected == 0 {
c.JSON(http.StatusNotFound, gin.H{"error": "Delivery partner not found"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Delivery partner deleted"})
}

// ---------------------------------------------------------------------------
// Order assignment (admin only)
// ---------------------------------------------------------------------------

// AssignDeliveryPartner godoc
// PUT /api/v1/admin/orders/:id/assign-delivery (admin only)
func AssignDeliveryPartner(c *gin.Context) {
orderID := c.Param("id")

var order models.Order
if err := database.DB.First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
return
}

if order.Status != models.OrderStatusConfirmed && order.Status != models.OrderStatusShipped {
c.JSON(http.StatusBadRequest, gin.H{"error": "Delivery partner can only be assigned to confirmed or shipped orders"})
return
}

var req models.AssignDeliveryPartnerRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var partner models.DeliveryPartner
if err := database.DB.First(&partner, req.DeliveryPartnerID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Delivery partner not found"})
return
}
if !partner.IsActive {
c.JSON(http.StatusBadRequest, gin.H{"error": "Delivery partner is not active"})
return
}

order.DeliveryPartnerID = &req.DeliveryPartnerID
if err := database.DB.Save(&order).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign delivery partner"})
return
}

database.DB.Preload("DeliveryPartner").First(&order, order.ID)

// Notify the partner's device(s) that a new order has been assigned.
go services.SendPushToPartner(
req.DeliveryPartnerID,
"New delivery assigned",
fmt.Sprintf("Order #%d has been assigned to you", order.ID),
)

c.JSON(http.StatusOK, gin.H{"message": "Delivery partner assigned", "order": order})
}

// ---------------------------------------------------------------------------
// Delivery Partner order actions (delivery partner only)
// ---------------------------------------------------------------------------

// GetMyDeliveries godoc
// GET /api/v1/delivery/orders (delivery partner only)
// Returns orders assigned to the logged-in delivery partner.
// Optional ?status= filter (e.g. confirmed, shipped, delivered).
func GetMyDeliveries(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)

query := database.DB.
Preload("Address").
Preload("Items.Product").
Where("delivery_partner_id = ?", partnerID)

if status := c.Query("status"); status != "" {
query = query.Where("status = ?", status)
}

var orders []models.Order
if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load assigned orders"})
return
}

c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// UpdateDeliveryOrderStatus godoc
// PUT /api/v1/delivery/orders/:id/status (delivery partner only)
// Lets the assigned partner move an order from confirmed -> shipped
// (picked up and out for delivery). Partners cannot set "delivered"
// here - see ConfirmDelivery below, which also handles COD collection.
func UpdateDeliveryOrderStatus(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var req struct {
Status string `json:"status" binding:"required,oneof=shipped"`
}
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var order models.Order
if err := database.DB.Preload("Address").Preload("Items.Product").Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or not assigned to you"})
return
}

if order.Status != models.OrderStatusConfirmed {
c.JSON(http.StatusBadRequest, gin.H{"error": "Order must be confirmed before it can be marked shipped"})
return
}

order.Status = req.Status
if err := database.DB.Save(&order).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
return
}

// Notify the customer that their order is out for delivery.
go services.SendPushToUser(
order.UserID,
"Order out for delivery",
fmt.Sprintf("Your order #%d is on its way", order.ID),
)

c.JSON(http.StatusOK, gin.H{"message": "Order status updated", "order": order})
}

// ConfirmDelivery godoc
// PUT /api/v1/delivery/orders/:id/deliver (delivery partner only)
// Marks the order delivered. For COD orders this also marks payment as
// paid, since cash is collected at the same time.
func ConfirmDelivery(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Preload("Address").Preload("Items.Product").Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or not assigned to you"})
return
}

if order.Status != models.OrderStatusShipped {
c.JSON(http.StatusBadRequest, gin.H{"error": "Order must be shipped before it can be marked delivered"})
return
}

order.Status = models.OrderStatusDelivered
if order.PaymentMethod == models.PaymentMethodCOD {
order.PaymentStatus = models.OrderPaymentStatusPaid
}

if err := database.DB.Save(&order).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm delivery"})
return
}

// Notify the customer that their order has been delivered.
go services.SendPushToUser(
order.UserID,
"Order delivered",
fmt.Sprintf("Your order #%d has been delivered. Enjoy!", order.ID),
)

c.JSON(http.StatusOK, gin.H{"message": "Delivery confirmed", "order": order})
}

// GetMyEarnings godoc
// GET /api/v1/delivery/earnings (delivery partner only)
// Returns delivery earnings summary: today's earnings, total earnings,
// total delivered count, and a list of delivered orders with the flat
// per-delivery payout. There is no separate earnings/payout table yet,
// so this is computed on the fly at a fixed rate per delivered order.
const perDeliveryEarning = 30.0

func GetMyEarnings(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)

var deliveredOrders []models.Order
if err := database.DB.
Where("delivery_partner_id = ? AND status = ?", partnerID, models.OrderStatusDelivered).
Order("updated_at DESC").
Find(&deliveredOrders).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load earnings"})
return
}

totalEarnings := float64(len(deliveredOrders)) * perDeliveryEarning

todayStart := time.Now().Truncate(24 * time.Hour)
todayCount := 0
todayEarnings := 0.0
type EarningEntry struct {
OrderID     uint      `json:"order_id"`
Amount      float64   `json:"amount"`
DeliveredAt time.Time `json:"delivered_at"`
}
entries := make([]EarningEntry, 0, len(deliveredOrders))

for _, o := range deliveredOrders {
if o.UpdatedAt.After(todayStart) {
todayCount++
todayEarnings += perDeliveryEarning
}
entries = append(entries, EarningEntry{
OrderID:     o.ID,
Amount:      perDeliveryEarning,
DeliveredAt: o.UpdatedAt,
})
}

c.JSON(http.StatusOK, gin.H{
"per_delivery_rate": perDeliveryEarning,
"today_earnings":    todayEarnings,
"today_deliveries":  todayCount,
"total_earnings":    totalEarnings,
"total_deliveries":  len(deliveredOrders),
"entries":           entries,
})
}

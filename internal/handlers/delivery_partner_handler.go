package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
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
c.JSON(http.StatusOK, gin.H{"message": "Delivery partner assigned", "order": order})
}

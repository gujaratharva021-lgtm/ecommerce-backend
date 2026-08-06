package handlers

import (
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// SendPartnerOTP godoc
// POST /api/v1/delivery/send-otp
// Sends an OTP to a phone number that belongs to an existing, active
// delivery partner (created by the admin). Unlike customer signup, this
// does NOT create a new partner — the admin must add them first.
func SendPartnerOTP(c *gin.Context) {
var req models.SendOTPRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var partner models.DeliveryPartner
if err := database.DB.Where("phone = ?", req.Phone).First(&partner).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No delivery partner found with this phone number"})
return
}
if !partner.IsActive {
c.JSON(http.StatusForbidden, gin.H{"error": "This delivery partner account is inactive"})
return
}

code, err := generateOTP()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
return
}

database.DB.Where("phone = ? AND verified = ?", req.Phone, false).Delete(&models.OTP{})
otp := models.OTP{
Phone:     req.Phone,
Code:      code,
ExpiresAt: time.Now().Add(otpValidityMinutes * time.Minute),
Verified:  false,
}
if err := database.DB.Create(&otp).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OTP"})
return
}

resp := gin.H{
"message":            "OTP sent successfully",
"expires_in_minutes": otpValidityMinutes,
}
if config.AppConfig.GinMode != "release" {
resp["otp"] = code
}
c.JSON(http.StatusOK, resp)
}

// VerifyPartnerOTP godoc
// POST /api/v1/delivery/verify-otp
// Verifies the OTP and returns a JWT with role "delivery_partner".
func VerifyPartnerOTP(c *gin.Context) {
var req models.VerifyOTPRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var otp models.OTP
err := database.DB.
Where("phone = ? AND code = ? AND verified = ?", req.Phone, req.OTP, false).
Order("created_at DESC").
First(&otp).Error
if err != nil {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
return
}
if time.Now().After(otp.ExpiresAt) {
c.JSON(http.StatusUnauthorized, gin.H{"error": "OTP has expired, please request a new one"})
return
}
database.DB.Model(&otp).Update("verified", true)

var partner models.DeliveryPartner
if err := database.DB.Where("phone = ?", req.Phone).First(&partner).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Delivery partner not found"})
return
}

token, err := utils.GenerateJWT(partner.ID, partner.Phone, "delivery_partner")
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
return
}

c.JSON(http.StatusOK, gin.H{"token": token, "delivery_partner": partner})
}

// UpdateLocation godoc
// PUT /api/v1/delivery/location (delivery partner only)
// The partner's app calls this periodically (e.g. every 15-30s) while
// out for delivery, to push their current GPS coordinates.
func UpdateLocation(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)

var req models.UpdateLocationRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

now := time.Now()
result := database.DB.Model(&models.DeliveryPartner{}).
Where("id = ?", partnerID).
Updates(map[string]interface{}{
"current_lat":          req.Lat,
"current_lng":          req.Lng,
"last_location_update": now,
})

if result.Error != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Location updated"})
}

// GetOrderTracking godoc
// GET /api/v1/orders/:id/tracking (protected — order owner only)
// Returns the assigned delivery partner's current live location, so the
// customer app can show them on a map.
func GetOrderTracking(c *gin.Context) {
userID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.
Preload("DeliveryPartner").
Where("id = ? AND user_id = ?", orderID, userID).
First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
return
}

if order.DeliveryPartner == nil {
c.JSON(http.StatusOK, gin.H{"message": "No delivery partner assigned yet", "tracking": nil})
return
}

c.JSON(http.StatusOK, gin.H{
"tracking": gin.H{
"delivery_partner_name": order.DeliveryPartner.Name,
"vehicle_number":        order.DeliveryPartner.VehicleNumber,
"phone":                 order.DeliveryPartner.Phone,
"current_lat":           order.DeliveryPartner.CurrentLat,
"current_lng":           order.DeliveryPartner.CurrentLng,
"last_updated":          order.DeliveryPartner.LastLocationUpdate,
"order_status":          order.Status,
},
})
}

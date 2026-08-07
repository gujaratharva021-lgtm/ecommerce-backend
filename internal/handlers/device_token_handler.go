package handlers

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

type RegisterDeviceTokenRequest struct {
Token    string `json:"token" binding:"required"`
Platform string `json:"platform"`
}

func RegisterDeviceToken(c *gin.Context) {
var req RegisterDeviceTokenRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var userID *uint
var deliveryPartnerID *uint
authHeader := c.GetHeader("Authorization")
if authHeader != "" {
parts := strings.Split(authHeader, " ")
if len(parts) == 2 && parts[0] == "Bearer" {
if claims, err := utils.ValidateJWT(parts[1]); err == nil {
id := claims.UserID
if claims.Role == "delivery_partner" {
deliveryPartnerID = &id
} else {
userID = &id
}
}
}
}

var existing models.DeviceToken
result := database.DB.Where("token = ?", req.Token).First(&existing)
if result.Error == nil {
existing.Platform = req.Platform
if userID != nil {
existing.UserID = userID
}
if deliveryPartnerID != nil {
existing.DeliveryPartnerID = deliveryPartnerID
}
database.DB.Save(&existing)
c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
return
}

token := models.DeviceToken{
Token:             req.Token,
Platform:          req.Platform,
UserID:            userID,
DeliveryPartnerID: deliveryPartnerID,
}
if err := database.DB.Create(&token).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save device token"})
return
}
c.JSON(http.StatusCreated, gin.H{"message": "Token registered"})
}

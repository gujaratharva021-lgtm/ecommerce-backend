package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// RegisterDeviceTokenRequest is the payload sent by the Flutter app when
// registering (or re-registering) its FCM push token.
type RegisterDeviceTokenRequest struct {
Token    string `json:"token" binding:"required"`
Platform string `json:"platform"`
}

// RegisterDeviceToken godoc
// POST /api/v1/device-token (public — app may call this before login)
// Upserts the token: if it already exists, just refreshes updated_at.
func RegisterDeviceToken(c *gin.Context) {
var req RegisterDeviceTokenRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var existing models.DeviceToken
result := database.DB.Where("token = ?", req.Token).First(&existing)

if result.Error == nil {
existing.Platform = req.Platform
database.DB.Save(&existing)
c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
return
}

token := models.DeviceToken{
Token:    req.Token,
Platform: req.Platform,
}
if err := database.DB.Create(&token).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save device token"})
return
}

c.JSON(http.StatusCreated, gin.H{"message": "Token registered"})
}

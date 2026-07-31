package handlers

import (
"net/http"
"strings"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
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
// If an "Authorization: Bearer <token>" header is present and valid, the
// token is linked to that user so order-status pushes can target them
// specifically. If not (e.g. called before login), it's saved unlinked.
func RegisterDeviceToken(c *gin.Context) {
var req RegisterDeviceTokenRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var userID *uint
authHeader := c.GetHeader("Authorization")
if authHeader != "" {
parts := strings.Split(authHeader, " ")
if len(parts) == 2 && parts[0] == "Bearer" {
if claims, err := utils.ValidateJWT(parts[1]); err == nil {
id := claims.UserID
userID = &id
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
database.DB.Save(&existing)
c.JSON(http.StatusOK, gin.H{"message": "Token refreshed"})
return
}

token := models.DeviceToken{
Token:    req.Token,
Platform: req.Platform,
UserID:   userID,
}
if err := database.DB.Create(&token).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save device token"})
return
}

c.JSON(http.StatusCreated, gin.H{"message": "Token registered"})
}

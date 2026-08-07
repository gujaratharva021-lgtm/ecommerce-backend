package handlers

import (
"crypto/rand"
"fmt"
"math/big"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

const otpValidityMinutes = 5

// generateOTP returns a random 6-digit numeric code, e.g. "042917".
func generateOTP() (string, error) {
n, err := rand.Int(rand.Reader, big.NewInt(1000000))
if err != nil {
return "", err
}
return fmt.Sprintf("%06d", n.Int64()), nil
}

// SendOTP godoc
// POST /api/v1/auth/send-otp
// TEST MODE: no real SMS is sent. A 6-digit OTP is generated, stored locally
// (old codes for the same phone are cleared first), and returned directly in
// the response so the frontend can display/auto-fill it. Swap this back to a
// real SMS provider before shipping to real users -- returning the code in
// the API response is only safe for internal testing.
func SendOTP(c *gin.Context) {
var req models.SendOTPRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

code, err := generateOTP()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
return
}

// Clear any older codes for this phone, then store the fresh one.
database.DB.Where("phone = ?", req.Phone).Delete(&models.OTP{})
otp := models.OTP{
Phone:     req.Phone,
Code:      code,
ExpiresAt: time.Now().Add(otpValidityMinutes * time.Minute),
}
if err := database.DB.Create(&otp).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OTP"})
return
}

c.JSON(http.StatusOK, gin.H{
"message": "OTP sent successfully",
"otp":     code, // TEST MODE ONLY -- remove this field once real SMS is wired up
})
}

// VerifyOTP godoc
// POST /api/v1/auth/verify-otp
// TEST MODE: checks the code against the local OTP table instead of Twilio.
// Creates the user on first login, logs them in on repeat visits.
func VerifyOTP(c *gin.Context) {
var req models.VerifyOTPRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var otp models.OTP
err := database.DB.Where("phone = ? AND code = ?", req.Phone, req.OTP).
Order("id desc").First(&otp).Error
if err != nil || time.Now().After(otp.ExpiresAt) {
c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
return
}
database.DB.Delete(&otp)

// Find or create the user -- first successful OTP verification = signup.
var user models.User
err = database.DB.Where("phone = ?", req.Phone).First(&user).Error
if err != nil {
user = models.User{
Phone: req.Phone,
Role:  "customer",
}
if err := database.DB.Create(&user).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
return
}
// Auto-create an empty cart for the new user
cart := models.Cart{UserID: user.ID}
database.DB.Create(&cart)
}

token, err := utils.GenerateJWT(user.ID, user.Phone, user.Role)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
return
}

c.JSON(http.StatusOK, models.AuthResponse{
Token: token,
User:  user,
})
}

// Me godoc - returns the logged-in user's profile
// GET /api/v1/auth/me (protected)
func Me(c *gin.Context) {
userID, _ := c.Get("user_id")

var user models.User
if err := database.DB.First(&user, userID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
return
}

c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc - updates the logged-in user's editable profile fields
// PUT /api/v1/auth/me (protected)
func UpdateProfile(c *gin.Context) {
userID, _ := c.Get("user_id")

var req models.UpdateProfileRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var user models.User
if err := database.DB.First(&user, userID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
return
}

user.Name = req.Name
if err := database.DB.Save(&user).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
return
}

c.JSON(http.StatusOK, user)
}

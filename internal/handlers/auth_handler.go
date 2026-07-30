package handlers

import (
	"crypto/rand"
	"fmt"
	"log"
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
// Generates a 6-digit OTP for the phone number and "sends" it.
//
// NOTE (dev only): There is no SMS gateway wired up yet, so the OTP is
// returned directly in the JSON response and logged to the server console
// so you can test the flow end-to-end locally. Before this goes anywhere
// near production, plug in a real SMS provider (e.g. MSG91, Fast2SMS,
// Twilio) here and DELETE the "otp" field from the response — returning
// the code in the API response is a placeholder for local testing only
// and must never ship like this.
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

	// Invalidate any previous unverified OTPs for this phone before issuing a new one.
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

	// Dev-only: log instead of actually sending an SMS.
	log.Printf("[DEV OTP] phone=%s code=%s (valid %d min)", req.Phone, code, otpValidityMinutes)

	c.JSON(http.StatusOK, gin.H{
		"message":            "OTP sent successfully",
		"otp":                code, // DEV ONLY — remove once a real SMS gateway is integrated
		"expires_in_minutes": otpValidityMinutes,
	})
}

// VerifyOTP godoc
// POST /api/v1/auth/verify-otp
// Verifies the OTP; creates the user on first login, logs them in on repeat
// visits. Returns a JWT either way — this single endpoint covers both
// "signup" and "login" since phone + OTP is the only credential.
func VerifyOTP(c *gin.Context) {
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

	// Mark this OTP as used so it can't be replayed.
	database.DB.Model(&otp).Update("verified", true)

	// Find or create the user — first successful OTP verification = signup.
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

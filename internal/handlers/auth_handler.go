package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"

	twilio "github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/verify/v2"
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

// getTwilioClient builds a Twilio REST client using TWILIO_ACCOUNT_SID and
// TWILIO_AUTH_TOKEN from the environment (set in .env / Render env vars).
func getTwilioClient() *twilio.RestClient {
	return twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: os.Getenv("TWILIO_ACCOUNT_SID"),
		Password: os.Getenv("TWILIO_AUTH_TOKEN"),
	})
}

// SendOTP godoc
// POST /api/v1/auth/send-otp
// Uses Twilio Verify to generate and send a real SMS OTP to the phone number.
// Twilio stores and expires the code itself — no local OTP table needed.
func SendOTP(c *gin.Context) {
	var req models.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := getTwilioClient()
	verifySid := os.Getenv("TWILIO_VERIFY_SID")

	params := &twilioApi.CreateVerificationParams{}
	params.SetTo(req.Phone)
	params.SetChannel("sms")

	_, err := client.VerifyV2.CreateVerification(verifySid, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
	})
}

// VerifyOTP godoc
// POST /api/v1/auth/verify-otp
// Verifies the OTP via Twilio Verify; creates the user on first login, logs
// them in on repeat visits. Returns a JWT either way.
func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := getTwilioClient()
	verifySid := os.Getenv("TWILIO_VERIFY_SID")

	params := &twilioApi.CreateVerificationCheckParams{}
	params.SetTo(req.Phone)
	params.SetCode(req.OTP)

	resp, err := client.VerifyV2.CreateVerificationCheck(verifySid, params)
	if err != nil || resp.Status == nil || *resp.Status != "approved" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

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
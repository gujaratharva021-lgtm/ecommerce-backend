package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
)

var (
	ErrCouponNotFound = errors.New("coupon not found")
	ErrCouponInactive = errors.New("this coupon is no longer active")
	ErrCouponExpired  = errors.New("this coupon has expired")
	ErrCouponUsedUp   = errors.New("this coupon has reached its usage limit")
	ErrMinOrderNotMet = errors.New("order amount does not meet the minimum required for this coupon")
)

// ValidateCoupon checks a coupon code against an order amount and returns the
// coupon plus the discount amount to apply. Shared by ValidateCouponHandler
// (POST /coupons/validate) and the Checkout handler, so preview and actual
// checkout always agree on the discount. Pass database.DB for a standalone
// check, or a *gorm.DB transaction (tx) when validating inside Checkout so
// the read is consistent with the rest of that transaction.
func ValidateCoupon(db *gorm.DB, code string, orderAmount float64) (*models.Coupon, float64, error) {
	var coupon models.Coupon
	if err := db.Where("UPPER(code) = UPPER(?)", code).First(&coupon).Error; err != nil {
		return nil, 0, ErrCouponNotFound
	}
	if !coupon.IsActive {
		return nil, 0, ErrCouponInactive
	}
	if time.Now().After(coupon.ExpiryDate) {
		return nil, 0, ErrCouponExpired
	}
	if coupon.UsedCount >= coupon.UsageLimit {
		return nil, 0, ErrCouponUsedUp
	}
	if orderAmount < coupon.MinOrderAmount {
		return nil, 0, ErrMinOrderNotMet
	}

	var discount float64
	if coupon.DiscountType == models.CouponTypeFlat {
		discount = coupon.DiscountValue
	} else {
		discount = orderAmount * (coupon.DiscountValue / 100)
	}
	if coupon.MaxDiscountAmount != nil && discount > *coupon.MaxDiscountAmount {
		discount = *coupon.MaxDiscountAmount
	}
	if discount > orderAmount {
		discount = orderAmount
	}
	return &coupon, discount, nil
}

// ApplyCoupon records coupon usage against an order and increments used_count.
// Call this inside the Checkout transaction (pass tx) right after the order
// row is saved, only when a coupon was actually used for that order — this
// way a failed/rolled-back checkout never burns a coupon use.
func ApplyCoupon(db *gorm.DB, coupon *models.Coupon, orderID uint, discount float64) error {
	// Atomically increment used_count only if the limit hasn't been reached
	// in the meantime — guards against two concurrent checkouts both passing
	// ValidateCoupon before either has committed (a plain read-then-write
	// increment could let usage exceed usage_limit under concurrent load).
	result := db.Model(&models.Coupon{}).
		Where("id = ? AND used_count < usage_limit", coupon.ID).
		UpdateColumn("used_count", gorm.Expr("used_count + 1"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCouponUsedUp
	}

	oc := models.OrderCoupon{
		OrderID:        orderID,
		CouponID:       coupon.ID,
		DiscountAmount: discount,
	}
	return db.Create(&oc).Error
}

// ValidateCouponHandler godoc
// POST /api/v1/coupons/validate (protected)
// Lets the frontend show a discount preview on the checkout screen before
// actually placing the order.
func ValidateCouponHandler(c *gin.Context) {
	var req models.ValidateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon, discount, err := ValidateCoupon(database.DB, req.Code, req.OrderAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":           true,
		"coupon_id":       coupon.ID,
		"code":            coupon.Code,
		"discount_amount": discount,
		"final_amount":    req.OrderAmount - discount,
	})
}

// ---------- Admin ----------

// CreateCoupon godoc
// POST /api/v1/admin/coupons (admin only)
func CreateCoupon(c *gin.Context) {
	var req models.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expiry, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiry_date, use YYYY-MM-DD"})
		return
	}
	usageLimit := req.UsageLimit
	if usageLimit <= 0 {
		usageLimit = 1
	}

	coupon := models.Coupon{
		Code:              req.Code,
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        usageLimit,
		ExpiryDate:        expiry,
		IsActive:          true,
	}

	if err := database.DB.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon (code may already exist)"})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// GetCoupons godoc
// GET /api/v1/admin/coupons (admin only)
func GetCoupons(c *gin.Context) {
	var coupons []models.Coupon
	if err := database.DB.Order("created_at desc").Find(&coupons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coupons"})
		return
	}
	c.JSON(http.StatusOK, coupons)
}

// UpdateCouponStatus godoc
// PUT /api/v1/admin/coupons/:id/status (admin only)
// body: { "is_active": true }
func UpdateCouponStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
		return
	}

	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var coupon models.Coupon
	if err := database.DB.First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}
	coupon.IsActive = body.IsActive
	if err := database.DB.Save(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update coupon"})
		return
	}

	c.JSON(http.StatusOK, coupon)
}


// DeleteCoupon godoc
// DELETE /api/v1/admin/coupons/:id (admin only, requires coupon:delete permission)
func DeleteCoupon(c *gin.Context) {
id, err := strconv.Atoi(c.Param("id"))
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coupon id"})
return
}

var coupon models.Coupon
if err := database.DB.First(&coupon, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
return
}

if err := database.DB.Delete(&coupon).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coupon"})
return
}

c.JSON(http.StatusOK, gin.H{"success": true})
}

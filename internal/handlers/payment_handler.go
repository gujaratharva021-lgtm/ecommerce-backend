package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// CreatePaymentOrder godoc
// POST /api/v1/orders/:id/payment (protected)
// Creates a Razorpay order for an existing app order (payment_method must
// be "online") and returns what the frontend needs to open Razorpay
// Checkout. Safe to call again to retry after a failed/abandoned attempt â€”
// it overwrites the same Payment row rather than creating duplicates.
func CreatePaymentOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this order"})
		return
	}
	if order.PaymentMethod != models.PaymentMethodOnline {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This order was placed as Cash on Delivery, not online payment"})
		return
	}
	if order.PaymentStatus == models.OrderPaymentStatusPaid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This order is already paid"})
		return
	}
	if order.Status == models.OrderStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot pay for a cancelled order"})
		return
	}

	rzpOrderID, err := utils.CreateRazorpayOrder(order.TotalAmount, "order_receipt_"+orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var payment models.Payment
	if err := database.DB.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
		payment = models.Payment{OrderID: order.ID}
	}
	payment.RazorpayOrderID = rzpOrderID
	payment.Amount = order.TotalAmount
	payment.Currency = "INR"
	payment.Status = models.PaymentStatusCreated

	if err := database.DB.Save(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record payment order"})
		return
	}

	c.JSON(http.StatusOK, models.CreatePaymentOrderResponse{
		RazorpayOrderID: rzpOrderID,
		Amount:          int64(order.TotalAmount*100 + 0.5),
		Currency:        "INR",
		KeyID:           config.AppConfig.RazorpayKeyID,
		OrderID:         order.ID,
	})
}

// VerifyPayment godoc
// POST /api/v1/orders/:id/payment/verify (protected)
// Verifies the Razorpay signature returned by Checkout on the frontend.
// On success, marks the Payment + Order as paid and auto-advances a
// still-pending order to "confirmed". On signature mismatch, the payment
// is marked failed and the order is left untouched â€” the frontend can
// retry via CreatePaymentOrder.
func VerifyPayment(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this order"})
		return
	}

	var req models.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var payment models.Payment
	if err := database.DB.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No payment order found for this order. Call POST /orders/:id/payment first."})
		return
	}
	if payment.RazorpayOrderID != req.RazorpayOrderID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Razorpay order ID does not match the one created for this order"})
		return
	}

	if !utils.VerifyRazorpaySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
		payment.Status = models.PaymentStatusFailed
		database.DB.Save(&payment)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment signature verification failed"})
		return
	}

	payment.RazorpayPaymentID = req.RazorpayPaymentID
	payment.RazorpaySignature = req.RazorpaySignature
	payment.Status = models.PaymentStatusPaid
	if err := database.DB.Save(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment"})
		return
	}

	order.PaymentStatus = models.OrderPaymentStatusPaid
	if order.Status == models.OrderStatusPending {
		order.Status = models.OrderStatusConfirmed
	}
	if err := database.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment was verified but failed to update the order â€” contact support"})
		return
	}
        go services.AutoAssignDeliveryPartner(order.ID)

	var addr models.Address
	database.DB.First(&addr, order.AddressID)
	message := "Payment received for order #" + orderID + ". Your order is now confirmed."
	utils.SendNotification(addr.Phone, message, "payment_received", &order.ID)
	services.SendPushToUser(order.UserID, "Payment Received", message)

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment verified successfully",
		"order":   order,
	})
}


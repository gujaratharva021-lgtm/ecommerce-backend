package handlers

import (
"errors"
"fmt"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"gorm.io/gorm"
"gorm.io/gorm/clause"
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

// errAssignConflict/errAssignValidation are internal sentinels used to
// short-circuit the transaction below with a specific, already-decided
// HTTP response, distinct from an unexpected DB error.
var (
	errAssignOrderNotFound   = errors.New("order not found")
	errAssignBadOrderStatus  = errors.New("order not eligible for assignment")
	errAssignPartnerNotFound = errors.New("delivery partner not found")
	errAssignPartnerInactive = errors.New("delivery partner is not active")
	errAssignPartnerOffline  = errors.New("delivery partner is offline")
	errAssignAlreadyActive   = errors.New("order already has an active delivery assignment")
)

// AssignDeliveryPartner godoc
// PUT /api/v1/admin/orders/:id/assign-delivery (admin only)
//
// Assigns (or re-assigns, after a rejection) a delivery partner to an
// order. Only ONLINE, active partners are eligible - offline partners
// must never receive new assignments. The whole check-then-write runs
// inside one transaction with the order row locked for its duration, so
// two concurrent assign requests for the same order can never both
// succeed (task #10: no double assignment).
func AssignDeliveryPartner(c *gin.Context) {
	orderID := c.Param("id")

	var req models.AssignDeliveryPartnerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Lock the order row for the duration of this transaction so a
		// concurrent assignment (admin or auto-assign) for the same
		// order has to wait, rather than racing on the read-then-write
		// below.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
			return errAssignOrderNotFound
		}

		if order.Status != models.OrderStatusConfirmed && order.Status != models.OrderStatusShipped {
			return errAssignBadOrderStatus
		}

		// Only re-assignable if there is no partner yet, or the
		// previous partner rejected the delivery. An ASSIGNED or
		// ACCEPTED order is already actively owned by a partner and
		// must not be silently reassigned here.
		if order.DeliveryPartnerID != nil &&
			(order.DeliveryAssignmentStatus == nil || *order.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusRejected) {
			return errAssignAlreadyActive
		}

		var partner models.DeliveryPartner
		if err := tx.First(&partner, req.DeliveryPartnerID).Error; err != nil {
			return errAssignPartnerNotFound
		}
		if !partner.IsActive {
			return errAssignPartnerInactive
		}
		if !partner.IsOnline {
			return errAssignPartnerOffline
		}

		newStatus := models.DeliveryAssignmentStatusAssigned
		order.DeliveryPartnerID = &req.DeliveryPartnerID
		order.DeliveryAssignmentStatus = &newStatus
		order.DeliveryRejectionReason = nil

		if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Updates(map[string]interface{}{
			"delivery_partner_id":        req.DeliveryPartnerID,
			"delivery_assignment_status": newStatus,
			"delivery_rejection_reason":  nil,
		}).Error; err != nil {
			return fmt.Errorf("failed to assign delivery partner: %w", err)
		}
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, errAssignOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		case errors.Is(err, errAssignBadOrderStatus):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Delivery partner can only be assigned to confirmed or shipped orders"})
		case errors.Is(err, errAssignAlreadyActive):
			c.JSON(http.StatusConflict, gin.H{"error": "Order already has an active delivery assignment"})
		case errors.Is(err, errAssignPartnerNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Delivery partner not found"})
		case errors.Is(err, errAssignPartnerInactive):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Delivery partner is not active"})
		case errors.Is(err, errAssignPartnerOffline):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Delivery partner is offline and cannot receive new assignments"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign delivery partner"})
		}
		return
	}

	database.DB.Preload("DeliveryPartner").First(&order, order.ID)

	// Notify the partner's device(s) that a new order has been assigned.
	go services.SendPushToPartner(
		req.DeliveryPartnerID,
		"New delivery assigned",
		fmt.Sprintf("Order #%d has been assigned to you", order.ID),
	)

	c.JSON(http.StatusOK, gin.H{"message": "Delivery partner assigned", "order": order})
}

// ---------------------------------------------------------------------------
// Delivery Partner order actions (delivery partner only)
// ---------------------------------------------------------------------------

// AssignedOrderSummary is the shape returned by GET /delivery/orders. It
// deliberately surfaces only what a courier needs to fulfil the delivery -
// it does not include the customer's account/user record, other saved
// addresses, product cost/margin fields, etc.
type AssignedOrderSummary struct {
	OrderID          uint       `json:"order_id"`
	Status           string     `json:"status"`
	AssignmentStatus *string    `json:"assignment_status,omitempty"`
	RejectionReason  *string    `json:"rejection_reason,omitempty"`
	DeliveryAddress  string     `json:"delivery_address"`
	CustomerName     string     `json:"customer_name"`
	CustomerPhone    string     `json:"customer_phone"`
	TotalAmount      float64    `json:"total_amount"`
	PaymentMethod    string     `json:"payment_method"`
	ItemCount        int        `json:"item_count"`
	CreatedAt        time.Time  `json:"created_at"`
}

func toAssignedOrderSummary(o models.Order) AssignedOrderSummary {
	addr := fmt.Sprintf("%s, %s, %s - %s", o.Address.Line1, o.Address.City, o.Address.State, o.Address.Pincode)
	if o.Address.Line2 != "" {
		addr = fmt.Sprintf("%s, %s, %s, %s - %s", o.Address.Line1, o.Address.Line2, o.Address.City, o.Address.State, o.Address.Pincode)
	}
	return AssignedOrderSummary{
		OrderID:          o.ID,
		Status:           o.Status,
		AssignmentStatus: o.DeliveryAssignmentStatus,
		RejectionReason:  o.DeliveryRejectionReason,
		DeliveryAddress:  addr,
		CustomerName:     o.Address.FullName,
		CustomerPhone:    o.Address.Phone,
		TotalAmount:      o.TotalAmount,
		PaymentMethod:    o.PaymentMethod,
		ItemCount:        len(o.Items),
		CreatedAt:        o.CreatedAt,
	}
}

// GetMyDeliveries godoc
// GET /api/v1/delivery/orders (delivery partner only)
// Returns orders assigned to the logged-in delivery partner, identified
// solely from the verified JWT ("user_id") - never from a client-supplied
// delivery_boy_id, so one partner can never list another's orders (IDOR).
// Optional ?status= filters by order status (e.g. confirmed, shipped,
// delivered); optional ?assignment_status= filters by
// assigned/accepted/rejected.
func GetMyDeliveries(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)

query := database.DB.
Preload("Address").
Preload("Items").
Where("delivery_partner_id = ?", partnerID)

if status := c.Query("status"); status != "" {
query = query.Where("status = ?", status)
}
if assignmentStatus := c.Query("assignment_status"); assignmentStatus != "" {
query = query.Where("delivery_assignment_status = ?", assignmentStatus)
}

var orders []models.Order
if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load assigned orders"})
return
}

summaries := make([]AssignedOrderSummary, 0, len(orders))
for _, o := range orders {
summaries = append(summaries, toAssignedOrderSummary(o))
}

c.JSON(http.StatusOK, gin.H{"orders": summaries})
}

// ---------------------------------------------------------------------------
// Assignment accept / reject (delivery partner only)
// ---------------------------------------------------------------------------

var (
	errAssignmentOrderNotOwned = errors.New("order not found or not assigned to you")
	errAssignmentNotPending    = errors.New("assignment is not pending")
)

// AcceptAssignment godoc
// PUT /api/v1/delivery/orders/:id/accept (delivery partner only)
// Moves a pending assignment ASSIGNED -> ACCEPTED. Only the partner the
// order is currently assigned to can accept it - the lookup is scoped by
// "id = ? AND delivery_partner_id = ?" using the authenticated partner's
// own ID, so changing the :id in the URL can never let one partner accept
// (or discover) another partner's order (IDOR/BOLA protection). The
// conditional UPDATE additionally guards against a double-accept race:
// if two requests for the same order land at once, only the one that
// still finds delivery_assignment_status = 'assigned' succeeds.
func AcceptAssignment(c *gin.Context) {
	respondToAssignment(c, models.DeliveryAssignmentStatusAccepted, "")
}

// RejectAssignment godoc
// PUT /api/v1/delivery/orders/:id/reject (delivery partner only)
// Moves a pending assignment ASSIGNED -> REJECTED. Same ownership/IDOR and
// concurrency guarantees as AcceptAssignment. The order is never deleted
// and is not automatically reassigned here - that's handled separately by
// an admin/dispatcher calling assign-delivery again (Phase 3 spec
// explicitly defers auto-reassignment to a later phase).
func RejectAssignment(c *gin.Context) {
	var req models.RejectAssignmentRequest
	// Body is optional - an empty/absent body just means no reason given.
	_ = c.ShouldBindJSON(&req)
	respondToAssignment(c, models.DeliveryAssignmentStatusRejected, req.Reason)
}

func respondToAssignment(c *gin.Context, newStatus string, reason string) {
	partnerID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).
			First(&order).Error; err != nil {
			return errAssignmentOrderNotOwned
		}

		if order.DeliveryAssignmentStatus == nil || *order.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAssigned {
			return errAssignmentNotPending
		}

		updates := map[string]interface{}{"delivery_assignment_status": newStatus}
		if newStatus == models.DeliveryAssignmentStatusRejected {
			if reason != "" {
				updates["delivery_rejection_reason"] = reason
			}
		}

		result := tx.Model(&models.Order{}).
			Where("id = ? AND delivery_partner_id = ? AND delivery_assignment_status = ?",
				order.ID, partnerID, models.DeliveryAssignmentStatusAssigned).
			Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("failed to update assignment: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			// Someone else (impossible for another partner, but a
			// concurrent request from the same partner) beat us to it.
			return errAssignmentNotPending
		}
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, errAssignmentOrderNotOwned):
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or not assigned to you"})
		case errors.Is(err, errAssignmentNotPending):
			c.JSON(http.StatusBadRequest, gin.H{"error": "This assignment has already been responded to or is not in a pending state"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignment"})
		}
		return
	}

	database.DB.Preload("Address").Preload("Items").First(&order, order.ID)

	msg := "Delivery accepted"
	if newStatus == models.DeliveryAssignmentStatusRejected {
		msg = "Delivery rejected"
	}
	c.JSON(http.StatusOK, gin.H{"message": msg, "order": toAssignedOrderSummary(order)})
}

// UpdateDeliveryOrderStatus godoc
// PUT /api/v1/delivery/orders/:id/status (delivery partner only)
// Lets the assigned partner move an order from confirmed -> shipped
// (picked up and out for delivery). Partners cannot set "delivered"
// here - see ConfirmDelivery below, which also handles COD collection.
func UpdateDeliveryOrderStatus(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var req struct {
Status string `json:"status" binding:"required,oneof=shipped"`
}
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var order models.Order
if err := database.DB.Preload("Address").Preload("Items.Product").Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or not assigned to you"})
return
}

if order.Status != models.OrderStatusHandedOver {
c.JSON(http.StatusBadRequest, gin.H{"error": "Order must be handed over by the warehouse before it can be marked shipped"})
return
}

order.Status = req.Status
if err := database.DB.Save(&order).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
return
}

// Notify the customer that their order is out for delivery.
go services.SendPushToUser(
order.UserID,
"Order out for delivery",
fmt.Sprintf("Your order #%d is on its way", order.ID),
)

c.JSON(http.StatusOK, gin.H{"message": "Order status updated", "order": order})
}

// ConfirmDelivery godoc
// PUT /api/v1/delivery/orders/:id/deliver (delivery partner only)
// Marks the order delivered. For COD orders this also marks payment as
// paid, since cash is collected at the same time.
func ConfirmDelivery(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Preload("Address").Preload("Items.Product").Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or not assigned to you"})
return
}

if order.Status != models.OrderStatusShipped {
c.JSON(http.StatusBadRequest, gin.H{"error": "Order must be shipped before it can be marked delivered"})
return
}

order.Status = models.OrderStatusDelivered
if order.PaymentMethod == models.PaymentMethodCOD {
order.PaymentStatus = models.OrderPaymentStatusPaid
}

if err := database.DB.Save(&order).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to confirm delivery"})
return
}

// Notify the customer that their order has been delivered.
go services.SendPushToUser(
order.UserID,
"Order delivered",
fmt.Sprintf("Your order #%d has been delivered. Enjoy!", order.ID),
)

c.JSON(http.StatusOK, gin.H{"message": "Delivery confirmed", "order": order})
}

// GetMyEarnings godoc
// GET /api/v1/delivery/earnings (delivery partner only)
// Returns delivery earnings summary: today's earnings, total earnings,
// total delivered count, and a list of delivered orders with the flat
// per-delivery payout. There is no separate earnings/payout table yet,
// so this is computed on the fly at a fixed rate per delivered order.
const perDeliveryEarning = 30.0

func GetMyEarnings(c *gin.Context) {
partnerID := c.MustGet("user_id").(uint)

var deliveredOrders []models.Order
if err := database.DB.
Where("delivery_partner_id = ? AND status = ?", partnerID, models.OrderStatusDelivered).
Order("updated_at DESC").
Find(&deliveredOrders).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load earnings"})
return
}

totalEarnings := float64(len(deliveredOrders)) * perDeliveryEarning

todayStart := time.Now().Truncate(24 * time.Hour)
todayCount := 0
todayEarnings := 0.0
type EarningEntry struct {
OrderID     uint      `json:"order_id"`
Amount      float64   `json:"amount"`
DeliveredAt time.Time `json:"delivered_at"`
}
entries := make([]EarningEntry, 0, len(deliveredOrders))

for _, o := range deliveredOrders {
if o.UpdatedAt.After(todayStart) {
todayCount++
todayEarnings += perDeliveryEarning
}
entries = append(entries, EarningEntry{
OrderID:     o.ID,
Amount:      perDeliveryEarning,
DeliveredAt: o.UpdatedAt,
})
}

c.JSON(http.StatusOK, gin.H{
"per_delivery_rate": perDeliveryEarning,
"today_earnings":    todayEarnings,
"today_deliveries":  todayCount,
"total_earnings":    totalEarnings,
"total_deliveries":  len(deliveredOrders),
"entries":           entries,
})
}

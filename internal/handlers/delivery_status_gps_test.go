package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/middleware"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// newStatusGPSTestRouter wires up exactly the routes exercised by this
// file: admin assign-delivery (to get an order into an assigned state),
// delivery accept (to reach ACCEPTED), the new delivery-status endpoint,
// and the GPS location endpoint - each guarded the same way as the real
// app.
func newStatusGPSTestRouter() *gin.Engine {
	r := gin.New()

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
	admin.PUT("/orders/:id/assign-delivery", AssignDeliveryPartner)

	delivery := r.Group("/api/v1/delivery")
	delivery.PUT("/location", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), UpdateLocation)
	delivery.PUT("/orders/:id/accept", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), AcceptAssignment)
	delivery.PUT("/orders/:id/delivery-status", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), UpdateDeliveryStatus)

	return r
}

// ---------------------------------------------------------------------------
// GPS tracking
// ---------------------------------------------------------------------------

func TestUpdateLocation_ValidCoordinatesSucceeds(t *testing.T) {
	r := newStatusGPSTestRouter()
	partner := seedAssignPartner(t, true)
	token := deliveryPartnerToken(t, partner)

	w := doRequest(r, http.MethodPut, "/api/v1/delivery/location", token, gin.H{"lat": 23.0225, "lng": 72.5714})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.DeliveryPartner
	database.DB.First(&fresh, partner.ID)
	if fresh.CurrentLat == nil || *fresh.CurrentLat != 23.0225 {
		t.Errorf("expected current_lat to be updated, got %v", fresh.CurrentLat)
	}
	if fresh.CurrentLng == nil || *fresh.CurrentLng != 72.5714 {
		t.Errorf("expected current_lng to be updated, got %v", fresh.CurrentLng)
	}
	if fresh.LastLocationUpdate == nil {
		t.Errorf("expected last_location_update to be set")
	}
}

func TestUpdateLocation_InvalidLatitudeRejected(t *testing.T) {
	r := newStatusGPSTestRouter()
	partner := seedAssignPartner(t, true)
	token := deliveryPartnerToken(t, partner)

	w := doRequest(r, http.MethodPut, "/api/v1/delivery/location", token, gin.H{"lat": 95.0, "lng": 72.5714})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range latitude, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.DeliveryPartner
	database.DB.First(&fresh, partner.ID)
	if fresh.CurrentLat != nil {
		t.Errorf("location must not be stored when latitude is invalid, got %v", fresh.CurrentLat)
	}
}

func TestUpdateLocation_InvalidLongitudeRejected(t *testing.T) {
	r := newStatusGPSTestRouter()
	partner := seedAssignPartner(t, true)
	token := deliveryPartnerToken(t, partner)

	w := doRequest(r, http.MethodPut, "/api/v1/delivery/location", token, gin.H{"lat": 23.0225, "lng": -200.0})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-range longitude, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.DeliveryPartner
	database.DB.First(&fresh, partner.ID)
	if fresh.CurrentLng != nil {
		t.Errorf("location must not be stored when longitude is invalid, got %v", fresh.CurrentLng)
	}
}

// ---------------------------------------------------------------------------
// Delivery status state machine
// ---------------------------------------------------------------------------

// acceptedOrderForStatusTest seeds a partner + order, assigns it (ASSIGNED),
// then accepts it (ACCEPTED) via the real handlers, so tests start from a
// realistic pre-condition rather than writing delivery_status directly.
func acceptedOrderForStatusTest(t *testing.T, r *gin.Engine) (models.DeliveryPartner, models.Order, string) {
	t.Helper()
	adminAssignTokenCache = adminAssignToken(t)
	partner := seedAssignPartner(t, true)
	order := seedAssignOrder(t, models.OrderStatusConfirmed)

	if resp := assignOrder(r, order.ID, partner.ID); resp.Code != http.StatusOK {
		t.Fatalf("setup: failed to assign order: %d %s", resp.Code, resp.Body)
	}

	token := deliveryPartnerToken(t, partner)
	w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/delivery/orders/%d/accept", order.ID), token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("setup: failed to accept order: %d %s", w.Code, w.Body.String())
	}

	var fresh models.Order
	database.DB.First(&fresh, order.ID)
	return partner, fresh, token
}

func TestUpdateDeliveryStatus_ValidTransitionSucceeds(t *testing.T) {
	r := newStatusGPSTestRouter()
	_, order, token := acceptedOrderForStatusTest(t, r)

	// ACCEPTED -> PICKED_UP is the first step a partner may take through
	// this endpoint (ASSIGNED -> ACCEPTED already happened via /accept).
	w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/delivery/orders/%d/delivery-status", order.ID), token, gin.H{"status": "picked_up"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for ACCEPTED->PICKED_UP, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.Order
	database.DB.First(&fresh, order.ID)
	if fresh.DeliveryStatus == nil || *fresh.DeliveryStatus != models.DeliveryStatusPickedUp {
		t.Errorf("expected delivery_status 'picked_up', got %v", fresh.DeliveryStatus)
	}

	// Continue the full chain to make sure every subsequent step also
	// succeeds in order.
	steps := []string{"out_for_delivery", "arrived", "delivered"}
	for _, step := range steps {
		w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/delivery/orders/%d/delivery-status", order.ID), token, gin.H{"status": step})
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 transitioning to %s, got %d: %s", step, w.Code, w.Body.String())
		}
	}

	database.DB.First(&fresh, order.ID)
	if fresh.DeliveryStatus == nil || *fresh.DeliveryStatus != models.DeliveryStatusDelivered {
		t.Errorf("expected final delivery_status 'delivered', got %v", fresh.DeliveryStatus)
	}
}

func TestUpdateDeliveryStatus_InvalidTransitionRejected(t *testing.T) {
	r := newStatusGPSTestRouter()
	_, order, token := acceptedOrderForStatusTest(t, r)

	// Current state is ACCEPTED; skipping straight to "arrived" (or
	// "delivered") must be rejected - only the immediate next step is
	// allowed.
	w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/delivery/orders/%d/delivery-status", order.ID), token, gin.H{"status": "arrived"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for out-of-order transition, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.Order
	database.DB.First(&fresh, order.ID)
	if fresh.DeliveryStatus == nil || *fresh.DeliveryStatus != models.DeliveryStatusAccepted {
		t.Errorf("delivery_status must remain 'accepted' after a rejected out-of-order transition, got %v", fresh.DeliveryStatus)
	}
}

func TestUpdateDeliveryStatus_UnauthorizedPartnerRejected(t *testing.T) {
	r := newStatusGPSTestRouter()
	_, order, _ := acceptedOrderForStatusTest(t, r)

	intruder := seedAssignPartner(t, true)
	intruderToken := deliveryPartnerToken(t, intruder)

	w := doRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/delivery/orders/%d/delivery-status", order.ID), intruderToken, gin.H{"status": "picked_up"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 (no IDOR leak) for another partner's order, got %d: %s", w.Code, w.Body.String())
	}

	var fresh models.Order
	database.DB.First(&fresh, order.ID)
	if fresh.DeliveryStatus == nil || *fresh.DeliveryStatus != models.DeliveryStatusAccepted {
		t.Errorf("delivery_status must remain 'accepted' after a rejected intrusion attempt, got %v", fresh.DeliveryStatus)
	}
}

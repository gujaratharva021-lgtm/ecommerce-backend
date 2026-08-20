package services

import (
	"errors"
	"fmt"

	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrDeliveryStatusOrderNotOwned / ErrDeliveryStatusInvalidTransition are
// the sentinel errors UpdateDeliveryStatus returns for the two ways a
// delivery-status update can be legitimately refused.
var (
	ErrDeliveryStatusOrderNotOwned     = errors.New("order not found or not assigned to you")
	ErrDeliveryStatusInvalidTransition = errors.New("invalid delivery status transition")
)

// deliveryStatusTransitions defines the only allowed forward moves in the
// granular, courier-driven delivery lifecycle:
//
//	ASSIGNED -> ACCEPTED -> PICKED_UP -> OUT_FOR_DELIVERY -> ARRIVED -> DELIVERED
//
// ASSIGNED and ACCEPTED are set automatically elsewhere (assign/auto-assign
// and accept flows) and are not reachable through UpdateDeliveryStatus -
// see the oneof binding on models.UpdateDeliveryStatusRequest.
var deliveryStatusTransitions = map[string]map[string]bool{
	models.DeliveryStatusAssigned:       {models.DeliveryStatusAccepted: true},
	models.DeliveryStatusAccepted:       {models.DeliveryStatusPickedUp: true},
	models.DeliveryStatusPickedUp:       {models.DeliveryStatusOutForDelivery: true},
	models.DeliveryStatusOutForDelivery: {models.DeliveryStatusArrived: true},
	models.DeliveryStatusArrived:        {models.DeliveryStatusDelivered: true},
}

// UpdateDeliveryStatus lets the assigned delivery partner advance an
// order's granular delivery status one step at a time. It is the single
// source of truth for that state transition - shared by the
// PUT /delivery/orders/:id/delivery-status handler and by tests.
//
// Ownership is enforced by scoping the lookup to
// "id = ? AND delivery_partner_id = ?" using the *caller-supplied*
// partnerID (the handler sets this from the verified JWT, never from a
// client-supplied field), so one partner can never advance or discover
// another partner's order (IDOR/BOLA protection) - mirrors
// RespondToAssignment in delivery_acceptance.go.
//
// The transition is only allowed along deliveryStatusTransitions, enforced
// both by an explicit check and by a conditional UPDATE ...
// WHERE COALESCE(delivery_status,”) = <the status just read>, so two
// concurrent updates for the same order (e.g. a double-tap) can only ever
// have one winner.
//
// This is intentionally independent of Order.Status - existing endpoints
// (UpdateDeliveryOrderStatus, ConfirmDelivery) keep driving Order.Status
// exactly as before; this only adds the finer-grained delivery_status
// column alongside it.
func UpdateDeliveryStatus(orderID, partnerID uint, newStatus string) (*models.Order, error) {
	var order models.Order
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND delivery_partner_id = ?", orderID, partnerID).
			First(&order).Error; err != nil {
			return ErrDeliveryStatusOrderNotOwned
		}

		current := ""
		if order.DeliveryStatus != nil {
			current = *order.DeliveryStatus
		}
		if !deliveryStatusTransitions[current][newStatus] {
			return ErrDeliveryStatusInvalidTransition
		}

		result := tx.Model(&models.Order{}).
			Where("id = ? AND delivery_partner_id = ? AND COALESCE(delivery_status, '') = ?", order.ID, partnerID, current).
			Update("delivery_status", newStatus)
		if result.Error != nil {
			return fmt.Errorf("failed to update delivery status: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrDeliveryStatusInvalidTransition
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	database.DB.Preload("Address").Preload("Items").First(&order, order.ID)
	return &order, nil
}

package services

import (
    "errors"
    "strconv"
    "strings"
    "testing"
    "time"

    "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
    "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

func formatIDs(ids ...uint) string {
    parts := make([]string, len(ids))
    for i, id := range ids {
        parts[i] = strconv.FormatUint(uint64(id), 10)
    }
    return strings.Join(parts, ",")
}

func reload(t *testing.T, orderID uint) models.Order {
    t.Helper()
    var o models.Order
    if err := database.DB.First(&o, orderID).Error; err != nil {
        t.Fatalf("failed to reload order %d: %v", orderID, err)
    }
    return o
}

// seedAssignedOrder creates an order already in the ASSIGNED state, offered
// to partnerID, with the given attempted-partners history and expiry.
func seedAssignedOrder(t *testing.T, addressID, partnerID uint, attempted string, expiresAt *time.Time) models.Order {
    t.Helper()
    status := models.DeliveryAssignmentStatusAssigned
    order := models.Order{
        UserID:                      1,
        AddressID:                   addressID,
        ItemsAmount:                 100,
        TotalAmount:                 100,
        Status:                      models.OrderStatusConfirmed,
        DeliveryPartnerID:           &partnerID,
        DeliveryAssignmentStatus:    &status,
        DeliveryAttemptedPartnerIDs: attempted,
        DeliveryAssignmentExpiresAt: expiresAt,
    }
    if err := database.DB.Create(&order).Error; err != nil {
        t.Fatalf("failed to seed assigned order: %v", err)
    }
    return order
}

// --- Accept ---------------------------------------------------------------

func TestRespondToAssignment_Accept(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000010", 19.0760, 72.8777)
    partner := seedActivePartner(t, "8000000010", 19.0761, 72.8778)
    future := time.Now().Add(5 * time.Minute)
    order := seedAssignedOrder(t, addrID, partner.ID, formatIDs(partner.ID), &future)

    updated, err := RespondToAssignment(order.ID, partner.ID, models.DeliveryAssignmentStatusAccepted, "")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if updated.DeliveryAssignmentStatus == nil || *updated.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAccepted {
        t.Errorf("expected status accepted, got %v", updated.DeliveryAssignmentStatus)
    }
    if updated.DeliveryAssignmentExpiresAt != nil {
        t.Errorf("expected expiry cleared after accept, got %v", updated.DeliveryAssignmentExpiresAt)
    }
}

func TestRespondToAssignment_WrongPartner_NotOwned(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000011", 19.0760, 72.8777)
    partner := seedActivePartner(t, "8000000011", 19.0761, 72.8778)
    other := seedActivePartner(t, "8000000012", 19.0762, 72.8779)
    order := seedAssignedOrder(t, addrID, partner.ID, formatIDs(partner.ID), nil)

    _, err := RespondToAssignment(order.ID, other.ID, models.DeliveryAssignmentStatusAccepted, "")
    if !errors.Is(err, ErrAssignmentOrderNotOwned) {
        t.Errorf("expected ErrAssignmentOrderNotOwned, got %v", err)
    }
}

// --- Duplicate assignment (never respond twice) ----------------------------

func TestRespondToAssignment_DuplicateResponse_NotPending(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000013", 19.0760, 72.8777)
    partner := seedActivePartner(t, "8000000013", 19.0761, 72.8778)
    order := seedAssignedOrder(t, addrID, partner.ID, formatIDs(partner.ID), nil)

    if _, err := RespondToAssignment(order.ID, partner.ID, models.DeliveryAssignmentStatusAccepted, ""); err != nil {
        t.Fatalf("first accept failed: %v", err)
    }

    // Same partner tries to respond again (e.g. a double-tap) - must be
    // rejected as no longer pending, and never overwrite the accepted state.
    _, err := RespondToAssignment(order.ID, partner.ID, models.DeliveryAssignmentStatusRejected, "changed my mind")
    if !errors.Is(err, ErrAssignmentNotPending) {
        t.Errorf("expected ErrAssignmentNotPending, got %v", err)
    }

    fresh := reload(t, order.ID)
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAccepted {
        t.Errorf("duplicate response must not change an already-accepted assignment, got %v", fresh.DeliveryAssignmentStatus)
    }
}

// --- Reject -> automatic reassignment ---------------------------------------

func TestRespondToAssignment_Reject_TriggersReassignment(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000014", 19.0760, 72.8777)
    p1 := seedActivePartner(t, "8000000014", 19.0761, 72.8778)
    p2 := seedActivePartner(t, "8000000015", 19.0762, 72.8779)
    order := seedAssignedOrder(t, addrID, p1.ID, formatIDs(p1.ID), nil)

    updated, err := RespondToAssignment(order.ID, p1.ID, models.DeliveryAssignmentStatusRejected, "too far")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if updated.DeliveryAssignmentStatus == nil || *updated.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusRejected {
        t.Errorf("expected status rejected immediately, got %v", updated.DeliveryAssignmentStatus)
    }

    // Reassignment happens in the background - poll for it.
    deadline := time.Now().Add(3 * time.Second)
    var fresh models.Order
    for time.Now().Before(deadline) {
        fresh = reload(t, order.ID)
        if fresh.DeliveryPartnerID != nil && *fresh.DeliveryPartnerID == p2.ID {
            break
        }
        time.Sleep(100 * time.Millisecond)
    }
    if fresh.DeliveryPartnerID == nil || *fresh.DeliveryPartnerID != p2.ID {
        t.Fatalf("expected order reassigned to partner %d, got %v", p2.ID, fresh.DeliveryPartnerID)
    }
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAssigned {
        t.Errorf("expected reassigned order back in ASSIGNED state, got %v", fresh.DeliveryAssignmentStatus)
    }
}

// --- Timeout / expiry --------------------------------------------------------

func TestExpireStaleAssignments_ExpiresAndReassigns(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000016", 19.0760, 72.8777)
    p1 := seedActivePartner(t, "8000000016", 19.0761, 72.8778)
    p2 := seedActivePartner(t, "8000000017", 19.0762, 72.8779)
    past := time.Now().Add(-1 * time.Minute)
    order := seedAssignedOrder(t, addrID, p1.ID, formatIDs(p1.ID), &past)

    ExpireStaleAssignments()

    fresh := reload(t, order.ID)
    if fresh.DeliveryPartnerID == nil || *fresh.DeliveryPartnerID != p2.ID {
        t.Fatalf("expected order reassigned to partner %d after expiry, got %v", p2.ID, fresh.DeliveryPartnerID)
    }
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAssigned {
        t.Errorf("expected order back in ASSIGNED state after reassignment, got %v", fresh.DeliveryAssignmentStatus)
    }
}

func TestExpireStaleAssignments_NotYetExpired_Untouched(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000018", 19.0760, 72.8777)
    p1 := seedActivePartner(t, "8000000018", 19.0761, 72.8778)
    future := time.Now().Add(5 * time.Minute)
    order := seedAssignedOrder(t, addrID, p1.ID, formatIDs(p1.ID), &future)

    ExpireStaleAssignments()

    fresh := reload(t, order.ID)
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAssigned {
        t.Errorf("order not yet past its expiry must remain ASSIGNED, got %v", fresh.DeliveryAssignmentStatus)
    }
    if fresh.DeliveryPartnerID == nil || *fresh.DeliveryPartnerID != p1.ID {
        t.Errorf("order not yet past its expiry must keep its current partner, got %v", fresh.DeliveryPartnerID)
    }
}

// --- Reassignment never repeats a partner, and is race-safe -----------------

func TestTryAssignNextPartner_NeverOffersSamePartnerTwice(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000019", 19.0760, 72.8777)
    p1 := seedActivePartner(t, "8000000019", 19.0761, 72.8778)
    p2 := seedActivePartner(t, "8000000020", 19.0762, 72.8779)

    // Both partners have already been tried and rejected this order.
    status := models.DeliveryAssignmentStatusRejected
    order := models.Order{
        UserID:                      1,
        AddressID:                   addrID,
        ItemsAmount:                 100,
        TotalAmount:                 100,
        Status:                      models.OrderStatusConfirmed,
        DeliveryPartnerID:           &p2.ID,
        DeliveryAssignmentStatus:    &status,
        DeliveryAttemptedPartnerIDs: formatIDs(p1.ID, p2.ID),
    }
    if err := database.DB.Create(&order).Error; err != nil {
        t.Fatalf("failed to seed order: %v", err)
    }

    TryAssignNextPartner(order.ID)

    fresh := reload(t, order.ID)
    if fresh.DeliveryPartnerID != nil {
        t.Errorf("expected order left unassigned once every partner has been tried, got partner %v", *fresh.DeliveryPartnerID)
    }
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusRejected {
        t.Errorf("expected terminal status preserved as rejected, got %v", fresh.DeliveryAssignmentStatus)
    }
}

func TestTryAssignNextPartner_ConcurrentCalls_OnlyOneReassigns(t *testing.T) {
    resetDeliveryAssignmentTables(t)

    addrID := seedUserAndAddress(t, "9000000021", 19.0760, 72.8777)
    p1 := seedActivePartner(t, "8000000021", 19.0761, 72.8778)
    p2 := seedActivePartner(t, "8000000022", 19.0762, 72.8779)

    status := models.DeliveryAssignmentStatusRejected
    order := models.Order{
        UserID:                      1,
        AddressID:                   addrID,
        ItemsAmount:                 100,
        TotalAmount:                 100,
        Status:                      models.OrderStatusConfirmed,
        DeliveryPartnerID:           &p1.ID,
        DeliveryAssignmentStatus:    &status,
        DeliveryAttemptedPartnerIDs: formatIDs(p1.ID),
    }
    if err := database.DB.Create(&order).Error; err != nil {
        t.Fatalf("failed to seed order: %v", err)
    }

    const attempts = 10
    done := make(chan struct{}, attempts)
    for i := 0; i < attempts; i++ {
        go func() {
            TryAssignNextPartner(order.ID)
            done <- struct{}{}
        }()
    }
    for i := 0; i < attempts; i++ {
        <-done
    }

    fresh := reload(t, order.ID)
    if fresh.DeliveryPartnerID == nil || *fresh.DeliveryPartnerID != p2.ID {
        t.Fatalf("expected order reassigned to the only remaining eligible partner %d, got %v", p2.ID, fresh.DeliveryPartnerID)
    }
    if fresh.DeliveryAssignmentStatus == nil || *fresh.DeliveryAssignmentStatus != models.DeliveryAssignmentStatusAssigned {
        t.Errorf("expected ASSIGNED after reassignment, got %v", fresh.DeliveryAssignmentStatus)
    }
}

package services

import (
"fmt"
"log"
"math"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

const staleLocationWindow = 30 * time.Minute

var terminalOrderStatuses = []string{
models.OrderStatusDelivered,
models.OrderStatusCancelled,
models.OrderStatusReturned,
}

func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
const earthRadiusKm = 6371.0
dLat := (lat2 - lat1) * math.Pi / 180.0
dLng := (lng2 - lng1) * math.Pi / 180.0
rLat1 := lat1 * math.Pi / 180.0
rLat2 := lat2 * math.Pi / 180.0

a := math.Sin(dLat/2)*math.Sin(dLat/2) +
math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
return earthRadiusKm * c
}

type partnerCandidate struct {
ID  uint
Lat float64
Lng float64
}

func selectNearestPartner(pickupLat, pickupLng float64, candidates []partnerCandidate) *uint {
if len(candidates) == 0 {
return nil
}

bestIdx := 0
bestDist := haversineKm(pickupLat, pickupLng, candidates[0].Lat, candidates[0].Lng)
for i := 1; i < len(candidates); i++ {
d := haversineKm(pickupLat, pickupLng, candidates[i].Lat, candidates[i].Lng)
if d < bestDist {
bestDist = d
bestIdx = i
}
}

id := candidates[bestIdx].ID
return &id
}

// AutoAssignDeliveryPartner picks the nearest eligible delivery partner and
// assigns them to the given order. It's a no-op (logs and returns) if the
// order already has a partner, if the order has no pickup warehouse, or if
// no eligible partner is available.
//
// Eligibility:
//   - the partner must be active (is_active = true)
//   - the partner must NOT already be handling another order that hasn't
//     reached a terminal status (delivered/cancelled/returned) - they're
//     excluded outright, not just deprioritized, so load is spread across
//     free partners instead of piling onto someone already out delivering
//   - the partner must have a valid, recent GPS location (within
//     staleLocationWindow); a partner with a missing or stale location is
//     skipped for this pass rather than guessed at
//
// "Nearest" is measured from the order's pickup warehouse to the partner's
// latest valid location - not from the customer's delivery address, since
// the partner starts their trip at the warehouse.
func AutoAssignDeliveryPartner(orderID uint) {
var order models.Order
if err := database.DB.Preload("Warehouse").First(&order, orderID).Error; err != nil {
log.Printf("[auto-assign] order %d not found: %v", orderID, err)
return
}

if order.DeliveryPartnerID != nil {
return
}
if order.Warehouse == nil {
log.Printf("[auto-assign] order %d has no pickup warehouse, skipping auto-assign", orderID)
return
}
pickupLat, pickupLng := order.Warehouse.Lat, order.Warehouse.Lng

var partners []models.DeliveryPartner
if err := database.DB.Where("is_active = ?", true).Find(&partners).Error; err != nil {
log.Printf("[auto-assign] failed to load delivery partners: %v", err)
return
}
if len(partners) == 0 {
log.Printf("[auto-assign] no active delivery partners available for order %d", orderID)
return
}

var busyPartnerIDs []uint
if err := database.DB.Model(&models.Order{}).
Where("delivery_partner_id IS NOT NULL AND status NOT IN ?", terminalOrderStatuses).
Distinct("delivery_partner_id").
Pluck("delivery_partner_id", &busyPartnerIDs).Error; err != nil {
log.Printf("[auto-assign] failed to load busy delivery partners: %v", err)
return
}
busy := make(map[uint]bool, len(busyPartnerIDs))
for _, id := range busyPartnerIDs {
busy[id] = true
}

staleCutoff := time.Now().Add(-staleLocationWindow)
partnerByID := make(map[uint]models.DeliveryPartner, len(partners))
candidates := make([]partnerCandidate, 0, len(partners))
for _, p := range partners {
if busy[p.ID] {
continue
}
hasValidLocation := p.CurrentLat != nil && p.CurrentLng != nil &&
p.LastLocationUpdate != nil && p.LastLocationUpdate.After(staleCutoff)
if !hasValidLocation {
continue
}
partnerByID[p.ID] = p
candidates = append(candidates, partnerCandidate{
ID:  p.ID,
Lat: *p.CurrentLat,
Lng: *p.CurrentLng,
})
}

bestID := selectNearestPartner(pickupLat, pickupLng, candidates)
if bestID == nil {
log.Printf("[auto-assign] no eligible nearby delivery partner (free + valid location) for order %d", orderID)
return
}
bestPartner := partnerByID[*bestID]

if err := database.DB.Model(&models.Order{}).Where("id = ?", order.ID).Update("delivery_partner_id", bestPartner.ID).Error; err != nil {
log.Printf("[auto-assign] failed to assign partner %d to order %d: %v", bestPartner.ID, order.ID, err)
return
}

log.Printf("[auto-assign] order %d assigned to delivery partner %d (%s)", order.ID, bestPartner.ID, bestPartner.Name)

go SendPushToPartner(
bestPartner.ID,
"New delivery assigned",
fmt.Sprintf("Order #%d has been assigned to you", order.ID),
)
}
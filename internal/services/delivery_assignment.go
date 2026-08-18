package services

import (
"fmt"
"log"
"math"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// staleLocationWindow is how recent a partner's last location update must
// be for them to be considered "trackable" for nearest-match purposes.
// Partners with an older (or missing) location are still eligible, but are
// ranked after trackable ones, since we can't judge their real distance.
const staleLocationWindow = 30 * time.Minute

// haversineKm returns the great-circle distance in kilometers between two
// lat/lng points. This is the standard formula used for short-to-medium
// range distance estimates where earth curvature matters slightly.
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

// AutoAssignDeliveryPartner picks the nearest active, currently-unburdened
// delivery partner and assigns them to the given order. It's a no-op (logs
// and returns) if the order already has a partner, if the order's address
// has no coordinates, or if no active partner is available.
//
// "Nearest" is computed against the partner's last known live location.
// Partners are also deprioritized if they already have an active
// (confirmed/shipped) delivery in progress, so load is spread out rather
// than piling every new order onto whoever happens to be closest.
func AutoAssignDeliveryPartner(orderID uint) {
var order models.Order
if err := database.DB.Preload("Address").First(&order, orderID).Error; err != nil {
log.Printf("[auto-assign] order %d not found: %v", orderID, err)
return
}

if order.DeliveryPartnerID != nil {
return // already assigned, nothing to do
}
if order.Address.Lat == nil || order.Address.Lng == nil {
log.Printf("[auto-assign] order %d's address has no coordinates, skipping auto-assign", orderID)
return
}

var partners []models.DeliveryPartner
if err := database.DB.Where("is_active = ?", true).Find(&partners).Error; err != nil {
log.Printf("[auto-assign] failed to load delivery partners: %v", err)
return
}
if len(partners) == 0 {
log.Printf("[auto-assign] no active delivery partners available for order %d", orderID)
return
}

// Load count of currently active (confirmed/shipped) deliveries per
// partner, so we can prefer less-loaded partners.
type loadRow struct {
DeliveryPartnerID uint
Cnt               int64
}
var loads []loadRow
database.DB.Model(&models.Order{}).
Select("delivery_partner_id, count(*) as cnt").
Where("delivery_partner_id IS NOT NULL AND status IN ?", []string{models.OrderStatusConfirmed, models.OrderStatusShipped}).
Group("delivery_partner_id").
Scan(&loads)
loadByPartner := make(map[uint]int64, len(loads))
for _, l := range loads {
loadByPartner[l.DeliveryPartnerID] = l.Cnt
}

custLat, custLng := *order.Address.Lat, *order.Address.Lng
staleCutoff := time.Now().Add(-staleLocationWindow)

var bestPartner *models.DeliveryPartner
var bestScore float64
var bestHasFreshLocation bool

for i := range partners {
p := &partners[i]
hasFreshLocation := p.CurrentLat != nil && p.CurrentLng != nil &&
p.LastLocationUpdate != nil && p.LastLocationUpdate.After(staleCutoff)

var distanceKm float64
if hasFreshLocation {
distanceKm = haversineKm(custLat, custLng, *p.CurrentLat, *p.CurrentLng)
} else {
// Unknown distance: treat as far away so trackable partners
// are preferred, but still selectable if no one else exists.
distanceKm = math.MaxFloat64 / 2
}

// Each active delivery adds a fixed "penalty" in km-equivalent
// terms so a partner already juggling orders isn't picked over a
// slightly-farther free partner.
const loadPenaltyKm = 3.0
score := distanceKm + float64(loadByPartner[p.ID])*loadPenaltyKm

if bestPartner == nil ||
(hasFreshLocation && !bestHasFreshLocation) ||
(hasFreshLocation == bestHasFreshLocation && score < bestScore) {
bestPartner = p
bestScore = score
bestHasFreshLocation = hasFreshLocation
}
}

if bestPartner == nil {
log.Printf("[auto-assign] could not select a partner for order %d", orderID)
return
}

partnerID := bestPartner.ID
if err := database.DB.Model(&models.Order{}).Where("id = ?", order.ID).Update("delivery_partner_id", partnerID).Error; err != nil {
log.Printf("[auto-assign] failed to assign partner %d to order %d: %v", partnerID, order.ID, err)
return
}

log.Printf("[auto-assign] order %d assigned to delivery partner %d (%s)", order.ID, partnerID, bestPartner.Name)

go SendPushToPartner(
partnerID,
"New delivery assigned",
fmt.Sprintf("Order #%d has been assigned to you", order.ID),
)
}

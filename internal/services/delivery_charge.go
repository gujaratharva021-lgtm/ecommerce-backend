package services

import (
"math"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// Delivery charge pricing: a base fee plus a per-km charge from the
// nearest active warehouse, capped at maxDeliveryFee so a far-out address
// doesn't produce an absurd charge. If the address has no coordinates yet
// (older addresses saved before lat/lng was collected) or there's no
// active warehouse to measure from, we fall back to the flat charge that
// was used before distance-based pricing existed.
const (
baseDeliveryFee    = 20.0
perKmDeliveryFee   = 8.0
maxDeliveryFee     = 150.0
fallbackFlatCharge = 50.0
)

// CalculateDeliveryCharge returns the delivery charge for an order going to
// the given address coordinates, based on distance from the nearest active
// warehouse. Callers are still responsible for applying any free-delivery-
// above-threshold rule on top of this (e.g. waiving it for large orders).
func CalculateDeliveryCharge(addrLat, addrLng *float64) float64 {
if addrLat == nil || addrLng == nil {
return fallbackFlatCharge
}

var warehouses []models.Warehouse
if err := database.DB.Where("is_active = ?", true).Find(&warehouses).Error; err != nil || len(warehouses) == 0 {
return fallbackFlatCharge
}

nearestKm := math.MaxFloat64
for _, w := range warehouses {
d := utils.HaversineKm(*addrLat, *addrLng, w.Lat, w.Lng)
if d < nearestKm {
nearestKm = d
}
}

charge := baseDeliveryFee + perKmDeliveryFee*nearestKm
if charge > maxDeliveryFee {
charge = maxDeliveryFee
}
// Round to 2 decimal places.
return math.Round(charge*100) / 100
}

package handlers

import (
"math"
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// ---------------------------------------------------------------------------
// Serviceability check (public)
// ---------------------------------------------------------------------------

const earthRadiusKm = 6371.0

// haversineDistanceKm returns the great-circle distance between two
// lat/lng points in kilometers.
func haversineDistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
lat1Rad := lat1 * math.Pi / 180
lat2Rad := lat2 * math.Pi / 180
deltaLat := (lat2 - lat1) * math.Pi / 180
deltaLng := (lng2 - lng1) * math.Pi / 180

a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
math.Cos(lat1Rad)*math.Cos(lat2Rad)*
math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

return earthRadiusKm * c
}

// CheckServiceability godoc
// GET /api/v1/serviceability?lat=&lng=
// Finds the nearest active warehouse to the given coordinates and reports
// whether it is within that warehouse's service radius.
func CheckServiceability(c *gin.Context) {
latStr := c.Query("lat")
lngStr := c.Query("lng")

if latStr == "" || lngStr == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng query params are required"})
return
}

lat, err := strconv.ParseFloat(latStr, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat"})
return
}
lng, err := strconv.ParseFloat(lngStr, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng"})
return
}

var warehouses []models.Warehouse
if err := database.DB.Where("is_active = ?", true).Find(&warehouses).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load warehouses"})
return
}

if len(warehouses) == 0 {
c.JSON(http.StatusOK, gin.H{
"serviceable": false,
"message":     "No active warehouses available",
})
return
}

var nearest models.Warehouse
nearestDistance := math.MaxFloat64

for _, wh := range warehouses {
dist := haversineDistanceKm(lat, lng, wh.Lat, wh.Lng)
if dist < nearestDistance {
nearestDistance = dist
nearest = wh
}
}

serviceable := nearestDistance <= nearest.ServiceRadiusKm

response := gin.H{
"serviceable":     serviceable,
"distance_km":     math.Round(nearestDistance*100) / 100,
"nearest_warehouse": gin.H{
"id":                nearest.ID,
"name":              nearest.Name,
"city":              nearest.City,
"service_radius_km": nearest.ServiceRadiusKm,
},
}

if !serviceable {
response["message"] = "Sorry, we don't deliver to this location yet"
}

c.JSON(http.StatusOK, response)
}
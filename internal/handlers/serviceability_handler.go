package handlers

import (
"math"
"net/http"
"sort"
"strconv"

"github.com/gin-gonic/gin"
	"log"
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

// FindNearestWarehouse returns the nearest active warehouse to the given
// coordinates and the distance in km. Shared by the public /serviceability
// endpoint and checkout (to assign + validate deliverability before an
// order is placed). Returns (nil, 0, nil) if there are no active warehouses.
func FindNearestWarehouse(lat, lng float64) (*models.Warehouse, float64, error) {
var warehouses []models.Warehouse
if err := database.DB.Where("is_active = ? AND status = ?", true, "open").Find(&warehouses).Error; err != nil {
return nil, 0, err
}
if len(warehouses) == 0 {
return nil, 0, nil
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
return &nearest, nearestDistance, nil
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

nearest, _, err := FindNearestWarehouse(lat, lng)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load warehouses"})
return
}
if nearest == nil {
c.JSON(http.StatusOK, gin.H{
"serviceable": false,
"message":     "No active warehouses available",
})
return
}

// Check every active warehouse in ascending order of distance, not just
// the single nearest one by center-point - a customer can fall outside
// the nearest warehouse's polygon while still being genuinely inside a
// slightly-farther warehouse's actual service area (Bug#10: Euclidean
// pre-selection wrongly rejected addresses serviceable by another
// warehouse).
type whCandidate struct {
ID              uint
Name            string
City            string
Lat             float64
Lng             float64
ServiceRadiusKm float64
HasPolygon      bool
ContainsPoint   bool
}
var candidates []whCandidate
if err := database.DB.Raw(
`SELECT id, name, city, lat, lng, service_radius_km,
service_area IS NOT NULL AS has_polygon,
CASE WHEN service_area IS NOT NULL THEN ST_Contains(service_area, ST_SetSRID(ST_MakePoint(?, ?), 4326)) ELSE false END AS contains_point
FROM warehouses WHERE is_active = ? AND status = ?`,
lng, lat, true, "open",
).Scan(&candidates).Error; err != nil {
log.Printf("serviceability check failed: %v", err)
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check serviceability"})
return
}
if len(candidates) == 0 {
c.JSON(http.StatusOK, gin.H{
"serviceable": false,
"message":     "No active warehouses available",
})
return
}

sort.Slice(candidates, func(i, j int) bool {
return haversineDistanceKm(lat, lng, candidates[i].Lat, candidates[i].Lng) < haversineDistanceKm(lat, lng, candidates[j].Lat, candidates[j].Lng)
})

var serviceable bool
method := "radius"
best := candidates[0]
bestDistance := haversineDistanceKm(lat, lng, best.Lat, best.Lng)
for _, cand := range candidates {
d := haversineDistanceKm(lat, lng, cand.Lat, cand.Lng)
var ok bool
m := "radius"
if cand.HasPolygon {
ok = cand.ContainsPoint
m = "polygon"
} else {
ok = d <= cand.ServiceRadiusKm
}
if ok {
serviceable = true
method = m
best = cand
bestDistance = d
break
}
}

response := gin.H{
"serviceable": serviceable,
"distance_km": math.Round(bestDistance*100) / 100,
"nearest_warehouse": gin.H{
"id":                best.ID,
"name":              best.Name,
"city":              best.City,
"service_radius_km": best.ServiceRadiusKm,
},
"method":      method,
}
if !serviceable {
response["message"] = "Sorry, we don't deliver to this location yet"
}

c.JSON(http.StatusOK, response)
}

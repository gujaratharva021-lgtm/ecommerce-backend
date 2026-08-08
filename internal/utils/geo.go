package utils

import "math"

// EarthRadiusKm is the mean radius of the Earth in kilometers, used by
// HaversineKm below.
const EarthRadiusKm = 6371.0

// HaversineKm returns the great-circle distance in kilometers between two
// lat/lng points. This is the standard formula used for short-to-medium
// range distance estimates where earth curvature matters slightly.
//
// L-12: this used to be implemented twice (once in
// internal/services/delivery_assignment.go as haversineKm, once in
// internal/handlers/serviceability_handler.go as haversineDistanceKm) with
// a duplicated earthRadiusKm constant. Both callers now use this single
// shared implementation.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	rLat1 := lat1 * math.Pi / 180.0
	rLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return EarthRadiusKm * c
}

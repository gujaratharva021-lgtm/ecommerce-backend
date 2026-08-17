package services

import "testing"

func TestHaversineKm_SamePoint(t *testing.T) {
d := haversineKm(19.0760, 72.8777, 19.0760, 72.8777)
if d != 0 {
t.Fatalf("expected 0 km for identical points, got %f", d)
}
}

func TestHaversineKm_KnownDistance(t *testing.T) {
d := haversineKm(19.0596, 72.8295, 19.1197, 72.8468)
if d < 5 || d > 15 {
t.Fatalf("expected distance in the ~5-15km range, got %f", d)
}
}

func TestSelectNearestPartner_NoCandidates(t *testing.T) {
got := selectNearestPartner(19.0760, 72.8777, nil)
if got != nil {
t.Fatalf("expected nil for empty candidates, got %v", *got)
}
}

func TestSelectNearestPartner_PicksClosest(t *testing.T) {
pickupLat, pickupLng := 19.0760, 72.8777

candidates := []partnerCandidate{
{ID: 1, Lat: 19.2000, Lng: 72.9700},
{ID: 2, Lat: 19.0770, Lng: 72.8780},
{ID: 3, Lat: 18.9000, Lng: 72.8000},
}

got := selectNearestPartner(pickupLat, pickupLng, candidates)
if got == nil {
t.Fatal("expected a partner to be selected, got nil")
}
if *got != 2 {
t.Fatalf("expected nearest partner ID 2, got %d", *got)
}
}

func TestSelectNearestPartner_SingleCandidate(t *testing.T) {
candidates := []partnerCandidate{
{ID: 42, Lat: 19.10, Lng: 72.90},
}
got := selectNearestPartner(19.0760, 72.8777, candidates)
if got == nil || *got != 42 {
t.Fatalf("expected partner 42, got %v", got)
}
}
package radial

import "testing"

func TestZoneOverlapsCandidateKeepsCenterPointBehaviorForZeroRadius(t *testing.T) {
	zone := Zone{InnerRadius: 10, OuterRadius: 20}

	tests := []struct {
		name     string
		distance float64
		want     bool
	}{
		{name: "inside_center_zone", distance: 10.1, want: true},
		{name: "inside_outer_zone", distance: 19.9, want: true},
		{name: "on_inner_boundary", distance: 10, want: true},
		{name: "on_outer_boundary", distance: 20, want: false},
		{name: "outside_zone", distance: 9.9, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zoneOverlapsCandidate(zone, tt.distance, 0); got != tt.want {
				t.Fatalf("zoneOverlapsCandidate(%v, 0) = %v, want %v", tt.distance, got, tt.want)
			}
		})
	}
}

func TestZoneOverlapsCandidateUsesCircularExtent(t *testing.T) {
	zone := Zone{InnerRadius: 10, OuterRadius: 20}

	tests := []struct {
		name     string
		distance float64
		radius   float64
		want     bool
	}{
		{name: "outer_edge_reaches_inner_radius", distance: 7, radius: 3, want: true},
		{name: "inner_edge_reaches_outer_radius", distance: 25, radius: 6, want: true},
		{name: "fully_before_inner_radius", distance: 6, radius: 3, want: false},
		{name: "fully_beyond_outer_radius", distance: 25, radius: 5, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := zoneOverlapsCandidate(zone, tt.distance, tt.radius); got != tt.want {
				t.Fatalf("zoneOverlapsCandidate(%v, %v) = %v, want %v", tt.distance, tt.radius, got, tt.want)
			}
		})
	}
}

func TestFillOverlapsCandidateUsesCircularExtent(t *testing.T) {
	if !fillOverlapsCandidate(10, 12, 3) {
		t.Fatal("expected fill to hit candidate whose inner extent is inside radius")
	}
	if fillOverlapsCandidate(10, 12, 2) {
		t.Fatal("expected fill to miss candidate whose inner extent is on radius boundary")
	}
}

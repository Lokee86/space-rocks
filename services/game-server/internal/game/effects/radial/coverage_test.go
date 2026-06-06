package radial

import "testing"

func TestZoneContainsDistance(t *testing.T) {
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
			if got := zoneContainsDistance(zone, tt.distance); got != tt.want {
				t.Fatalf("zoneContainsDistance(%v) = %v, want %v", tt.distance, got, tt.want)
			}
		})
	}
}

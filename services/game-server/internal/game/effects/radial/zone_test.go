package radial

import "testing"

const floatEpsilon = 1e-9

func TestBuildZonesAnnularWaveSimultaneous(t *testing.T) {
	spec := Spec{
		CoverageMode:       CoverageAnnularWave,
		ExpirationMode:     ExpirationSimultaneous,
		ZoneCount:          4,
		ZoneWidth:          10,
		ZoneSpawnSeconds:   0.1,
		TickSeconds:        0.1,
		TotalSeconds:       0.4,
		ZoneLifetimeSeconds: 0,
	}

	zones := buildZones(spec)
	if got, want := len(zones), 4; got != want {
		t.Fatalf("len(zones) = %d, want %d", got, want)
	}

	assertZone := func(idx int, inner, outer, startsAt, expiresAt, nextTickAt float64) {
		t.Helper()
		zone := zones[idx]
		if zone.Index != idx {
			t.Fatalf("zone %d index = %d, want %d", idx+1, zone.Index, idx)
		}
		if !almostEqual(zone.InnerRadius, inner) {
			t.Fatalf("zone %d inner radius = %v, want %v", idx+1, zone.InnerRadius, inner)
		}
		if !almostEqual(zone.OuterRadius, outer) {
			t.Fatalf("zone %d outer radius = %v, want %v", idx+1, zone.OuterRadius, outer)
		}
		if !almostEqual(zone.StartsAt, startsAt) {
			t.Fatalf("zone %d startsAt = %v, want %v", idx+1, zone.StartsAt, startsAt)
		}
		if !almostEqual(zone.ExpiresAt, expiresAt) {
			t.Fatalf("zone %d expiresAt = %v, want %v", idx+1, zone.ExpiresAt, expiresAt)
		}
		if !almostEqual(zone.NextTickAt, nextTickAt) {
			t.Fatalf("zone %d nextTickAt = %v, want %v", idx+1, zone.NextTickAt, nextTickAt)
		}
	}

	assertZone(0, 0, 10, 0, 0.4, 0)
	assertZone(1, 10, 20, 0.1, 0.4, 0.1)
	assertZone(2, 20, 30, 0.2, 0.4, 0.2)
	assertZone(3, 30, 40, 0.3, 0.4, 0.3)
}

func TestBuildZonesAnnularWaveSequential(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageAnnularWave,
		ExpirationMode:      ExpirationSequential,
		ZoneCount:           4,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.4,
		ZoneLifetimeSeconds: 0.15,
	}

	zones := buildZones(spec)
	if got, want := len(zones), 4; got != want {
		t.Fatalf("len(zones) = %d, want %d", got, want)
	}

	assertZoneSequential := func(idx int, expectedExpiresAt float64) {
		t.Helper()
		if !almostEqual(zones[idx].ExpiresAt, expectedExpiresAt) {
			t.Fatalf("zone %d expiresAt = %v, want %v", idx+1, zones[idx].ExpiresAt, expectedExpiresAt)
		}
	}

	assertZoneSequential(0, 0.15)
	assertZoneSequential(1, 0.25)
	assertZoneSequential(2, 0.35)
	assertZoneSequential(3, 0.45)
}

func TestSequentialZonesKeepOuterZonesActiveAfterInnerExpire(t *testing.T) {
	spec := Spec{
		CoverageMode:        CoverageAnnularWave,
		ExpirationMode:      ExpirationSequential,
		ZoneCount:           3,
		ZoneWidth:           10,
		ZoneSpawnSeconds:    0.1,
		TickSeconds:         0.1,
		TotalSeconds:        0.3,
		ZoneLifetimeSeconds: 0.15,
	}

	zones := buildZones(spec)
	if !(zones[2].ExpiresAt > zones[0].ExpiresAt) {
		t.Fatalf("expected outer zone to expire later than inner zone, got inner=%v outer=%v", zones[0].ExpiresAt, zones[2].ExpiresAt)
	}
	if !almostEqual(zones[0].ExpiresAt, 0.15) {
		t.Fatalf("zone 1 expiresAt = %v, want %v", zones[0].ExpiresAt, 0.15)
	}
	if !almostEqual(zones[2].ExpiresAt, 0.35) {
		t.Fatalf("zone 3 expiresAt = %v, want %v", zones[2].ExpiresAt, 0.35)
	}
}

func almostEqual(a, b float64) bool {
	if a > b {
		return a-b < floatEpsilon
	}
	return b-a < floatEpsilon
}

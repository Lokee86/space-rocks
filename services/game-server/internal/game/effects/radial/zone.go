package radial

type Zone struct {
	Index       int
	InnerRadius float64
	OuterRadius float64
	StartsAt    float64
	ExpiresAt   float64
	NextTickAt  float64
}

func buildZones(spec Spec) []Zone {
	if spec.CoverageMode != CoverageAnnularWave && spec.CoverageMode != CoverageExpandingFill {
		return nil
	}
	if spec.ExpirationMode != ExpirationSimultaneous && spec.ExpirationMode != ExpirationSequential {
		return nil
	}
	if spec.ZoneCount <= 0 || spec.ZoneWidth <= 0 || spec.ZoneSpawnSeconds < 0 || spec.TotalSeconds <= 0 {
		return nil
	}

	zones := make([]Zone, 0, spec.ZoneCount)
	for i := 0; i < spec.ZoneCount; i++ {
		startsAt := float64(i) * spec.ZoneSpawnSeconds
		if startsAt > spec.TotalSeconds {
			return nil
		}

		innerRadius := float64(i) * spec.ZoneWidth
		outerRadius := innerRadius + spec.ZoneWidth
		zones = append(zones, Zone{
			Index:       i,
			InnerRadius: innerRadius,
			OuterRadius: outerRadius,
			StartsAt:    startsAt,
			ExpiresAt:   zoneExpirationAt(spec, startsAt),
			NextTickAt:  startsAt,
		})
	}

	return zones
}

func zoneExpirationAt(spec Spec, startsAt float64) float64 {
	if spec.ExpirationMode == ExpirationSequential {
		return startsAt + spec.ZoneLifetimeSeconds
	}
	return spec.TotalSeconds
}

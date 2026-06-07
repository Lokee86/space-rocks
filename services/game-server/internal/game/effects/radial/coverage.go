package radial

import "math"

func zoneOverlapsCandidate(zone Zone, distance float64, radius float64) bool {
	return candidateOuterDistance(distance, radius) >= zone.InnerRadius &&
		candidateInnerDistance(distance, radius) < zone.OuterRadius
}

func fillOverlapsCandidate(fillRadius float64, distance float64, radius float64) bool {
	return candidateInnerDistance(distance, radius) < fillRadius
}

func candidateOuterDistance(distance float64, radius float64) float64 {
	return distance + radius
}

func candidateInnerDistance(distance float64, radius float64) float64 {
	return math.Max(distance-radius, 0)
}

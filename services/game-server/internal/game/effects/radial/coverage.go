package radial

func zoneContainsDistance(zone Zone, distance float64) bool {
	return distance >= zone.InnerRadius && distance < zone.OuterRadius
}

func fillContainsDistance(radius float64, distance float64) bool {
	return distance >= 0 && distance < radius
}

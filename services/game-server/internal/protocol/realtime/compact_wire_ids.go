package realtime

import "strconv"

func compactWirePackPrefixedNumericID(value any, prefix string) any {
	id, ok := value.(string)
	if !ok {
		return value
	}

	if len(id) < len(prefix) || id[:len(prefix)] != prefix {
		return value
	}

	parsed, err := strconv.Atoi(id[len(prefix):])
	if err != nil {
		return value
	}

	return parsed
}

func compactWirePackAsteroidID(value any) any {
	return compactWirePackPrefixedNumericID(value, "asteroid-")
}

func compactWirePackBulletID(value any) any {
	return compactWirePackPrefixedNumericID(value, "bullet-")
}

func compactWirePackPickupID(value any) any {
	return compactWirePackPrefixedNumericID(value, "pickup-")
}

func compactWirePackShipID(value any) any {
	return compactWirePackPrefixedNumericID(value, "ship-")
}

func compactWirePackTableID(value any) any {
	return compactWirePackPrefixedNumericID(value, "table-")
}

func compactWirePackPlayerID(value any) any {
	return compactWirePackPrefixedNumericID(value, "player-")
}

func compactWirePackPresentationEventID(value any) any {
	return compactWirePackPrefixedNumericID(value, "presentation-event-")
}

func compactWirePackEventBatchID(value any) any {
	return compactWirePackPrefixedNumericID(value, "event-batch-")
}

func compactWirePackKnownTaggedID(value any) any {
	id, ok := value.(string)
	if !ok {
		return value
	}

	switch {
	case len(id) > len("player-") && id[:len("player-")] == "player-":
		if parsed, err := strconv.Atoi(id[len("player-"):]); err == nil {
			return []any{"p", parsed}
		}
	case len(id) > len("bullet-") && id[:len("bullet-")] == "bullet-":
		if parsed, err := strconv.Atoi(id[len("bullet-"):]); err == nil {
			return []any{"b", parsed}
		}
	case len(id) > len("asteroid-") && id[:len("asteroid-")] == "asteroid-":
		if parsed, err := strconv.Atoi(id[len("asteroid-"):]); err == nil {
			return []any{"a", parsed}
		}
	case len(id) > len("pickup-") && id[:len("pickup-")] == "pickup-":
		if parsed, err := strconv.Atoi(id[len("pickup-"):]); err == nil {
			return []any{"pk", parsed}
		}
	case len(id) > len("ship-") && id[:len("ship-")] == "ship-":
		if parsed, err := strconv.Atoi(id[len("ship-"):]); err == nil {
			return []any{"s", parsed}
		}
	case len(id) > len("table-") && id[:len("table-")] == "table-":
		if parsed, err := strconv.Atoi(id[len("table-"):]); err == nil {
			return []any{"tbl", parsed}
		}
	case len(id) > len("presentation-event-") && id[:len("presentation-event-")] == "presentation-event-":
		if parsed, err := strconv.Atoi(id[len("presentation-event-"):]); err == nil {
			return []any{"pe", parsed}
		}
	case len(id) > len("event-batch-") && id[:len("event-batch-")] == "event-batch-":
		if parsed, err := strconv.Atoi(id[len("event-batch-"):]); err == nil {
			return []any{"eb", parsed}
		}
	}

	return value
}

func compactWirePackSourceID(sourceType any, value any) any {
	switch sourceType {
	case "player":
		return compactWirePackPlayerID(value)
	case "ship":
		return compactWirePackShipID(value)
	case "projectile", "bullet":
		return compactWirePackBulletID(value)
	case "asteroid":
		return compactWirePackAsteroidID(value)
	case "pickup":
		return compactWirePackPickupID(value)
	default:
		return compactWirePackKnownTaggedID(value)
	}
}

func compactWirePackTargetID(targetKind any, value any) any {
	switch targetKind {
	case "player":
		return compactWirePackPlayerID(value)
	case "ship":
		return compactWirePackShipID(value)
	case "bullet", "projectile":
		return compactWirePackBulletID(value)
	case "asteroid":
		return compactWirePackAsteroidID(value)
	case "pickup":
		return compactWirePackPickupID(value)
	default:
		return compactWirePackKnownTaggedID(value)
	}
}

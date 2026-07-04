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

func compactWirePackPlayerID(value any) any {
	return compactWirePackPrefixedNumericID(value, "player-")
}

func compactWirePackPresentationEventID(value any) any {
	return compactWirePackPrefixedNumericID(value, "presentation-event-")
}

func compactWirePackEventBatchID(value any) any {
	return compactWirePackPrefixedNumericID(value, "event-batch-")
}

package realtime

func compactWirePackAsteroids(packet map[string]any) map[string]any {
	switch packet["t"] {
	case "wf":
		return compactWirePackWorldFullAsteroids(packet)
	case "wd":
		return compactWirePackWorldDeltaAsteroids(packet)
	default:
		return packet
	}
}

func compactWirePackWorldFullAsteroids(packet map[string]any) map[string]any {
	asteroids, ok := packet["asteroids"].([]any)
	if !ok {
		return packet
	}

	packed := make([]any, len(asteroids))
	for i, asteroid := range asteroids {
		packed[i] = compactWirePackAsteroidRecord(asteroid)
	}

	next := compactWireClonePacket(packet)
	next["asteroids"] = packed
	return next
}

func compactWirePackWorldDeltaAsteroids(packet map[string]any) map[string]any {
	next := packet
	cloned := false

	if creates, ok := packet["ac"].([]any); ok {
		packedCreates := make([]any, len(creates))
		for i, asteroid := range creates {
			packedCreates[i] = compactWirePackAsteroidRecord(asteroid)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["ac"] = packedCreates
	}

	if updates, ok := packet["au"].([]any); ok {
		packedUpdates := make([]any, len(updates))
		for i, asteroid := range updates {
			packedUpdates[i] = compactWirePackAsteroidMovementUpdate(asteroid)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["au"] = packedUpdates
	}

	if deletes, ok := packet["ax"].([]any); ok {
		packedDeletes := make([]any, len(deletes))
		for i, asteroidID := range deletes {
			packedDeletes[i] = compactWirePackAsteroidID(asteroidID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["ax"] = packedDeletes
	}

	return next
}

func compactWireClonePacket(packet map[string]any) map[string]any {
	next := make(map[string]any, len(packet))
	for key, value := range packet {
		next[key] = value
	}
	return next
}

func compactWirePackAsteroidRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	return []any{
		compactWirePackAsteroidID(record["i"]),
		record["x"],
		record["y"],
		record["sz"],
		record["h"],
		record["sl"],
		record["v"],
	}
}

func compactWirePackAsteroidMovementUpdate(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	tuple := []any{compactWirePackAsteroidID(record["i"])}
	if x, ok := record["x"]; ok {
		tuple = append(tuple, x)
		if y, ok := record["y"]; ok {
			tuple = append(tuple, y)
		}
		return tuple
	}
	if y, ok := record["y"]; ok {
		return append(tuple, nil, y)
	}
	return tuple
}




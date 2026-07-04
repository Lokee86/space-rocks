package realtime

func compactWirePackShips(packet map[string]any) map[string]any {
	switch packet["t"] {
	case "wf":
		return compactWirePackWorldFullShips(packet)
	case "wd":
		return compactWirePackWorldDeltaShips(packet)
	default:
		return packet
	}
}

func compactWirePackWorldFullShips(packet map[string]any) map[string]any {
	ships, ok := packet["ships"].([]any)
	if !ok {
		return packet
	}

	packed := make([]any, len(ships))
	for i, ship := range ships {
		packed[i] = compactWirePackShipRecord(ship)
	}

	next := compactWireClonePacket(packet)
	next["ships"] = packed
	return next
}

func compactWirePackWorldDeltaShips(packet map[string]any) map[string]any {
	next := packet
	cloned := false

	if creates, ok := packet["sc"].([]any); ok {
		packedCreates := make([]any, len(creates))
		for i, ship := range creates {
			packedCreates[i] = compactWirePackShipRecord(ship)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["sc"] = packedCreates
	}

	if updates, ok := packet["su"].([]any); ok {
		packedUpdates := make([]any, len(updates))
		for i, ship := range updates {
			packedUpdates[i] = compactWirePackShipMovementUpdate(ship)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["su"] = packedUpdates
	}

	if deletes, ok := packet["sx"].([]any); ok {
		packedDeletes := make([]any, len(deletes))
		for i, shipID := range deletes {
			packedDeletes[i] = compactWirePackPlayerID(shipID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["sx"] = packedDeletes
	}

	return next
}

func compactWirePackShipRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	return []any{
		compactWirePackPlayerID(record["i"]),
		record["st"],
		record["x"],
		record["y"],
		record["r"],
		record["h"],
		record["sh"],
		record["th"],
		record["tk"],
		compactWirePackTargetID(record["tk"], record["tid"]),
	}
}

func compactWirePackShipMovementUpdate(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	tuple := []any{compactWirePackPlayerID(record["i"]), nil, nil, nil, nil}

	if x, ok := record["x"]; ok {
		tuple[1] = x
	}
	if y, ok := record["y"]; ok {
		tuple[2] = y
	}
	if rotation, ok := record["r"]; ok {
		tuple[3] = rotation
	}
	if thrusting, ok := record["th"]; ok {
		tuple[4] = thrusting
	}

	for len(tuple) > 1 && tuple[len(tuple)-1] == nil {
		tuple = tuple[:len(tuple)-1]
	}

	return tuple
}

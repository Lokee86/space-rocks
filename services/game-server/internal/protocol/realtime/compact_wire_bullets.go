package realtime

func compactWirePackBullets(packet map[string]any) map[string]any {
	switch packet["t"] {
	case "wf":
		return compactWirePackWorldFullBullets(packet)
	case "wd":
		return compactWirePackWorldDeltaBullets(packet)
	case "bd":
		return compactWirePackBulletDelta(packet)
	default:
		return packet
	}
}

func compactWirePackWorldFullBullets(packet map[string]any) map[string]any {
	bullets, ok := packet["bullets"].([]any)
	if !ok {
		return packet
	}

	packed := make([]any, len(bullets))
	for i, bullet := range bullets {
		packed[i] = compactWirePackBulletRecord(bullet)
	}

	next := compactWireClonePacket(packet)
	next["bullets"] = packed
	return next
}

func compactWirePackWorldDeltaBullets(packet map[string]any) map[string]any {
	next := packet
	cloned := false

	if creates, ok := packet["bc"].([]any); ok {
		packedCreates := make([]any, len(creates))
		for i, bullet := range creates {
			packedCreates[i] = compactWirePackBulletRecord(bullet)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["bc"] = packedCreates
	}

	if updates, ok := packet["bu"].([]any); ok {
		packedUpdates := make([]any, len(updates))
		for i, bullet := range updates {
			packedUpdates[i] = compactWirePackBulletMovementUpdate(bullet)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["bu"] = packedUpdates
	}

	if deletes, ok := packet["bx"].([]any); ok {
		packedDeletes := make([]any, len(deletes))
		for i, bulletID := range deletes {
			packedDeletes[i] = compactWirePackBulletID(bulletID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["bx"] = packedDeletes
	}

	return next
}

func compactWirePackBulletRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	return []any{
		compactWirePackBulletID(record["i"]),
		compactWirePackPlayerID(record["oi"]),
		record["x"],
		record["y"],
		record["r"],
		record["wid"],
		record["pt"],
	}
}

func compactWirePackBulletMovementUpdate(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	tuple := []any{compactWirePackBulletID(record["i"]), nil, nil, nil}

	if x, ok := record["x"]; ok {
		tuple[1] = x
	}
	if y, ok := record["y"]; ok {
		tuple[2] = y
	}
	if rotation, ok := record["r"]; ok {
		tuple[3] = rotation
	}

	for len(tuple) > 1 && tuple[len(tuple)-1] == nil {
		tuple = tuple[:len(tuple)-1]
	}

	return tuple
}

func compactWirePackBulletDelta(packet map[string]any) map[string]any {
	return compactWirePackWorldDeltaBullets(packet)
}

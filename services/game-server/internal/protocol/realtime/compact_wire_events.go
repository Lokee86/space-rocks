package realtime

func compactWirePackEvents(packet map[string]any) map[string]any {
	if packet["t"] != "eb" {
		return packet
	}

	next := packet
	cloned := false

	if batchID, ok := packet["bid"]; ok {
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["bid"] = compactWirePackEventBatchID(batchID)
	}

	if events, ok := packet["ev"].([]any); ok {
		packedEvents := make([]any, len(events))
		for i, event := range events {
			packedEvents[i] = compactWirePackEventRecord(event)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["ev"] = packedEvents
	}

	return next
}

func compactWirePackEventRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	switch record["t"] {
	case "bb":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), record["x"], record["y"]}
	case "shd":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), compactWirePackPlayerID(record["pid"]), record["lv"], record["rd"], record["x"], record["y"]}
	case "dmg":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), record["srct"], compactWirePackSourceID(record["srct"], record["src"]), record["fx"], record["amt"], record["x"], record["y"]}
	case "dots":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), record["srct"], compactWirePackSourceID(record["srct"], record["src"]), record["fx"], record["amt"]}
	case "dott":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), record["srct"], compactWirePackSourceID(record["srct"], record["src"]), record["fx"], record["amt"], record["x"], record["y"]}
	case "rfx":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), record["srct"], compactWirePackSourceID(record["srct"], record["src"]), record["fx"], record["x"], record["y"]}
	case "pcol":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), compactWirePackPlayerID(record["pid"]), compactWirePackPickupID(record["pkid"]), record["pkt"], record["x"], record["y"]}
	case "pea":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), compactWirePackPlayerID(record["pid"]), compactWirePackPickupID(record["pkid"]), record["pkt"], record["fx"], record["amt"], record["lva"]}
	case "pexp":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), compactWirePackPickupID(record["pkid"]), record["pkt"], record["x"], record["y"]}
	case "pdr":
		return []any{record["t"], compactWirePackPresentationEventID(record["ei"]), compactWirePackPickupID(record["pkid"]), record["pkt"], record["srct"], compactWirePackSourceID(record["srct"], record["src"]), compactWirePackTableID(record["tbl"]), record["x"], record["y"]}
	default:
		return record
	}
}

package realtime

var compactWireSessionPlayerUpdateFieldAliases = map[string]string{
	"score":             "sco",
	"lives":             "lv",
	"respawn_cooldown":  "rcd",
	"ship_type":         "st",
	"primary_weapon_id": "pwid",
	"primary_ammo_policy": "pap",
	"secondary_weapon_id": "swid",
	"secondary_ammo_policy": "sap",
	"spawn_x":           "spx",
	"spawn_y":           "spy",
}

var compactWireSessionLifecycleFieldAliases = map[string]string{
	"player_id": "pid",
	"status":    "stat",
}

func compactWirePackSessionPlayers(packet map[string]any) map[string]any {
	switch packet["t"] {
	case "sf":
		return compactWirePackSessionFullPlayers(packet)
	case "sd":
		return compactWirePackSessionDeltaPlayers(packet)
	default:
		return packet
	}
}

func compactWirePackSessionFullPlayers(packet map[string]any) map[string]any {
	next := packet
	cloned := false

	if players, ok := packet["pl"].([]any); ok {
		packed := make([]any, len(players))
		for i, player := range players {
			packed[i] = compactWirePackSessionPlayerRecord(player)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["pl"] = packed
	}

	if lifecycles, ok := packet["plc"].([]any); ok {
		packed := make([]any, len(lifecycles))
		for i, lifecycle := range lifecycles {
			packed[i] = compactWirePackSessionLifecycleRecord(lifecycle)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plc"] = packed
	}

	if updates, ok := packet["plu"].([]any); ok {
		packed := make([]any, len(updates))
		for i, lifecycle := range updates {
			packed[i] = compactWirePackSessionLifecycleRecord(lifecycle)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plu"] = packed
	}

	if deletes, ok := packet["plx"].([]any); ok {
		packed := make([]any, len(deletes))
		for i, playerID := range deletes {
			packed[i] = compactWirePackPlayerID(playerID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plx"] = packed
	}

	return next
}

func compactWirePackSessionDeltaPlayers(packet map[string]any) map[string]any {
	next := packet
	cloned := false

	if lifecycles, ok := packet["plc"].([]any); ok {
		packed := make([]any, len(lifecycles))
		for i, lifecycle := range lifecycles {
			packed[i] = compactWirePackSessionLifecycleRecord(lifecycle)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plc"] = packed
	}

	if updates, ok := packet["plu"].([]any); ok {
		packed := make([]any, len(updates))
		for i, lifecycle := range updates {
			packed[i] = compactWirePackSessionLifecycleRecord(lifecycle)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plu"] = packed
	}

	if deletes, ok := packet["plx"].([]any); ok {
		packed := make([]any, len(deletes))
		for i, playerID := range deletes {
			packed[i] = compactWirePackPlayerID(playerID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["plx"] = packed
	}

	if updates, ok := packet["psu"].([]any); ok {
		packedUpdates := make([]any, len(updates))
		for i, player := range updates {
			packedUpdates[i] = compactWirePackSessionPlayerUpdate(player)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["psu"] = packedUpdates
	}

	if deletes, ok := packet["psx"].([]any); ok {
		packedDeletes := make([]any, len(deletes))
		for i, playerID := range deletes {
			packedDeletes[i] = compactWirePackPlayerID(playerID)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["psx"] = packedDeletes
	}

	if players, ok := packet["pl"].([]any); ok {
		packedPlayers := make([]any, len(players))
		for i, player := range players {
			packedPlayers[i] = compactWirePackSessionPlayerRecord(player)
		}
		if !cloned {
			next = compactWireClonePacket(packet)
			cloned = true
		}
		next["pl"] = packedPlayers
	}

	return next
}

func compactWirePackSessionLifecycleRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	return []any{
		compactWirePackPlayerID(record["pid"]),
		record["stat"],
	}
}

func compactWirePackSessionPlayerRecord(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	return []any{
		compactWirePackPlayerID(record["i"]),
		record["st"],
		record["sco"],
		record["lv"],
		record["rcd"],
		record["pwid"],
		record["pap"],
		record["swid"],
		record["sap"],
		record["spx"],
		record["spy"],
	}
}

func compactWirePackSessionPlayerUpdate(value any) any {
	record, ok := value.(map[string]any)
	if !ok {
		return value
	}

	tuple := []any{compactWirePackPlayerID(record["i"])}
	for _, field := range []string{"score", "lives", "respawn_cooldown", "ship_type", "primary_weapon_id", "primary_ammo_policy", "secondary_weapon_id", "secondary_ammo_policy", "spawn_x", "spawn_y"} {
		alias := compactWireSessionPlayerUpdateFieldAliases[field]
		if alias == "" {
			alias = field
		}
		if value, ok := record[alias]; ok {
			tuple = append(tuple, alias, value)
			continue
		}
		if value, ok := record[field]; ok {
			tuple = append(tuple, alias, value)
		}
	}
	for key, value := range record {
		if key == "i" || key == "id" {
			continue
		}
		if _, known := compactWireSessionPlayerUpdateFieldAliases[key]; known {
			continue
		}
		if key == "sco" || key == "lv" || key == "rcd" || key == "st" || key == "pwid" || key == "pap" || key == "swid" || key == "sap" || key == "spx" || key == "spy" {
			continue
		}
		tuple = append(tuple, key, value)
	}
	return tuple
}

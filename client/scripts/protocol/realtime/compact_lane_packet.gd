extends RefCounted

const LaneMetadata = preload("res://scripts/protocol/realtime/lane_metadata.gd")

const _KEY_MAP := {
	"t": "type",
	"l": "lane",
	"q": "sequence",
	"b": "baseline_id",
	"bq": "baseline_sequence",
	"sid": "snapshot_id",
	"ms": "server_sent_msec",
	"k": "snapshot_kind",
	"ci": "chunk_index",
	"cc": "chunk_count",
	"fc": "is_final_chunk",
	"sc": "ship_creates",
	"su": "ship_updates",
	"sx": "ship_deletes",
	"bc": "bullet_creates",
	"bu": "bullet_updates",
	"bx": "bullet_deletes",
	"ac": "asteroid_creates",
	"au": "asteroid_updates",
	"ax": "asteroid_deletes",
	"pc": "pickup_creates",
	"pu": "pickup_updates",
	"px": "pickup_deletes",
	"rc": "receiver_creates",
	"ru": "receiver_updates",
	"rx": "receiver_deletes",
	"pl": "players",
	"psu": "player_session_updates",
	"psx": "player_session_deletes",
	"plc": "player_lifecycle",
	"plu": "player_lifecycle_updates",
	"plx": "player_lifecycle_deletes",
	"ta": "total_asteroids",
	"bid": "batch_id",
	"ev": "events",
	"ei": "event_id",
	"i": "id",
	"pid": "player_id",
	"self": "self_id",
	"srct": "source_type",
	"src": "source_id",
	"stat": "status",
	"r": "rotation",
	"h": "health",
	"sco": "score",
	"lv": "lives",
	"rcd": "respawn_cooldown",
	"st": "ship_type",
	"sh": "shields",
	"th": "thrusting",
	"tk": "target_kind",
	"tid": "target_id",
	"tt": "target_type",
	"dt": "damage_type",
	"dc": "damage_cause",
	"ba": "base_amount",
	"ma": "modified_amount",
	"ah": "applied_to_health",
	"abs": "absorbed_by_shield",
	"rh": "remaining_health",
	"rs": "remaining_shield",
	"oi": "owner_id",
	"wid": "weapon_id",
	"pt": "projectile_type",
	"sz": "size",
	"sl": "scale",
	"v": "variant",
	"pkid": "pickup_id",
	"pkt": "pickup_type",
	"pcl": "pickup_class",
	"age": "age_seconds",
	"life": "lifespan_seconds",
	"tbl": "table_id",
	"lva": "lives_after",
	"fx": "effect_type",
	"amt": "amount",
	"rd": "respawn_delay",
	"pwid": "primary_weapon_id",
	"pap": "primary_ammo_policy",
	"pcr": "primary_cooldown_remaining",
	"par": "primary_ammo_remaining",
	"swid": "secondary_weapon_id",
	"sap": "secondary_ammo_policy",
	"scr": "secondary_cooldown_remaining",
	"sar": "secondary_ammo_remaining",
	"spx": "spawn_x",
	"spy": "spawn_y",
}

const _VALUE_MAPS := {
	"type": {
		"wf": "world_full",
		"wd": "world_delta",
		"ad": "asteroid_delta",
		"bd": "bullet_delta",
		"of": "overlay_full",
		"od": "overlay_delta",
		"sf": "session_full",
		"sd": "session_delta",
		"eb": "event_batch",
		"bb": "bullet_blast",
		"shd": "ship_death",
		"rfx": "radial_effect_started",
		"pcol": "pickup_collected",
		"pea": "pickup_effect_applied",
		"pexp": "pickup_expired",
		"pdr": "pickup_dropped",
		"dmg": "damage_applied",
		"dots": "damage_over_time_started",
		"dott": "damage_over_time_tick",
	},
	"lane": {
		"w": "world",
		"o": "overlay",
		"s": "session",
	},
	"snapshot_kind": {
		"f": "full",
		"d": "delta",
	},
}

static func expand_packet(packet: Dictionary) -> Dictionary:
	var packet_type = packet.get("type", packet.get("t"))
	var expanded: Dictionary = _expand_value(packet, null, packet_type)
	_normalize_runtime_metadata(expanded)
	return expanded


static func expand_compact_asteroid_id(value):
	if value is int:
		return "asteroid-%d" % value
	if value is float:
		if is_equal_approx(value, floor(value)):
			return "asteroid-%d" % int(value)
		return value
	if value is String:
		if value.begins_with("asteroid-"):
			return value
		if value.is_valid_int():
			return "asteroid-%s" % value
	return value


static func expand_compact_pickup_id(value):
	return expand_compact_prefixed_id(value, "pickup-")


static func expand_compact_table_id(value):
	return expand_compact_prefixed_id(value, "table-")


static func expand_compact_ship_id(value):
	return expand_compact_prefixed_id(value, "ship-")


static func expand_compact_prefixed_id(value, prefix):
	if value is int:
		return "%s%d" % [prefix, value]
	if value is float:
		if is_equal_approx(value, floor(value)):
			return "%s%d" % [prefix, int(value)]
		return value
	if value is String:
		if value.begins_with(prefix):
			return value
		if value.is_valid_int():
			return "%s%s" % [prefix, value]
	return value


static func expand_compact_bullet_id(value):
	return expand_compact_prefixed_id(value, "bullet-")


static func expand_compact_presentation_event_id(value):
	if value is Array and value.size() == 2 and value[0] == "pe":
		return "presentation-event-%s" % str(value[1])
	return expand_compact_prefixed_id(value, "presentation-event-")


static func expand_compact_event_batch_id(value):
	if value is Array and value.size() == 2 and value[0] == "eb":
		return "event-batch-%s" % str(value[1])
	return expand_compact_prefixed_id(value, "event-batch-")


static func expand_compact_player_id(value):
	return expand_compact_prefixed_id(value, "player-")


static func expand_compact_tagged_id(value):
	if not (value is Array and value.size() == 2):
		return value
	var tag = value[0]
	var suffix = value[1]
	match tag:
		"p":
			return expand_compact_player_id(suffix)
		"b":
			return expand_compact_bullet_id(suffix)
		"a":
			return expand_compact_asteroid_id(suffix)
		"pk":
			return expand_compact_pickup_id(suffix)
		"s":
			return expand_compact_ship_id(suffix)
		"tbl":
			return expand_compact_table_id(suffix)
		"pe":
			return expand_compact_presentation_event_id(suffix)
		"eb":
			return expand_compact_event_batch_id(suffix)
		_:
			return value


static func _expand_source_or_tagged_id(source_type, value):
	match source_type:
		"player":
			return expand_compact_player_id(value)
		"ship":
			return expand_compact_ship_id(value)
		"projectile", "bullet":
			return expand_compact_bullet_id(value)
		"asteroid":
			return expand_compact_asteroid_id(value)
		"pickup":
			return expand_compact_pickup_id(value)
		_:
			return expand_compact_tagged_id(value)


static func _expand_target_or_tagged_id(target_kind, value):
	match target_kind:
		"player":
			return expand_compact_player_id(value)
		"ship":
			return expand_compact_ship_id(value)
		"projectile", "bullet":
			return expand_compact_bullet_id(value)
		"asteroid":
			return expand_compact_asteroid_id(value)
		"pickup":
			return expand_compact_pickup_id(value)
		_:
			return expand_compact_tagged_id(value)


static func _expand_session_player_record(value, packet_type = null):
	if value is Array and value.size() == 11:
		return {
			"id": expand_compact_player_id(value[0]),
			"ship_type": value[1],
			"score": value[2],
			"lives": value[3],
			"respawn_cooldown": value[4],
			"primary_weapon_id": value[5],
			"primary_ammo_policy": value[6],
			"secondary_weapon_id": value[7],
			"secondary_ammo_policy": value[8],
			"spawn_x": value[9],
			"spawn_y": value[10],
		}
	return _expand_value(value, "players", null)


static func _expand_session_player_update_record(value, packet_type = null):
	if value is Array and value.size() >= 1:
		var expanded := {
			"id": expand_compact_player_id(value[0]),
		}
		var index := 1
		while index + 1 < value.size():
			var key = str(value[index])
			var expanded_key = _KEY_MAP.get(key, key)
			expanded[expanded_key] = value[index + 1]
			index += 2
		return expanded
	return _expand_value(value, "player_session_updates", null)


static func _expand_player_lifecycle_record(value, packet_type = null):
	if value is Array and value.size() >= 2:
		return {
			"player_id": expand_compact_player_id(value[0]),
			"status": value[1],
		}
	return _expand_value(value, "player_lifecycle", null)


static func _expand_player_lifecycle_delete(value, packet_type = null):
	return expand_compact_player_id(value)


static func _expand_ship_record(value, packet_type = null):
	if value is Array and value.size() == 10:
		return {
			"id": expand_compact_player_id(value[0]),
			"ship_type": value[1],
			"x": value[2],
			"y": value[3],
			"rotation": value[4],
			"health": value[5],
			"shields": value[6],
			"thrusting": value[7],
			"target_kind": value[8],
			"target_id": _expand_target_or_tagged_id(value[8], value[9]),
		}
	return _expand_value(value, "ships", null)


static func _expand_ship_update_record(value, packet_type = null):
	if value is Array and value.size() >= 1:
		var expanded := {
			"id": expand_compact_player_id(value[0]),
		}
		if value.size() > 1 and value[1] != null:
			expanded["x"] = value[1]
		if value.size() > 2 and value[2] != null:
			expanded["y"] = value[2]
		if value.size() > 3 and value[3] != null:
			expanded["rotation"] = value[3]
		if value.size() > 4 and value[4] != null:
			expanded["thrusting"] = value[4]
		return expanded
	return _expand_value(value, "ship_updates", null)


static func _expand_bullet_record(value, packet_type = null):
	if value is Array and value.size() == 7:
		return {
			"id": expand_compact_bullet_id(value[0]),
			"owner_id": expand_compact_player_id(value[1]),
			"x": value[2],
			"y": value[3],
			"rotation": value[4],
			"weapon_id": value[5],
			"projectile_type": value[6],
		}
	return _expand_value(value, "bullets", null)


static func _expand_bullet_update_record(value, packet_type = null):
	if value is Array and value.size() >= 1:
		var expanded := {
			"id": expand_compact_bullet_id(value[0]),
		}
		if value.size() > 1 and value[1] != null:
			expanded["x"] = value[1]
		if value.size() > 2 and value[2] != null:
			expanded["y"] = value[2]
		if value.size() > 3 and value[3] != null:
			expanded["rotation"] = value[3]
		return expanded
	return _expand_value(value, "bullet_updates", null)


static func _expand_event_record(value):
	if value is Array and value.size() >= 2:
		var expanded := {
			"event_id": expand_compact_presentation_event_id(value[1]),
		}
		match value[0]:
			"bb":
				expanded["type"] = "bullet_blast"
				expanded["x"] = value[2]
				expanded["y"] = value[3]
			"shd":
				expanded["type"] = "ship_death"
				expanded["player_id"] = expand_compact_player_id(value[2])
				expanded["lives"] = value[3]
				expanded["respawn_delay"] = value[4]
				expanded["x"] = value[5]
				expanded["y"] = value[6]
			"dmg":
				expanded["type"] = "damage_applied"
				expanded["source_type"] = value[2]
				expanded["source_id"] = _expand_source_or_tagged_id(value[2], value[3])
				expanded["effect_type"] = value[4]
				expanded["amount"] = value[5]
				expanded["x"] = value[6]
				expanded["y"] = value[7]
			"dots":
				expanded["type"] = "damage_over_time_started"
				expanded["source_type"] = value[2]
				expanded["source_id"] = _expand_source_or_tagged_id(value[2], value[3])
				expanded["effect_type"] = value[4]
				expanded["amount"] = value[5]
			"dott":
				expanded["type"] = "damage_over_time_tick"
				expanded["source_type"] = value[2]
				expanded["source_id"] = _expand_source_or_tagged_id(value[2], value[3])
				expanded["effect_type"] = value[4]
				expanded["amount"] = value[5]
				expanded["x"] = value[6]
				expanded["y"] = value[7]
			"rfx":
				expanded["type"] = "radial_effect_started"
				expanded["source_type"] = value[2]
				expanded["source_id"] = _expand_source_or_tagged_id(value[2], value[3])
				expanded["effect_type"] = value[4]
				expanded["x"] = value[5]
				expanded["y"] = value[6]
			"pcol":
				expanded["type"] = "pickup_collected"
				expanded["player_id"] = expand_compact_player_id(value[2])
				expanded["pickup_id"] = expand_compact_pickup_id(value[3])
				expanded["pickup_type"] = value[4]
				expanded["x"] = value[5]
				expanded["y"] = value[6]
			"pea":
				expanded["type"] = "pickup_effect_applied"
				expanded["player_id"] = expand_compact_player_id(value[2])
				expanded["pickup_id"] = expand_compact_pickup_id(value[3])
				expanded["pickup_type"] = value[4]
				expanded["effect_type"] = value[5]
				expanded["amount"] = value[6]
				expanded["lives_after"] = value[7]
			"pexp":
				expanded["type"] = "pickup_expired"
				expanded["pickup_id"] = expand_compact_pickup_id(value[2])
				expanded["pickup_type"] = value[3]
				expanded["x"] = value[4]
				expanded["y"] = value[5]
			"pdr":
				expanded["type"] = "pickup_dropped"
				expanded["pickup_id"] = expand_compact_pickup_id(value[2])
				expanded["pickup_type"] = value[3]
				expanded["source_type"] = value[4]
				expanded["source_id"] = _expand_source_or_tagged_id(value[4], value[5])
				expanded["table_id"] = expand_compact_table_id(value[6])
				expanded["x"] = value[7]
				expanded["y"] = value[8]
			_:
				return _expand_value(value, "events", null)
		return expanded
	return _expand_value(value, "events", null)


static func _expand_asteroid_record(value, packet_type = null):
	if value is Array and value.size() == 7:
		return {
			"id": expand_compact_asteroid_id(value[0]),
			"x": value[1],
			"y": value[2],
			"size": value[3],
			"health": value[4],
			"scale": value[5],
			"variant": value[6],
		}
	return _expand_value(value, "asteroids", null)


static func _expand_asteroid_update_record(value, packet_type = null):
	if value is Array and value.size() >= 1:
		var expanded := {
			"id": expand_compact_asteroid_id(value[0]),
		}
		if value.size() > 1 and value[1] != null:
			expanded["x"] = value[1]
		if value.size() > 2 and value[2] != null:
			expanded["y"] = value[2]
		return expanded
	return _expand_value(value, "asteroid_updates", null)


static func _is_compact_asteroid_record_list(parent_key) -> bool:
	return parent_key == "asteroids" or parent_key == "asteroid_creates"


static func _expand_value(value, parent_key, packet_type):
	if value is Dictionary:
		var expanded := {}
		for raw_key in value.keys():
			var key = str(raw_key)
			var expanded_key = _KEY_MAP.get(key, key)
			var expanded_value = _expand_value(value[raw_key], expanded_key, packet_type)
			if key == "bid" or expanded_key == "batch_id":
				expanded_value = expand_compact_event_batch_id(expanded_value)
			expanded[expanded_key] = expanded_value
		return expanded
	if value is Array:
		if parent_key == "players":
			if packet_type == "session_full" or packet_type == "session_delta" or packet_type == "sf" or packet_type == "sd" or value.size() == 11:
				var expanded_players := []
				expanded_players.resize(value.size())
				for index in range(value.size()):
					expanded_players[index] = _expand_session_player_record(value[index], packet_type)
				return expanded_players
		if parent_key == "player_session_updates":
			var expanded_player_session_updates := []
			expanded_player_session_updates.resize(value.size())
			for index in range(value.size()):
				expanded_player_session_updates[index] = _expand_session_player_update_record(value[index], packet_type)
			return expanded_player_session_updates
		if parent_key == "player_session_deletes":
			var expanded_player_session_deletes := []
			expanded_player_session_deletes.resize(value.size())
			for index in range(value.size()):
				expanded_player_session_deletes[index] = expand_compact_player_id(value[index])
			return expanded_player_session_deletes
		if parent_key == "player_lifecycle" or parent_key == "player_lifecycle_updates":
			var expanded_player_lifecycle := []
			expanded_player_lifecycle.resize(value.size())
			for index in range(value.size()):
				expanded_player_lifecycle[index] = _expand_player_lifecycle_record(value[index], packet_type)
			return expanded_player_lifecycle
		if parent_key == "player_lifecycle_deletes":
			var expanded_player_lifecycle_deletes := []
			expanded_player_lifecycle_deletes.resize(value.size())
			for index in range(value.size()):
				expanded_player_lifecycle_deletes[index] = _expand_player_lifecycle_delete(value[index], packet_type)
			return expanded_player_lifecycle_deletes
		if parent_key == "events":
			var expanded_events := []
			expanded_events.resize(value.size())
			for index in range(value.size()):
				expanded_events[index] = _expand_event_record(value[index])
			return expanded_events
		if parent_key == "ship_updates":
			var expanded_ship_updates := []
			expanded_ship_updates.resize(value.size())
			for index in range(value.size()):
				expanded_ship_updates[index] = _expand_ship_update_record(value[index], packet_type)
			return expanded_ship_updates
		if parent_key == "ship_deletes":
			var expanded_ship_deletes := []
			expanded_ship_deletes.resize(value.size())
			for index in range(value.size()):
				expanded_ship_deletes[index] = expand_compact_player_id(value[index])
			return expanded_ship_deletes
		if parent_key == "ships" or parent_key == "ship_creates":
			var expanded_ships := []
			expanded_ships.resize(value.size())
			for index in range(value.size()):
				expanded_ships[index] = _expand_ship_record(value[index])
			return expanded_ships
		if parent_key == "bullet_updates":
			var expanded_bullet_updates := []
			expanded_bullet_updates.resize(value.size())
			for index in range(value.size()):
				expanded_bullet_updates[index] = _expand_bullet_update_record(value[index], packet_type)
			return expanded_bullet_updates
		if parent_key == "bullet_deletes":
			var expanded_bullet_deletes := []
			expanded_bullet_deletes.resize(value.size())
			for index in range(value.size()):
				expanded_bullet_deletes[index] = expand_compact_bullet_id(value[index])
			return expanded_bullet_deletes
		if parent_key == "bullets" or parent_key == "bullet_creates":
			var expanded_bullets := []
			expanded_bullets.resize(value.size())
			for index in range(value.size()):
				expanded_bullets[index] = _expand_bullet_record(value[index])
			return expanded_bullets
		if parent_key == "asteroid_updates":
			var expanded_asteroid_updates := []
			expanded_asteroid_updates.resize(value.size())
			for index in range(value.size()):
				expanded_asteroid_updates[index] = _expand_asteroid_update_record(value[index], packet_type)
			return expanded_asteroid_updates
		if parent_key == "asteroid_deletes":
			var expanded_asteroid_deletes := []
			expanded_asteroid_deletes.resize(value.size())
			for index in range(value.size()):
				expanded_asteroid_deletes[index] = expand_compact_asteroid_id(value[index])
			return expanded_asteroid_deletes
		if _is_compact_asteroid_record_list(parent_key):
			var expanded_asteroids := []
			expanded_asteroids.resize(value.size())
			for index in range(value.size()):
				expanded_asteroids[index] = _expand_asteroid_record(value[index], packet_type)
			return expanded_asteroids
		var expanded_array := []
		expanded_array.resize(value.size())
		for index in range(value.size()):
			expanded_array[index] = _expand_value(value[index], parent_key, packet_type)
		return expanded_array
	if parent_key != null:
		var value_map = _VALUE_MAPS.get(parent_key)
		if value_map != null:
			var string_value = str(value)
			if parent_key == "batch_id":
				return expand_compact_event_batch_id(value)
			if value_map.has(string_value):
				return value_map[string_value]
	return value


static func _normalize_runtime_metadata(packet: Dictionary) -> void:
	var packet_type = packet.get("type")
	if not _is_runtime_packet_type(packet_type):
		return

	var lane = packet.get("lane")
	if lane == null:
		lane = LaneMetadata.PACKET_TYPE_TO_LANE.get(packet_type)
		if lane != null:
			packet["lane"] = lane

	var snapshot_kind = packet.get("snapshot_kind")
	if snapshot_kind == null:
		snapshot_kind = _snapshot_kind_from_type(packet_type)
		if snapshot_kind != "":
			packet["snapshot_kind"] = snapshot_kind

	var sequence = packet.get("sequence")
	if lane != null and sequence != null:
		if packet.get("snapshot_id") == null:
			var snapshot_id = _default_snapshot_id(lane, snapshot_kind, sequence)
			if snapshot_id != "":
				packet["snapshot_id"] = snapshot_id
		if packet.get("baseline_id") == null:
			var baseline_id = _default_baseline_id(lane, snapshot_kind, sequence, packet.get("baseline_sequence"))
			if baseline_id != "":
				packet["baseline_id"] = baseline_id

	if packet.get("chunk_index") == null:
		packet["chunk_index"] = 0
	if packet.get("chunk_count") == null:
		packet["chunk_count"] = 1
	if packet.get("is_final_chunk") == null:
		var chunk_index: int = int(packet.get("chunk_index", 0))
		var chunk_count: int = int(packet.get("chunk_count", 1))
		packet["is_final_chunk"] = chunk_count <= 1 or chunk_index == chunk_count - 1


static func _is_runtime_packet_type(packet_type) -> bool:
	return packet_type in LaneMetadata.PACKET_FAMILY_WORLD \
		or packet_type in LaneMetadata.PACKET_FAMILY_OVERLAY \
		or packet_type in LaneMetadata.PACKET_FAMILY_SESSION \
		or packet_type == "asteroid_delta" \
		or packet_type == "bullet_delta"


static func _snapshot_kind_from_type(packet_type) -> String:
	var type_string := str(packet_type)
	if type_string.ends_with("_full"):
		return "full"
	if type_string.ends_with("_delta"):
		return "delta"
	return ""


static func _default_snapshot_id(lane, snapshot_kind, sequence) -> String:
	if snapshot_kind == "full":
		return "%s-baseline-%s" % [lane, str(sequence)]
	if snapshot_kind == "delta":
		return "%s-snapshot-%s" % [lane, str(sequence)]
	return ""


static func _default_baseline_id(lane, snapshot_kind, sequence, baseline_sequence) -> String:
	if snapshot_kind == "full":
		return "%s-baseline-%s" % [lane, str(sequence)]
	if snapshot_kind == "delta" and baseline_sequence != null:
		return "%s-baseline-%s" % [lane, str(baseline_sequence)]
	return ""

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
	var expanded: Dictionary = _expand_value(packet, null)
	_normalize_runtime_metadata(expanded)
	return expanded

static func _expand_value(value, parent_key):
	if value is Dictionary:
		var expanded := {}
		for raw_key in value.keys():
			var key = str(raw_key)
			var expanded_key = _KEY_MAP.get(key, key)
			expanded[expanded_key] = _expand_value(value[raw_key], expanded_key)
		return expanded
	if value is Array:
		var expanded_array := []
		expanded_array.resize(value.size())
		for index in range(value.size()):
			expanded_array[index] = _expand_value(value[index], parent_key)
		return expanded_array
	if parent_key != null:
		var value_map = _VALUE_MAPS.get(parent_key)
		if value_map != null:
			var string_value = str(value)
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
		or packet_type in LaneMetadata.PACKET_FAMILY_SESSION


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

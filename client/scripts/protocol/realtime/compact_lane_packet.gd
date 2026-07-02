extends RefCounted

const _KEY_MAP := {
	"t": "type",
	"l": "lane",
	"q": "sequence",
	"b": "baseline_id",
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
	"i": "id",
	"pid": "player_id",
	"self": "self_id",
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
	"oi": "owner_id",
	"wid": "weapon_id",
	"pt": "projectile_type",
	"sz": "size",
	"sl": "scale",
	"v": "variant",
	"pcl": "pickup_class",
	"age": "age_seconds",
	"life": "lifespan_seconds",
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
	return _expand_value(packet, null)

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

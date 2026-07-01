extends RefCounted

const OverlayLaneState = preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const SessionLaneState = preload("res://scripts/protocol/realtime/session_lane_state.gd")

const POLICY_FLOAT_GENERIC = "float_generic"
const POLICY_RATIO_0_1 = "ratio_0_1"
const POLICY_PERCENT_0_100 = "percent_0_100"
const POLICY_SECONDS = "seconds"
const POLICY_SIGNED_SECONDS = "signed_seconds"
const POLICY_ANGLE_TURN = "angle_turn"
const POLICY_POSITION = "position"
const POLICY_VELOCITY = "velocity"
const POLICY_ANGULAR_VELOCITY = "angular_velocity"

const MODE_REGULAR_SCALE = "regular_scale"
const MODE_RATIO = "ratio"
const MODE_ANGLE_TURN = "angle_turn"

const _POLICIES = {
	POLICY_FLOAT_GENERIC: {"scale": 1000.0, "mode": MODE_REGULAR_SCALE},
	POLICY_RATIO_0_1: {"scale": 65535.0, "mode": MODE_RATIO},
	POLICY_PERCENT_0_100: {"scale": 100.0, "mode": MODE_REGULAR_SCALE},
	POLICY_SECONDS: {"scale": 1000.0, "mode": MODE_REGULAR_SCALE},
	POLICY_SIGNED_SECONDS: {"scale": 1000.0, "mode": MODE_REGULAR_SCALE},
	POLICY_ANGLE_TURN: {"scale": 65535.0, "mode": MODE_ANGLE_TURN},
	POLICY_POSITION: {"scale": 10.0, "mode": MODE_REGULAR_SCALE},
	POLICY_VELOCITY: {"scale": 10.0, "mode": MODE_REGULAR_SCALE},
	POLICY_ANGULAR_VELOCITY: {"scale": 1000.0, "mode": MODE_REGULAR_SCALE},
}

static func decode(policy_name: String, encoded):
	var policy = _POLICIES.get(policy_name, _POLICIES[POLICY_FLOAT_GENERIC])
	match policy_name:
		POLICY_RATIO_0_1:
			return float(encoded) / 65535.0
		POLICY_ANGLE_TURN:
			return float(encoded) / 65535.0
		_:
			return float(encoded) / policy.scale


static func decode_session_value(value):
	return _decode_value(value, "session")

static func decode_overlay_value(value):
	return _decode_value(value, "overlay")

static func _decode_value(value, lane: String, field_path: String = ""):
	if value is Dictionary:
		var decoded: Dictionary = {}
		for key in value.keys():
			var child_path := _join_field_path(field_path, str(key))
			decoded[key] = _decode_value(value.get(key), lane, child_path)
		return decoded
	if value is Array:
		var decoded_array := []
		for index in range(value.size()):
			var child_path := _index_field_path(field_path, index)
			decoded_array.append(_decode_value(value[index], lane, child_path))
		return decoded_array
	if value is float:
		return decode(_lookup_policy_name(lane, field_path), value)
	return value



static func decode_overlay_state(overlay_lane_state):
	if overlay_lane_state == null:
		return null
	var decoded = OverlayLaneState.new()
	decoded.self_id = overlay_lane_state.self_id
	decoded.lives = overlay_lane_state.lives
	decoded.score = overlay_lane_state.score
	decoded.respawn_cooldown = decode(POLICY_SECONDS, overlay_lane_state.respawn_cooldown)
	decoded.primary_weapon_id = overlay_lane_state.primary_weapon_id
	decoded.secondary_weapon_id = overlay_lane_state.secondary_weapon_id
	decoded.primary_ammo_policy = overlay_lane_state.primary_ammo_policy
	decoded.secondary_ammo_policy = overlay_lane_state.secondary_ammo_policy
	decoded.primary_cooldown_remaining = decode(POLICY_SECONDS, overlay_lane_state.primary_cooldown_remaining)
	decoded.secondary_cooldown_remaining = decode(POLICY_SECONDS, overlay_lane_state.secondary_cooldown_remaining)
	decoded.primary_ammo_remaining = overlay_lane_state.primary_ammo_remaining
	decoded.secondary_ammo_remaining = overlay_lane_state.secondary_ammo_remaining
	return decoded

static func decode_session_state(session_lane_state):
	if session_lane_state == null:
		return null
	var decoded = SessionLaneState.new()
	decoded.player_sessions = _decode_session_sessions(session_lane_state.player_sessions)
	decoded.player_lifecycle = _decode_session_lifecycle(session_lane_state.player_lifecycle)
	decoded.total_asteroids = session_lane_state.total_asteroids
	return decoded

static func _decode_session_sessions(value):
	var decoded := {}
	if value is Dictionary:
		for key in value.keys():
			decoded[key] = _decode_session_player(value.get(key))
	return decoded

static func _decode_session_player(player):
	if not (player is Dictionary):
		return player
	var decoded: Dictionary = player.duplicate(true)
	decoded["respawn_cooldown"] = decode(POLICY_SECONDS, player.get("respawn_cooldown", 0.0))
	decoded["spawn_x"] = decode(POLICY_POSITION, player.get("spawn_x", 0.0))
	decoded["spawn_y"] = decode(POLICY_POSITION, player.get("spawn_y", 0.0))
	return decoded

static func _decode_session_lifecycle(value):
	return value.duplicate(true) if value is Dictionary else {}

static func decode_world_ship_record(record: Dictionary) -> Dictionary:
	var decoded: Dictionary = record.duplicate(true)
	_decode_field_if_present(decoded, "x", POLICY_POSITION)
	_decode_field_if_present(decoded, "y", POLICY_POSITION)
	_decode_field_if_present(decoded, "rotation", POLICY_FLOAT_GENERIC)
	return decoded

static func decode_world_bullet_record(record: Dictionary) -> Dictionary:
	var decoded: Dictionary = record.duplicate(true)
	_decode_field_if_present(decoded, "x", POLICY_POSITION)
	_decode_field_if_present(decoded, "y", POLICY_POSITION)
	_decode_field_if_present(decoded, "rotation", POLICY_FLOAT_GENERIC)
	return decoded

static func decode_world_asteroid_record(record: Dictionary) -> Dictionary:
	var decoded: Dictionary = record.duplicate(true)
	_decode_field_if_present(decoded, "x", POLICY_POSITION)
	_decode_field_if_present(decoded, "y", POLICY_POSITION)
	_decode_field_if_present(decoded, "scale", POLICY_FLOAT_GENERIC)
	return decoded

static func decode_world_pickup_record(record: Dictionary) -> Dictionary:
	var decoded: Dictionary = record.duplicate(true)
	_decode_field_if_present(decoded, "x", POLICY_POSITION)
	_decode_field_if_present(decoded, "y", POLICY_POSITION)
	_decode_field_if_present(decoded, "age_seconds", POLICY_SECONDS)
	_decode_field_if_present(decoded, "lifespan_seconds", POLICY_SECONDS)
	return decoded

static func _decode_field_if_present(decoded: Dictionary, field: String, policy_name: String) -> void:
	if not decoded.has(field):
		return
	decoded[field] = decode(policy_name, decoded[field])

static func _lookup_policy_name(lane: String, field_path: String) -> String:
	match lane:
		"session":
			match field_path:
				"players.respawn_cooldown":
					return POLICY_SECONDS
				"players.spawn_x":
					return POLICY_POSITION
				"players.spawn_y":
					return POLICY_POSITION
				"elapsed":
					return POLICY_SECONDS
				"duration":
					return POLICY_SECONDS
				"timer":
					return POLICY_SECONDS
				"lifetime":
					return POLICY_SECONDS
			return POLICY_FLOAT_GENERIC
		"overlay":
			match field_path:
				"respawn_cooldown":
					return POLICY_SECONDS
				"primary_cooldown_remaining":
					return POLICY_SECONDS
				"secondary_cooldown_remaining":
					return POLICY_SECONDS
			return POLICY_FLOAT_GENERIC
	return POLICY_FLOAT_GENERIC

static func _join_field_path(parent: String, child: String) -> String:
	if parent == "":
		return child
	return "%s.%s" % [parent, child]

static func _index_field_path(parent: String, index: int) -> String:
	if parent == "":
		return "[%d]" % index
	return "%s[%d]" % [parent, index]


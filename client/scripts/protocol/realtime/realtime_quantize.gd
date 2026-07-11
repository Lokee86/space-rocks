extends RefCounted

const OverlayLaneState = preload("res://scripts/protocol/realtime/overlay_lane_state.gd")
const SessionLaneState = preload("res://scripts/protocol/realtime/session_lane_state.gd")
const RealtimeWireGenerated = preload("res://scripts/generated/networking/realtime_wire_generated.gd")

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
	if encoded == null:
		return null
	var policy = _POLICIES.get(policy_name, _POLICIES[POLICY_FLOAT_GENERIC])
	match policy_name:
		POLICY_RATIO_0_1, POLICY_ANGLE_TURN:
			return float(encoded) / 65535.0
		_:
			return float(encoded) / policy.scale

static func decode_session_value(value):
	return _decode_value(value, "session")

static func decode_overlay_value(value):
	return _decode_value(value, "overlay")

static func _decode_value(value, path: String):
	if value is Dictionary:
		var decoded: Dictionary = {}
		for key in value.keys():
			var child_path := _join_field_path(path, str(key))
			if value.get(key) is Dictionary and not _has_policy_prefix(child_path):
				child_path = path
			decoded[key] = _decode_value(value.get(key), child_path)
		return decoded
	if value is Array:
		var decoded_array: Array = []
		for item in value:
			decoded_array.append(_decode_value(item, path))
		return decoded_array
	if value is int or value is float:
		if RealtimeWireGenerated.QUANTIZATION_POLICY_BY_PATH.has(path):
			return decode(_lookup_policy_name(path), value)
		return value
	return value

static func _lookup_policy_name(path: String) -> String:
	return RealtimeWireGenerated.QUANTIZATION_POLICY_BY_PATH.get(path, POLICY_FLOAT_GENERIC)

static func _has_policy_prefix(path: String) -> bool:
	for policy_path in RealtimeWireGenerated.QUANTIZATION_POLICY_BY_PATH:
		if str(policy_path).begins_with(path + "."):
			return true
	return false

static func decode_overlay_state(overlay_lane_state):
	if overlay_lane_state == null:
		return null
	var decoded = OverlayLaneState.new()
	decoded.self_id = overlay_lane_state.self_id
	decoded.lives = overlay_lane_state.lives
	decoded.score = overlay_lane_state.score
	decoded.respawn_cooldown = _decode_field(overlay_lane_state.respawn_cooldown, "overlay.respawn_cooldown")
	decoded.primary_weapon_id = overlay_lane_state.primary_weapon_id
	decoded.secondary_weapon_id = overlay_lane_state.secondary_weapon_id
	decoded.primary_ammo_policy = overlay_lane_state.primary_ammo_policy
	decoded.secondary_ammo_policy = overlay_lane_state.secondary_ammo_policy
	decoded.primary_cooldown_remaining = _decode_field(overlay_lane_state.primary_cooldown_remaining, "overlay.primary_cooldown_remaining")
	decoded.secondary_cooldown_remaining = _decode_field(overlay_lane_state.secondary_cooldown_remaining, "overlay.secondary_cooldown_remaining")
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
	return _decode_value(player, "session.players")

static func _decode_session_lifecycle(value):
	return value.duplicate(true) if value is Dictionary else {}

static func _decode_record(record: Dictionary, prefix: String) -> Dictionary:
	return _decode_value(record, prefix)

static func decode_world_ship_record(record: Dictionary) -> Dictionary:
	return _decode_record(record, "world.ships")

static func decode_world_bullet_record(record: Dictionary) -> Dictionary:
	return _decode_record(record, "world.bullets")

static func decode_world_asteroid_record(record: Dictionary) -> Dictionary:
	return _decode_record(record, "world.asteroids")

static func decode_world_pickup_record(record: Dictionary) -> Dictionary:
	return _decode_record(record, "world.pickups")

static func decode_event_record(event: Dictionary) -> Dictionary:
	var event_type := str(event.get("type", ""))
	return _decode_record(event, "event." + event_type)

static func _decode_field(value, path: String):
	if value == null:
		return null
	return decode(_lookup_policy_name(path), value)

static func _join_field_path(parent: String, child: String) -> String:
	if parent == "":
		return child
	return "%s.%s" % [parent, child]

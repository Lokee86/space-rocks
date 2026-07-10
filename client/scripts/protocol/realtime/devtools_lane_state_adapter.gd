extends RefCounted

const RealtimeQuantize = preload("res://scripts/protocol/realtime/realtime_quantize.gd")

func build_state(presentation_state) -> Dictionary:
	var state := {
		"world": {
			"ships": {},
			"asteroids": {},
			"bullets": {},
			"pickups": {},
		},
		"session": {
			"players": {},
			"player_lifecycle": {},
		},
		"overlay": {
			"self_id": "",
		},
	}

	if presentation_state == null:
		return state

	var overlay_lane_state = _field_or_key(presentation_state, "overlay_lane_state")
	var overlay_self_id = _field_or_key(overlay_lane_state, "self_id")
	if overlay_self_id != null:
		state["overlay"]["self_id"] = str(overlay_self_id)

	var world_lane_state = _field_or_key(presentation_state, "world_lane_state")
	if world_lane_state != null:
		state["world"]["ships"] = _dictionary_field_or_key(world_lane_state, "ships")
		state["world"]["asteroids"] = _dictionary_field_or_key(world_lane_state, "asteroids")
		state["world"]["bullets"] = _dictionary_field_or_key(world_lane_state, "bullets")
		state["world"]["pickups"] = _dictionary_field_or_key(world_lane_state, "pickups")

	var session_lane_state = _field_or_key(presentation_state, "session_lane_state")
	if session_lane_state != null:
		if session_lane_state is Dictionary:
			state["session"]["players"] = _dictionary_field_or_key(session_lane_state, "player_sessions")
			state["session"]["player_lifecycle"] = _dictionary_field_or_key(session_lane_state, "player_lifecycle")
		else:
			var decoded_session_state = RealtimeQuantize.decode_session_state(session_lane_state)
			state["session"]["players"] = _dictionary_field_or_key(decoded_session_state, "player_sessions")
			state["session"]["player_lifecycle"] = _dictionary_field_or_key(decoded_session_state, "player_lifecycle")

	return state


func _duplicate_dictionary(value) -> Dictionary:
	if value is Dictionary:
		return value.duplicate(true)
	return {}


func _field_or_key(value, name: String, default_value = null):
	if value == null:
		return default_value
	if value is Dictionary:
		return value.get(name, default_value)
	if value is Object:
		var result = value.get(name)
		return result if result != null else default_value
	return default_value


func _dictionary_field_or_key(value, name: String) -> Dictionary:
	return _duplicate_dictionary(_field_or_key(value, name, {}))

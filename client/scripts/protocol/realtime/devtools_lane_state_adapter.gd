extends RefCounted

func build_state(router) -> Dictionary:
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

	if router == null:
		return state

	if router.overlay_lane_state != null and router.overlay_lane_state.self_id != null:
		state["overlay"]["self_id"] = str(router.overlay_lane_state.self_id)

	if router.world_lane_state != null:
		state["world"]["ships"] = _duplicate_dictionary(router.world_lane_state.ships)
		state["world"]["asteroids"] = _duplicate_dictionary(router.world_lane_state.asteroids)
		state["world"]["bullets"] = _duplicate_dictionary(router.world_lane_state.bullets)
		state["world"]["pickups"] = _duplicate_dictionary(router.world_lane_state.pickups)

	if router.session_lane_state != null:
		state["session"]["players"] = _duplicate_dictionary(router.session_lane_state.player_sessions)
		state["session"]["player_lifecycle"] = _duplicate_dictionary(router.session_lane_state.player_lifecycle)

	return state


func _duplicate_dictionary(value) -> Dictionary:
	if value is Dictionary:
		return value.duplicate(true)
	return {}

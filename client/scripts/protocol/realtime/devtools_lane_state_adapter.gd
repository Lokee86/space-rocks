extends RefCounted

func build_state(router) -> Dictionary:
	var state := {
		"self_id": "",
		"server_players": {},
		"player_sessions": {},
		"server_asteroids": {},
		"server_bullets": {},
		"server_pickups": {},
		"player_lifecycle": {},
	}

	if router == null:
		return state

	if router.overlay_lane_state != null and router.overlay_lane_state.self_id != null:
		state["self_id"] = str(router.overlay_lane_state.self_id)

	if router.world_lane_state != null:
		state["server_players"] = _duplicate_dictionary(router.world_lane_state.ships)
		state["server_asteroids"] = _duplicate_dictionary(router.world_lane_state.asteroids)
		state["server_bullets"] = _duplicate_dictionary(router.world_lane_state.bullets)
		state["server_pickups"] = _duplicate_dictionary(router.world_lane_state.pickups)

	if router.session_lane_state != null:
		state["player_sessions"] = _duplicate_dictionary(router.session_lane_state.player_sessions)
		state["player_lifecycle"] = _duplicate_dictionary(router.session_lane_state.player_lifecycle)

	return state


func _duplicate_dictionary(value) -> Dictionary:
	if value is Dictionary:
		return value.duplicate(true)
	return {}

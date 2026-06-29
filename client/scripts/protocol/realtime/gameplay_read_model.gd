extends RefCounted

func build_gameplay_read_model(world_lane_state, overlay_lane_state, session_lane_state, event_batch_applier = null) -> Dictionary:
	var read_model := {}

	read_model["self_id"] = overlay_lane_state.self_id
	read_model["lives"] = _pick_first_non_null(overlay_lane_state.lives, session_lane_state.player_sessions.get(overlay_lane_state.self_id, {}).get("lives"))
	read_model["players"] = _build_players(world_lane_state, overlay_lane_state, session_lane_state)
	read_model["player_sessions"] = session_lane_state.player_sessions.duplicate(true)
	read_model["player_lifecycle"] = session_lane_state.player_lifecycle.duplicate(true)
	read_model["bullets"] = world_lane_state.bullets.duplicate(true)
	read_model["asteroids"] = world_lane_state.asteroids.duplicate(true)
	read_model["pickups"] = world_lane_state.pickups.duplicate(true)
	read_model["total_asteroids"] = session_lane_state.total_asteroids
	if event_batch_applier != null:
		read_model["events"] = _collect_events(event_batch_applier)
	return read_model

func _build_players(world_lane_state, overlay_lane_state, session_lane_state) -> Array:
	var players := []
	for player_id in world_lane_state.ships.keys():
		var raw_ship = world_lane_state.ships[player_id]
		var ship: Dictionary = raw_ship as Dictionary
		var player := ship.duplicate(true)
		player["player_id"] = player_id
		player["self_id"] = overlay_lane_state.self_id
		if session_lane_state.player_sessions.has(player_id):
			player["player_session"] = session_lane_state.player_sessions[player_id].duplicate(true)
		if session_lane_state.player_lifecycle.has(player_id):
			player["player_lifecycle"] = session_lane_state.player_lifecycle[player_id].duplicate(true)
		players.append(player)
	return players

func _collect_events(event_batch_applier) -> Array:
	if event_batch_applier == null:
		return []
	if event_batch_applier.has_method("get_applied_events"):
		return event_batch_applier.get_applied_events()
	if event_batch_applier.has_method("events"):
		return event_batch_applier.events
	return []

func _pick_first_non_null(first_value, second_value):
	if first_value != null:
		return first_value
	return second_value



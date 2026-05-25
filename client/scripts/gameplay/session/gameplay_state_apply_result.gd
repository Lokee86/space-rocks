extends RefCounted

const Packets = preload("res://scripts/networking/packets/packets.gd")


static func from_packet_state(
	state: Dictionary,
	existing_has_received_state: bool,
	existing_has_initial_spawn: bool
) -> Dictionary:
	var self_id = state["self_id"]
	var server_players: Dictionary = state["server_players"]

	return {
		"self_id": self_id,
		"server_players": server_players,
		"server_bullets": state["server_bullets"],
		"server_asteroids": state["server_asteroids"],
		"server_events": state["server_events"],
		"player_lifecycle": state["player_lifecycle"],
		"has_lives": state["has_lives"],
		"lives": state["lives"],
		"has_received_state": true,
		"has_initial_spawn": existing_has_initial_spawn || server_players.has(self_id),
	}


static func apply_lives_to_hud(apply_result: Dictionary, hud_controller) -> void:
	if apply_result["has_lives"]:
		hud_controller.set_lives(apply_result["lives"])
	else:
		push_warning("State packet missing lives")


static func apply_world_sync(
	apply_result: Dictionary,
	world_sync,
	existing_has_received_state: bool
) -> void:
	world_sync.apply_state(
		apply_result["self_id"],
		apply_result["server_players"],
		apply_result["server_bullets"],
		apply_result["server_asteroids"],
		existing_has_received_state
	)


static func apply_score_to_hud(apply_result: Dictionary, hud_controller) -> void:
	var self_id = apply_result["self_id"]
	var server_players: Dictionary = apply_result["server_players"]
	if server_players.has(self_id):
		hud_controller.set_score(int(server_players[self_id].get(Packets.FIELD_SCORE, 0)))


static func apply_server_events(
	apply_result: Dictionary,
	gameplay_event_controller,
	self_death_callback: Callable
) -> void:
	var server_events: Array = apply_result["server_events"]
	if server_events.is_empty():
		return

	gameplay_event_controller.apply_server_events(
		server_events,
		apply_result["self_id"],
		self_death_callback
	)


static func confirm_alive_if_spawned(
	apply_result: Dictionary,
	hud_controller,
	gameplay_lifecycle_controller
) -> void:
	if !apply_result["server_players"].has(apply_result["self_id"]):
		return
	if hud_controller.is_dead && gameplay_lifecycle_controller.is_awaiting_respawn_confirmation():
		gameplay_lifecycle_controller.set_alive_state()

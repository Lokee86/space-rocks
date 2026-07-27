class_name MultiplayerLobbyPresenter
extends RefCounted

const MultiplayerLobbyScene := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const ObservabilityContract := preload("res://scripts/generated/observability/contract_generated.gd")

var multiplayer_lobby: MultiplayerLobby


func show_lobby(canvas_layer: CanvasLayer, state: LobbySessionState, callbacks: Dictionary) -> MultiplayerLobby:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		var lobby_instance: Node = MultiplayerLobbyScene.instantiate()
		multiplayer_lobby = lobby_instance as MultiplayerLobby
		if multiplayer_lobby == null:
			ClientLogger.emit_canonical(
			ObservabilityContract.EVENT_CLIENT_PRESENTATION_CONTRACT_VIOLATION,
			"Multiplayer lobby scene must instantiate its presentation root",
			{},
			{
				"subsystem": "lobby",
				"failure_mode": "wrong_scene_root",
				"resource_kind": "scene",
				"expected_type": "MultiplayerLobby",
				"actual_type": lobby_instance.get_class(),
				"resource_path": MultiplayerLobbyScene.resource_path,
			}
		)
			lobby_instance.queue_free()
			return null
		canvas_layer.add_child(multiplayer_lobby)
		_connect_lobby_signals(callbacks)

	multiplayer_lobby.apply_lobby_state(
		state.room_code,
		state.room_state,
		state.local_player_id,
		state.owner_id,
		state.max_players,
		state.members,
		state.team_structure,
		state.team_assignment_mode,
		state.team_count,
		state.team_assignments_locked,
		state.can_start_game()
	)
	multiplayer_lobby.set_start_enabled(state.can_start_game())
	multiplayer_lobby.show()
	return multiplayer_lobby


func clear_lobby() -> void:
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		multiplayer_lobby.queue_free()
	multiplayer_lobby = null


func current_lobby() -> MultiplayerLobby:
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		return multiplayer_lobby
	return null


func _connect_lobby_signals(callbacks: Dictionary) -> void:
	var ready_handler: Callable = callbacks.get("ready_requested", Callable())
	if !ready_handler.is_null() && !multiplayer_lobby.ready_requested.is_connected(ready_handler):
		multiplayer_lobby.ready_requested.connect(ready_handler)
	var start_handler: Callable = callbacks.get("start_game_requested", Callable())
	if !start_handler.is_null() && !multiplayer_lobby.start_game_requested.is_connected(start_handler):
		multiplayer_lobby.start_game_requested.connect(start_handler)
	var add_bot_handler: Callable = callbacks.get("add_bot_requested", Callable())
	if !add_bot_handler.is_null() && !multiplayer_lobby.add_bot_requested.is_connected(add_bot_handler):
		multiplayer_lobby.add_bot_requested.connect(add_bot_handler)
	var remove_handler: Callable = callbacks.get("remove_member_requested", Callable())
	if !remove_handler.is_null() && !multiplayer_lobby.remove_member_requested.is_connected(remove_handler):
		multiplayer_lobby.remove_member_requested.connect(remove_handler)
	var team_handler: Callable = callbacks.get("team_assignment_requested", Callable())
	if !team_handler.is_null() && !multiplayer_lobby.team_assignment_requested.is_connected(team_handler):
		multiplayer_lobby.team_assignment_requested.connect(team_handler)
	var leave_handler: Callable = callbacks.get("leave_requested", Callable())
	if !leave_handler.is_null() && !multiplayer_lobby.leave_requested.is_connected(leave_handler):
		multiplayer_lobby.leave_requested.connect(leave_handler)

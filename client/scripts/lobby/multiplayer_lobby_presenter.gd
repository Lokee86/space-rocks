class_name MultiplayerLobbyPresenter
extends RefCounted

const MultiplayerLobbyScene := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")

var multiplayer_lobby: MultiplayerLobby


func show_lobby(canvas_layer: CanvasLayer, state: LobbySessionState, callbacks: Dictionary) -> MultiplayerLobby:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		var lobby_instance: Node = MultiplayerLobbyScene.instantiate()
		multiplayer_lobby = lobby_instance as MultiplayerLobby
		if multiplayer_lobby == null:
			push_error("Multiplayer lobby scene must instantiate MultiplayerLobby; got %s" % lobby_instance.get_class())
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
	var leave_handler: Callable = callbacks.get("leave_requested", Callable())
	if !leave_handler.is_null() && !multiplayer_lobby.leave_requested.is_connected(leave_handler):
		multiplayer_lobby.leave_requested.connect(leave_handler)

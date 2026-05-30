extends RefCounted

const MultiplayerLobbyScene := preload("res://scenes/ui/dialogs/multiplayer_lobby.tscn")

var multiplayer_lobby: Control


func show_lobby(canvas_layer: CanvasLayer, state, callbacks: Dictionary) -> Control:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		multiplayer_lobby = MultiplayerLobbyScene.instantiate()
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
	if multiplayer_lobby.has_method("set_start_enabled"):
		multiplayer_lobby.set_start_enabled(state.can_start_game())
	multiplayer_lobby.show()
	return multiplayer_lobby


func clear_lobby() -> void:
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		multiplayer_lobby.queue_free()
	multiplayer_lobby = null


func current_lobby() -> Control:
	if multiplayer_lobby != null && is_instance_valid(multiplayer_lobby):
		return multiplayer_lobby
	return null


func _connect_lobby_signals(callbacks: Dictionary) -> void:
	_connect_lobby_signal("ready_requested", callbacks.get("ready_requested", Callable()))
	_connect_lobby_signal("start_game_requested", callbacks.get("start_game_requested", Callable()))
	_connect_lobby_signal("leave_requested", callbacks.get("leave_requested", Callable()))


func _connect_lobby_signal(signal_name: StringName, handler: Callable) -> void:
	if handler.is_null():
		return
	if multiplayer_lobby.has_signal(signal_name) && !multiplayer_lobby.is_connected(signal_name, handler):
		multiplayer_lobby.connect(signal_name, handler)

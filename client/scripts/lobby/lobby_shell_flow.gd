extends RefCounted

var lobby_flow
var session_context
var lobby_network_actions
var lobby_return_flow
var multiplayer_lobby_presenter
var main_menu: Control
var canvas_layer: CanvasLayer
var logger: Callable


func _init(
	lobby_flow_ref,
	session_context_ref,
	lobby_network_actions_ref,
	lobby_return_flow_ref,
	multiplayer_lobby_presenter_ref,
	main_menu_ref: Control,
	canvas_layer_ref: CanvasLayer,
	logger_callable: Callable
) -> void:
	lobby_flow = lobby_flow_ref
	session_context = session_context_ref
	lobby_network_actions = lobby_network_actions_ref
	lobby_return_flow = lobby_return_flow_ref
	multiplayer_lobby_presenter = multiplayer_lobby_presenter_ref
	main_menu = main_menu_ref
	canvas_layer = canvas_layer_ref
	logger = logger_callable


func apply_room_snapshot(packet: Dictionary) -> void:
	lobby_flow.apply_room_snapshot(packet)
	var state = lobby_flow.current_state()
	session_context.activate_requested_mode()
	if session_context.should_show_multiplayer_lobby(state.room_state):
		_show_multiplayer_lobby(state)
	else:
		if multiplayer_lobby_presenter != null:
			multiplayer_lobby_presenter.clear_lobby()


func _show_multiplayer_lobby(state) -> void:
	if main_menu != null:
		main_menu.hide()
	var callbacks := {
		"ready_requested": Callable(self, "_on_lobby_ready_requested"),
		"start_game_requested": Callable(self, "_on_lobby_start_game_requested"),
		"add_bot_requested": Callable(self, "_on_lobby_add_bot_requested"),
		"remove_member_requested": Callable(self, "_on_lobby_remove_member_requested"),
		"team_assignment_requested": Callable(self, "_on_lobby_team_assignment_requested"),
		"leave_requested": Callable(self, "_on_lobby_leave_requested"),
	}
	multiplayer_lobby_presenter.show_lobby(canvas_layer, state, callbacks)


func _on_lobby_ready_requested(ready: bool) -> void:
	lobby_network_actions.send_ready_requested(ready)


func _on_lobby_start_game_requested() -> void:
	lobby_network_actions.send_start_game_requested()


func _on_lobby_add_bot_requested() -> void:
	lobby_network_actions.send_add_bot_requested()


func _on_lobby_remove_member_requested(player_id: String) -> void:
	lobby_network_actions.send_remove_member_requested(player_id)


func _on_lobby_team_assignment_requested(player_id: String, team_id: String) -> void:
	lobby_network_actions.send_team_assignment_requested(player_id, team_id)


func _on_lobby_leave_requested() -> void:
	lobby_network_actions.send_leave_requested()
	lobby_return_flow.return_after_leave()

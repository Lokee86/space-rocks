extends RefCounted

var lobby_flow
var session_context
var connection_service
var multiplayer_lobby_presenter
var main_menu: Control
var canvas_layer: CanvasLayer
var logger: Callable
var return_to_menu_callback: Callable


func _init(
	lobby_flow_ref,
	session_context_ref,
	connection_service_ref,
	multiplayer_lobby_presenter_ref,
	main_menu_ref: Control,
	canvas_layer_ref: CanvasLayer,
	logger_callable: Callable,
	return_to_menu_callable: Callable
) -> void:
	lobby_flow = lobby_flow_ref
	session_context = session_context_ref
	connection_service = connection_service_ref
	multiplayer_lobby_presenter = multiplayer_lobby_presenter_ref
	main_menu = main_menu_ref
	canvas_layer = canvas_layer_ref
	logger = logger_callable
	return_to_menu_callback = return_to_menu_callable


func apply_room_snapshot(packet: Dictionary) -> void:
	var summary := lobby_flow.apply_room_snapshot(packet)
	_log("V2 lobby updated: %s" % summary)
	var state = lobby_flow.current_state()
	session_context.activate_requested_mode()
	if session_context.should_show_multiplayer_lobby(state.room_state):
		_show_multiplayer_lobby(state)
	else:
		_log("V2 room snapshot received; multiplayer lobby mount skipped for session mode")


func clear_lobby_and_show_main_menu() -> void:
	if lobby_flow != null:
		lobby_flow.clear()
	multiplayer_lobby_presenter.clear_lobby()
	if main_menu != null:
		main_menu.show()


func _show_multiplayer_lobby(state) -> void:
	if main_menu != null:
		main_menu.hide()
	var callbacks := {
		"ready_requested": Callable(self, "_on_lobby_ready_requested"),
		"start_game_requested": Callable(self, "_on_lobby_start_game_requested"),
		"leave_requested": Callable(self, "_on_lobby_leave_requested"),
	}
	multiplayer_lobby_presenter.show_lobby(canvas_layer, state, callbacks)


func _on_lobby_ready_requested(ready: bool) -> void:
	connection_service.send_set_ready_request(ready)
	_log("V2 lobby ready requested: %s" % ready)


func _on_lobby_start_game_requested() -> void:
	connection_service.send_start_game_request()
	_log("V2 lobby start game requested")


func _on_lobby_leave_requested() -> void:
	connection_service.send_leave_room_request()
	_log("V2 lobby leave requested")
	clear_lobby_and_show_main_menu()
	if !return_to_menu_callback.is_null():
		return_to_menu_callback.call()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)

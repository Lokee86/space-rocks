extends RefCounted

var connection_service
var shell_boot_flow
var room_session_controller
var gameplay_session_controller
var logger: Callable
var handlers := {}


func configure(
	connection_service_ref,
	shell_boot_flow_ref,
	logger_callable: Callable,
	handlers_ref: Dictionary
) -> void:
	connection_service = connection_service_ref
	shell_boot_flow = shell_boot_flow_ref
	logger = logger_callable
	handlers = handlers_ref


func configure_room_session_controller(room_session_controller_ref) -> void:
	room_session_controller = room_session_controller_ref


func configure_gameplay_session_controller(gameplay_session_controller_ref) -> void:
	gameplay_session_controller = gameplay_session_controller_ref


func connect_connection_signals() -> void:
	_connect_connection_signal("connected", Callable(self, "_on_connection_connected"))
	_connect_connection_signal("closed", Callable(self, "_on_connection_closed"))
	_connect_connection_signal("packet_parse_failed", Callable(self, "_on_packet_parse_failed"))
	_connect_connection_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))


func connect_room_signals() -> void:
	_connect_connection_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_connection_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_connection_signal("room_error_received", Callable(self, "_on_room_error_received"))


func connect_gameplay_signals() -> void:
	_connect_connection_signal("gameplay_state_received", Callable(self, "_on_gameplay_state_received"))


func _connect_connection_signal(signal_name: StringName, handler: Callable) -> void:
	if connection_service == null:
		return
	if !connection_service.has_signal(signal_name):
		return
	if connection_service.is_connected(signal_name, handler):
		return

	connection_service.connect(signal_name, handler)


func _on_connection_connected() -> void:
	_log("V2 connection connected")
	shell_boot_flow.send_pending_boot_request()


func _on_connection_closed() -> void:
	_log("V2 connection closed")


func _on_packet_parse_failed(text: String) -> void:
	_log("V2 packet parse failed: %s" % text)


func _on_unknown_packet_received(_packet: Dictionary) -> void:
	_log("V2 unknown packet received")


func _on_room_snapshot_received(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_snapshot(packet)
	_refresh_game_over_menu_state()


func _on_room_state_changed(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_state_changed(packet)
	_refresh_game_over_menu_state()


func _on_room_error_received(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_error(packet)


func _on_gameplay_state_received(packet: Dictionary) -> void:
	if gameplay_session_controller == null:
		return
	gameplay_session_controller.handle_gameplay_state(packet)


func _refresh_game_over_menu_state() -> void:
	if gameplay_session_controller != null && gameplay_session_controller.has_method("refresh_game_over_menu_state"):
		gameplay_session_controller.refresh_game_over_menu_state()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)

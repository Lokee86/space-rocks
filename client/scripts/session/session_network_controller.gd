extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")

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
	_connect_connection_signal("websocket_auth_result_received", Callable(self, "_on_websocket_auth_result_received"))


func connect_room_signals() -> void:
	_connect_connection_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_connection_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_connection_signal("room_error_received", Callable(self, "_on_room_error_received"))


func connect_gameplay_signals() -> void:
	_connect_connection_signal("gameplay_packet_received", Callable(self, "_on_gameplay_packet_received"))
	_connect_connection_signal("debug_shape_catalog_received", Callable(self, "_on_debug_shape_catalog_received"))
	_connect_connection_signal("debug_status_received", Callable(self, "_on_debug_status_received"))
	_connect_connection_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))


func _connect_connection_signal(signal_name: StringName, handler: Callable) -> void:
	if connection_service == null:
		return
	if !connection_service.has_signal(signal_name):
		return
	if connection_service.is_connected(signal_name, handler):
		return

	connection_service.connect(signal_name, handler)


func _on_connection_connected() -> void:
	_log("Connection connected")
	if shell_boot_flow == null:
		return

	if shell_boot_flow.pending_request_is_single_player():
		shell_boot_flow.send_pending_boot_request()
		return

	if shell_boot_flow.pending_request_is_multiplayer():
		if connection_service.is_websocket_auth_authenticated():
			shell_boot_flow.send_pending_boot_request()
		else:
			_log("Waiting for websocket auth before sending multiplayer boot request")


func _on_connection_closed() -> void:
	_log("Connection closed")


func _on_packet_parse_failed(text: String) -> void:
	_log("Packet parse failed: %s" % text)


func _on_unknown_packet_received(_packet: Dictionary) -> void:
	_log("Unknown packet received")


func _on_websocket_auth_result_received(packet: Dictionary) -> void:
	if shell_boot_flow == null:
		return

	if bool(packet.get("authenticated", false)):
		shell_boot_flow.send_pending_boot_request()
	else:
		var error_code := str(packet.get("error_code", ""))
		if error_code == "token_verification_unavailable":
			_log("Websocket auth unavailable; sending pending multiplayer request for server-side admission")
			shell_boot_flow.send_pending_boot_request()
		else:
			_log("Websocket auth failed before multiplayer boot request")


func _on_room_snapshot_received(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_snapshot(packet)
	if room_session_controller.current_room_state() == Constants.ROOM_STATE_IN_GAME && gameplay_session_controller != null:
		gameplay_session_controller.begin_accepting_gameplay_packets()
	_refresh_match_end_state()


func _on_room_state_changed(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_state_changed(packet)
	if room_session_controller.current_room_state() == Constants.ROOM_STATE_IN_GAME && gameplay_session_controller != null:
		gameplay_session_controller.begin_accepting_gameplay_packets()
	_refresh_match_end_state()


func _on_room_error_received(packet: Dictionary) -> void:
	if room_session_controller == null:
		return
	room_session_controller.handle_room_error(packet)


func _on_gameplay_packet_received(packet: Dictionary) -> void:
	if gameplay_session_controller == null:
		return
	gameplay_session_controller.handle_gameplay_packet(packet)


func _on_debug_shape_catalog_received(packet: Dictionary) -> void:
	if gameplay_session_controller == null:
		return
	gameplay_session_controller.handle_debug_shape_catalog_packet(packet)


func _on_debug_status_received(packet: Dictionary) -> void:
	if gameplay_session_controller == null:
		return
	if gameplay_session_controller.has_method("handle_debug_status_packet"):
		gameplay_session_controller.handle_debug_status_packet(packet)


func _on_player_pause_state_received(packet: Dictionary) -> void:
	if gameplay_session_controller == null:
		return
	gameplay_session_controller.handle_player_pause_state(packet)


func _refresh_match_end_state() -> void:
	if gameplay_session_controller != null && gameplay_session_controller.has_method("refresh_match_end_state"):
		gameplay_session_controller.refresh_match_end_state()


func _log(message: String) -> void:
	if !logger.is_null():
		logger.call(message)


func get_protocol_mode() -> String:
	if connection_service == null:
		return "legacy_state"
	return connection_service.get_protocol_mode()


func set_protocol_mode(protocol_mode: String) -> void:
	if connection_service == null:
		return
	connection_service.set_protocol_mode(protocol_mode)

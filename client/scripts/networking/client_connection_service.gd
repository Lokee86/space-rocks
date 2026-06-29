extends Node

const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
const RealtimeRouter := preload("res://scripts/protocol/realtime/realtime_router.gd")
const Constants := preload("res://scripts/generated/constants/constants.gd")
const Packets := preload("res://scripts/generated/networking/packets/packets.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")

signal connected
signal closed
signal packet_parse_failed(text: String)
signal room_snapshot_received(packet: Dictionary)
signal websocket_auth_result_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal world_full_received(packet: Dictionary)
signal world_delta_received(packet: Dictionary)
signal overlay_full_received(packet: Dictionary)
signal overlay_delta_received(packet: Dictionary)
signal session_full_received(packet: Dictionary)
signal session_delta_received(packet: Dictionary)
signal event_batch_received(packet: Dictionary)
signal resync_request_received(packet: Dictionary)
signal resync_required_received(packet: Dictionary)
signal debug_shape_catalog_received(packet: Dictionary)
signal debug_status_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)
signal gameplay_packet_received(packet: Dictionary)

var network_client: NetworkClient
var client_packet_sender: ClientPacketSender
var server_packet_dispatcher: ServerPacketDispatcher
var realtime_router: RealtimeRouter
var has_started_connection := false
var auth_session_controller
var websocket_auth_authenticated := false
var websocket_auth_user_id = null
var websocket_auth_display_name := ""
var _lane_route_log_emitted := {}


func _ready() -> void:
	process_priority = Constants.NETWORK_POLL_PROCESS_PRIORITY
	network_client = NetworkClient.new()
	client_packet_sender = ClientPacketSender.new(network_client)
	server_packet_dispatcher = ServerPacketDispatcher.new()
	realtime_router = RealtimeRouter.new()
	add_child(network_client)
	add_child(server_packet_dispatcher)
	_connect_server_packet_dispatcher_signals()
	_connect_network_client_signals()


func _process(_delta: float) -> void:
	if has_started_connection && network_client != null:
		network_client.poll()


func connect_to_server(url: String) -> Error:
	reset_realtime_protocol_state()
	has_started_connection = true
	return network_client.connect_to_server(url)


func reset_realtime_protocol_state() -> void:
	realtime_router = RealtimeRouter.new()
	_lane_route_log_emitted.clear()
	ClientLogger.network_info("realtime protocol state reset")


func is_server_connected() -> bool:
	return network_client != null && network_client.is_connected_to_server()


func is_websocket_auth_authenticated() -> bool:
	return websocket_auth_authenticated


func has_websocket_auth_identity() -> bool:
	return websocket_auth_authenticated && websocket_auth_user_id != null


func begin_graceful_close() -> void:
	if network_client != null:
		network_client.begin_graceful_close()


func close_gracefully() -> void:
	if network_client != null:
		await network_client.close_gracefully()


func set_auth_session_controller(auth_session_controller_ref) -> void:
	auth_session_controller = auth_session_controller_ref


func send_start_single_player_request(local_profile_id := "") -> void:
	if client_packet_sender != null:
		client_packet_sender.send_start_single_player_request(local_profile_id)


func send_create_room_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_create_room_request()


func send_join_room_request(room_code: String) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_join_room_request(room_code)


func send_set_ready_request(is_ready: bool) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_set_ready_request(is_ready)


func send_start_game_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_start_game_request()


func send_input_packet(packet: Dictionary) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_input_packet(packet)


func send_packet(packet: Dictionary) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_packet(packet)


func send_respawn_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_respawn_request()


func send_pause_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_pause_request()


func send_telemetry_ping(sequence: int, client_sent_msec: int) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_telemetry_ping(sequence, client_sent_msec)


func send_debug_kill_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	if client_packet_sender != null:
		client_packet_sender.send_debug_kill_player_request(target_scope, target_player_id)


func send_debug_kill_target_player_request(target_player_id: String, target_scope: String = "") -> void:
	if client_packet_sender != null:
		client_packet_sender.send_debug_kill_target_player_request(target_player_id, target_scope)


func send_leave_room_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_leave_room_request()


func send_return_to_lobby_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_return_to_lobby_request()


func _connect_network_client_signals() -> void:
	_connect_network_signal("connected_to_server", Callable(self, "_on_connected"))
	_connect_network_signal("connection_closed", Callable(self, "_on_closed"))
	_connect_network_signal("packet_parse_failed", Callable(self, "_on_packet_parse_failed"))
	_connect_network_signal("packet_received", Callable(self, "_on_packet_received"))


func _connect_server_packet_dispatcher_signals() -> void:
	_connect_dispatcher_signal("authenticate_result_received", Callable(self, "_on_authenticate_result_received"))
	_connect_dispatcher_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_dispatcher_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_dispatcher_signal("room_error_received", Callable(self, "_on_room_error_received"))
	_connect_dispatcher_signal("world_full_received", Callable(self, "_on_world_full_received"))
	_connect_dispatcher_signal("world_delta_received", Callable(self, "_on_world_delta_received"))
	_connect_dispatcher_signal("overlay_full_received", Callable(self, "_on_overlay_full_received"))
	_connect_dispatcher_signal("overlay_delta_received", Callable(self, "_on_overlay_delta_received"))
	_connect_dispatcher_signal("session_full_received", Callable(self, "_on_session_full_received"))
	_connect_dispatcher_signal("session_delta_received", Callable(self, "_on_session_delta_received"))
	_connect_dispatcher_signal("event_batch_received", Callable(self, "_on_event_batch_received"))
	_connect_dispatcher_signal("resync_request_received", Callable(self, "_on_resync_request_received"))
	_connect_dispatcher_signal("resync_required_received", Callable(self, "_on_resync_required_received"))
	_connect_dispatcher_signal("debug_shape_catalog_received", Callable(self, "_on_debug_shape_catalog_received"))
	_connect_dispatcher_signal("debug_status_received", Callable(self, "_on_debug_status_received"))
	_connect_dispatcher_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))
	_connect_dispatcher_signal("telemetry_pong_received", Callable(self, "_on_telemetry_pong_received"))
	_connect_dispatcher_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))


func _connect_network_signal(signal_name: StringName, handler: Callable) -> void:
	if network_client.has_signal(signal_name):
		network_client.connect(signal_name, handler)


func _connect_dispatcher_signal(signal_name: StringName, handler: Callable) -> void:
	if server_packet_dispatcher.has_signal(signal_name):
		server_packet_dispatcher.connect(signal_name, handler)


func _route_gameplay_packet(packet: Dictionary) -> void:
	if realtime_router == null:
		return

	realtime_router.route_packet_for_protocol_mode(packet, "lane_protocol")
	var packet_type := str(packet.get("type", packet.get("Type", "")))
	if !_lane_route_log_emitted.has(packet_type):
		_lane_route_log_emitted[packet_type] = true
		var readiness := false
		if realtime_router.has_method("is_presentable"):
			readiness = realtime_router.is_presentable()
		elif realtime_router.has_method("get_gameplay_readiness"):
			var gameplay_readiness = realtime_router.get_gameplay_readiness()
			if gameplay_readiness != null and gameplay_readiness.has_method("is_gameplay_ready"):
				readiness = gameplay_readiness.is_gameplay_ready()
		ClientLogger.network_info("lane packet routed: type=%s readiness=%s" % [packet_type, str(readiness)])


func _emit_gameplay_packet(packet: Dictionary) -> void:
	gameplay_packet_received.emit(packet)


func _on_connected() -> void:
	_send_authenticate_request_if_token_exists()
	connected.emit()


func _on_closed() -> void:
	reset_realtime_protocol_state()
	websocket_auth_authenticated = false
	websocket_auth_user_id = null
	websocket_auth_display_name = ""
	closed.emit()


func _on_packet_parse_failed(text: String) -> void:
	packet_parse_failed.emit(text)


func _on_packet_received(packet: Dictionary) -> void:
	if server_packet_dispatcher != null:
		server_packet_dispatcher.dispatch(packet)


func _on_room_snapshot_received(packet: Dictionary) -> void:
	room_snapshot_received.emit(packet)


func _on_authenticate_result_received(packet: Dictionary) -> void:
	websocket_auth_authenticated = bool(packet.get(Packets.FIELD_AUTHENTICATED, false))
	websocket_auth_user_id = packet.get(Packets.FIELD_USER_ID, null)
	websocket_auth_display_name = str(packet.get(Packets.FIELD_DISPLAY_NAME, ""))
	websocket_auth_result_received.emit(packet)


func _on_room_state_changed(packet: Dictionary) -> void:
	room_state_changed.emit(packet)


func _on_room_error_received(packet: Dictionary) -> void:
	room_error_received.emit(packet)


func _on_world_full_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	world_full_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_world_delta_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	world_delta_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_overlay_full_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	overlay_full_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_overlay_delta_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	overlay_delta_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_session_full_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	session_full_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_session_delta_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	session_delta_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_event_batch_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	event_batch_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_resync_request_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	resync_request_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_resync_required_received(packet: Dictionary) -> void:
	_route_gameplay_packet(packet)
	resync_required_received.emit(packet)
	_emit_gameplay_packet(packet)


func _on_debug_shape_catalog_received(packet: Dictionary) -> void:
	debug_shape_catalog_received.emit(packet)


func _on_debug_status_received(packet: Dictionary) -> void:
	debug_status_received.emit(packet)


func _on_player_pause_state_received(packet: Dictionary) -> void:
	player_pause_state_received.emit(packet)


func _on_telemetry_pong_received(packet: Dictionary) -> void:
	telemetry_pong_received.emit(packet)


func _on_unknown_packet_received(packet: Dictionary) -> void:
	unknown_packet_received.emit(packet)


func _send_authenticate_request_if_token_exists() -> void:
	if network_client == null || auth_session_controller == null:
		return

	var auth_session = auth_session_controller.get_session()
	if auth_session == null:
		return

	var token: String = auth_session.token
	if token.is_empty():
		return

	network_client.send_authenticate_request(token)


func get_gameplay_readiness():
	if realtime_router != null:
		return realtime_router.gameplay_readiness
	if server_packet_dispatcher == null:
		return null
	if server_packet_dispatcher.has_method("get_gameplay_readiness"):
		return server_packet_dispatcher.get_gameplay_readiness()
	return null


func get_realtime_router():
	return realtime_router

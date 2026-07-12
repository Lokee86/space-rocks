extends Node

const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
const ClientInboundCoordinator := preload("res://scripts/networking/inbound/client_inbound_coordinator.gd")
const RealtimePacketPipeline := preload("res://scripts/networking/realtime/realtime_packet_pipeline.gd")
const RealtimeTransportSession := preload("res://scripts/networking/webrtc/realtime_transport_session.gd")
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
signal debug_shape_catalog_received(packet: Dictionary)
signal debug_status_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal realtime_transport_ready
signal unknown_packet_received(packet: Dictionary)

var network_client: NetworkClient
var client_packet_sender: ClientPacketSender:
	set(value):
		client_packet_sender = value
		_missing_client_packet_sender_reported = false
var server_packet_dispatcher: ServerPacketDispatcher
var client_inbound_coordinator: ClientInboundCoordinator
var realtime_packet_pipeline: RealtimePacketPipeline
var realtime_transport_session: RealtimeTransportSession
var webrtc_transport_factory: Callable
var server_clock_offset_ms := -1

var has_started_connection := false
var auth_session_controller: AuthSessionController
var websocket_auth_authenticated := false
const NO_WEBSOCKET_AUTH_USER_ID := -1
var websocket_auth_user_id: int = NO_WEBSOCKET_AUTH_USER_ID
var websocket_auth_display_name := ""
var _resync_signal_bound := false
var _missing_client_packet_sender_reported := false


func _ready() -> void:
	process_priority = Constants.NETWORK_POLL_PROCESS_PRIORITY
	if network_client == null:
		network_client = NetworkClient.new()
	if client_packet_sender == null:
		client_packet_sender = ClientPacketSender.new(network_client)
	if server_packet_dispatcher == null:
		server_packet_dispatcher = ServerPacketDispatcher.new()
	if realtime_packet_pipeline == null:
		realtime_packet_pipeline = RealtimePacketPipeline.new()
	if client_inbound_coordinator == null:
		client_inbound_coordinator = ClientInboundCoordinator.new()
	client_inbound_coordinator.configure(server_packet_dispatcher, realtime_packet_pipeline, realtime_transport_session)
	_bind_resync_request_signal()
	_connect_coordinator_signal("authenticate_result_received", Callable(self, "_on_authenticate_result_received"))
	_connect_coordinator_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_coordinator_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_coordinator_signal("room_error_received", Callable(self, "_on_room_error_received"))
	_connect_coordinator_signal("debug_shape_catalog_received", Callable(self, "_on_debug_shape_catalog_received"))
	_connect_coordinator_signal("debug_status_received", Callable(self, "_on_debug_status_received"))
	_connect_coordinator_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))
	_connect_coordinator_signal("telemetry_pong_received", Callable(self, "_on_telemetry_pong_received"))
	_connect_coordinator_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))
	var ready_handler := Callable(self, "_on_realtime_transport_ready")
	if !client_inbound_coordinator.is_connected("realtime_transport_ready", ready_handler):
		client_inbound_coordinator.connect("realtime_transport_ready", ready_handler)
	if network_client != null and network_client.get_parent() == null:
		add_child(network_client)
	if server_packet_dispatcher != null and server_packet_dispatcher.get_parent() == null:
		add_child(server_packet_dispatcher)
	_connect_network_client_signals()


func _process(_delta: float) -> void:
	if has_started_connection && network_client != null:
		network_client.poll()
	if realtime_transport_session != null:
		realtime_transport_session.poll()


func connect_to_server(url: String) -> Error:
	reset_realtime_session()
	has_started_connection = true
	return network_client.connect_to_server(url)


func begin_realtime_match(match_id: String) -> void:
	if realtime_packet_pipeline != null:
		realtime_packet_pipeline.begin_match(match_id)

func end_realtime_match() -> void:
	if realtime_packet_pipeline != null:
		realtime_packet_pipeline.end_match()


func reset_realtime_session() -> void:
	if realtime_packet_pipeline != null:
		realtime_packet_pipeline.reset()
	_clear_realtime_transport_session()
	server_clock_offset_ms = -1
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"realtime_protocol_state_reset",
		"Realtime protocol state reset"
	)


func is_server_connected() -> bool:
	return network_client != null && network_client.is_connected_to_server()


func is_websocket_auth_authenticated() -> bool:
	return websocket_auth_authenticated


func has_websocket_auth_identity() -> bool:
	return websocket_auth_authenticated && websocket_auth_user_id != NO_WEBSOCKET_AUTH_USER_ID


func begin_graceful_close() -> void:
	if network_client != null:
		network_client.begin_graceful_close()


func close_gracefully() -> void:
	if network_client != null:
		await network_client.close_gracefully()


func set_auth_session_controller(auth_session_controller_ref: AuthSessionController) -> void:
	auth_session_controller = auth_session_controller_ref


func send_start_single_player_request(local_profile_id := "") -> void:
	if _can_send_outbound():
		client_packet_sender.send_start_single_player_request(local_profile_id)


func send_create_room_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_create_room_request()


func send_join_room_request(room_code: String) -> void:
	if _can_send_outbound():
		client_packet_sender.send_join_room_request(room_code)


func send_set_ready_request(is_ready: bool) -> void:
	if _can_send_outbound():
		client_packet_sender.send_set_ready_request(is_ready)


func send_start_game_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_start_game_request()


func send_input_packet(packet: Dictionary) -> void:
	if _can_send_outbound():
		client_packet_sender.send_input_packet(packet)

func _bind_resync_request_signal() -> void:
	if _resync_signal_bound or realtime_packet_pipeline == null:
		return
	realtime_packet_pipeline.resync_request_required.connect(_on_resync_request_required)
	_resync_signal_bound = true

func _on_resync_request_required(lane, baseline_id, sequence, reason) -> void:
	if realtime_packet_pipeline == null:
		return
	var match_id := realtime_packet_pipeline.active_match_id()
	if match_id.is_empty():
		return
	if _can_send_outbound():
		client_packet_sender.send_resync_request(match_id, lane, baseline_id, sequence, reason)


func send_packet(packet: Dictionary) -> void:
	if _can_send_outbound():
		client_packet_sender.send_packet(packet)


func send_webrtc_offer(description_type: String, sdp: String) -> void:
	if _can_send_outbound():
		client_packet_sender.send_webrtc_offer(description_type, sdp)


func send_webrtc_ice_candidate(media: String, index: int, name: String) -> void:
	if _can_send_outbound():
		client_packet_sender.send_webrtc_ice_candidate(media, index, name)


func send_webrtc_smoke(smoke_id: String, origin: String, message: String) -> void:
	if _can_send_outbound():
		client_packet_sender.send_webrtc_smoke(smoke_id, origin, message)


func send_webrtc_failed(error_code: String, message: String) -> void:
	if _can_send_outbound():
		client_packet_sender.send_webrtc_failed(error_code, message)


func send_respawn_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_respawn_request()


func send_pause_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_pause_request()


func send_telemetry_ping(sequence: int, client_sent_msec: int) -> void:
	if _can_send_outbound():
		client_packet_sender.send_telemetry_ping(sequence, client_sent_msec)


func send_debug_kill_player_request(target_scope: String = "", target_player_id: String = "") -> void:
	if _can_send_outbound():
		client_packet_sender.send_debug_kill_player_request(target_scope, target_player_id)


func send_debug_kill_target_player_request(target_player_id: String, target_scope: String = "") -> void:
	if _can_send_outbound():
		client_packet_sender.send_debug_kill_target_player_request(target_player_id, target_scope)


func send_leave_room_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_leave_room_request()


func send_return_to_lobby_request() -> void:
	if _can_send_outbound():
		client_packet_sender.send_return_to_lobby_request()


func network_metrics_snapshot() -> Dictionary:
	if network_client != null and network_client.has_method("network_metrics_snapshot"):
		return network_client.network_metrics_snapshot()
	return {}


func _connect_network_client_signals() -> void:
	_connect_network_signal("connected_to_server", Callable(self, "_on_connected"))
	_connect_network_signal("connection_closed", Callable(self, "_on_closed"))
	_connect_network_signal("packet_parse_failed", Callable(self, "_on_packet_parse_failed"))
	_connect_network_signal("packet_received", Callable(self, "_on_packet_received"))


func _connect_network_signal(signal_name: StringName, handler: Callable) -> void:
	if network_client.has_signal(signal_name):
		network_client.connect(signal_name, handler)


func _connect_coordinator_signal(signal_name: StringName, handler: Callable) -> void:
	if client_inbound_coordinator.has_signal(signal_name) and !client_inbound_coordinator.is_connected(signal_name, handler):
		client_inbound_coordinator.connect(signal_name, handler)


func _on_connected() -> void:
	_ensure_realtime_transport_session()
	if realtime_transport_session != null:
		realtime_transport_session.start()
	_send_authenticate_request_if_token_exists()
	connected.emit()


func _on_closed() -> void:
	reset_realtime_session()
	websocket_auth_authenticated = false
	websocket_auth_user_id = NO_WEBSOCKET_AUTH_USER_ID
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
	var raw_user_id = packet.get(Packets.FIELD_USER_ID, NO_WEBSOCKET_AUTH_USER_ID)
	if websocket_auth_authenticated and raw_user_id is int:
		websocket_auth_user_id = int(raw_user_id)
	else:
		websocket_auth_user_id = NO_WEBSOCKET_AUTH_USER_ID
	websocket_auth_display_name = str(packet.get(Packets.FIELD_DISPLAY_NAME, ""))
	websocket_auth_result_received.emit(packet)


func _on_room_state_changed(packet: Dictionary) -> void:
	room_state_changed.emit(packet)


func _on_room_error_received(packet: Dictionary) -> void:
	room_error_received.emit(packet)


func _on_debug_shape_catalog_received(packet: Dictionary) -> void:
	debug_shape_catalog_received.emit(packet)


func _on_debug_status_received(packet: Dictionary) -> void:
	debug_status_received.emit(packet)


func _on_player_pause_state_received(packet: Dictionary) -> void:
	player_pause_state_received.emit(packet)


func _on_telemetry_pong_received(packet: Dictionary) -> void:
	telemetry_pong_received.emit(packet)


func set_server_clock_offset_ms(offset_ms: int) -> void:
	server_clock_offset_ms = offset_ms
	if realtime_transport_session != null:
		realtime_transport_session.set_server_clock_offset_ms(offset_ms)


func _on_realtime_transport_ready() -> void:
	realtime_transport_ready.emit()


func _on_unknown_packet_received(packet: Dictionary) -> void:
	unknown_packet_received.emit(packet)


func _ensure_realtime_transport_session() -> void:
	if realtime_transport_session != null:
		return
	realtime_transport_session = RealtimeTransportSession.new()
	realtime_transport_session.set_server_clock_offset_ms(server_clock_offset_ms)
	realtime_transport_session.transport_factory = webrtc_transport_factory
	realtime_transport_session.dispatch_packet = func(packet: Dictionary) -> void:
		if server_packet_dispatcher != null:
			server_packet_dispatcher.dispatch(packet)
	realtime_transport_session.send_offer = Callable(self, "send_webrtc_offer")
	realtime_transport_session.send_ice_candidate = Callable(self, "send_webrtc_ice_candidate")
	realtime_transport_session.send_failed = Callable(self, "send_webrtc_failed")
	if client_inbound_coordinator != null:
		client_inbound_coordinator.set_realtime_transport_session(realtime_transport_session)


func _clear_realtime_transport_session() -> void:
	if realtime_transport_session != null:
		realtime_transport_session.close()
		realtime_transport_session = null
	if client_inbound_coordinator != null:
		client_inbound_coordinator.set_realtime_transport_session(null)


func _send_authenticate_request_if_token_exists() -> void:
	if auth_session_controller == null:
		return

	var auth_session: AuthSession = auth_session_controller.get_session()
	if auth_session == null:
		return

	var token: String = auth_session.token
	if token.is_empty():
		return

	if _can_send_outbound():
		client_packet_sender.send_authenticate_request(token)


func _can_send_outbound() -> bool:
	if client_packet_sender != null:
		return true
	if !_missing_client_packet_sender_reported:
		_missing_client_packet_sender_reported = true
		ClientLogger.error(ClientLogger.CATEGORY_NETWORK, "Client packet sender is not configured")
	return false


func get_realtime_packet_pipeline() -> RealtimePacketPipeline:
	return realtime_packet_pipeline

extends Node

const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
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
signal webrtc_answer_received(packet: Dictionary)
signal webrtc_ice_candidate_received(packet: Dictionary)
signal webrtc_ready_received(packet: Dictionary)
signal webrtc_smoke_received(packet: Dictionary)
signal webrtc_failed_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)

var network_client: NetworkClient
var client_packet_sender: ClientPacketSender
var server_packet_dispatcher: ServerPacketDispatcher
var realtime_packet_pipeline: RealtimePacketPipeline
var realtime_transport_session: RealtimeTransportSession
var webrtc_transport_factory: Callable

var has_started_connection := false
var auth_session_controller
var websocket_auth_authenticated := false
var websocket_auth_user_id = null
var websocket_auth_display_name := ""


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
	if network_client != null and network_client.get_parent() == null:
		add_child(network_client)
	if server_packet_dispatcher != null and server_packet_dispatcher.get_parent() == null:
		add_child(server_packet_dispatcher)
	_connect_server_packet_dispatcher_signals()
	_connect_network_client_signals()


func _process(_delta: float) -> void:
	if has_started_connection && network_client != null:
		network_client.poll()
	if realtime_transport_session != null:
		realtime_transport_session.poll()


func connect_to_server(url: String) -> Error:
	reset_realtime_protocol_state()
	has_started_connection = true
	return network_client.connect_to_server(url)


func reset_realtime_protocol_state() -> void:
	if realtime_packet_pipeline != null:
		realtime_packet_pipeline.reset()
	_clear_webrtc_transport()
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


func send_webrtc_offer(description_type: String, sdp: String) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_webrtc_offer(description_type, sdp)


func send_webrtc_ice_candidate(media: String, index: int, name: String) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_webrtc_ice_candidate(media, index, name)


func send_webrtc_smoke(smoke_id: String, origin: String, message: String) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_webrtc_smoke(smoke_id, origin, message)


func send_webrtc_failed(error_code: String, message: String) -> void:
	if client_packet_sender != null:
		client_packet_sender.send_webrtc_failed(error_code, message)


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


func network_metrics_snapshot() -> Dictionary:
	if network_client != null and network_client.has_method("network_metrics_snapshot"):
		return network_client.network_metrics_snapshot()
	return {}


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
	_connect_dispatcher_signal("world_full_received", Callable(realtime_packet_pipeline, "apply_world_full"))
	_connect_dispatcher_signal("world_delta_received", Callable(realtime_packet_pipeline, "apply_world_delta"))
	_connect_dispatcher_signal("asteroid_delta_received", Callable(realtime_packet_pipeline, "apply_asteroid_delta"))
	_connect_dispatcher_signal("bullet_delta_received", Callable(realtime_packet_pipeline, "apply_bullet_delta"))
	_connect_dispatcher_signal("asteroids_lifecycle_received", Callable(realtime_packet_pipeline, "apply_asteroids_lifecycle"))
	_connect_dispatcher_signal("bullets_lifecycle_received", Callable(realtime_packet_pipeline, "apply_bullets_lifecycle"))
	_connect_dispatcher_signal("overlay_full_received", Callable(realtime_packet_pipeline, "apply_overlay_full"))
	_connect_dispatcher_signal("overlay_delta_received", Callable(realtime_packet_pipeline, "apply_overlay_delta"))
	_connect_dispatcher_signal("session_full_received", Callable(realtime_packet_pipeline, "apply_session_full"))
	_connect_dispatcher_signal("session_delta_received", Callable(realtime_packet_pipeline, "apply_session_delta"))
	_connect_dispatcher_signal("event_batch_received", Callable(realtime_packet_pipeline, "apply_event_batch"))
	_connect_dispatcher_signal("resync_request_received", Callable(realtime_packet_pipeline, "apply_resync_request"))
	_connect_dispatcher_signal("resync_required_received", Callable(realtime_packet_pipeline, "apply_resync_required"))
	_connect_dispatcher_signal("debug_shape_catalog_received", Callable(self, "_on_debug_shape_catalog_received"))
	_connect_dispatcher_signal("debug_status_received", Callable(self, "_on_debug_status_received"))
	_connect_dispatcher_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))
	_connect_dispatcher_signal("telemetry_pong_received", Callable(self, "_on_telemetry_pong_received"))
	_connect_dispatcher_signal("webrtc_answer_received", Callable(self, "_on_webrtc_answer_received"))
	_connect_dispatcher_signal("webrtc_ice_candidate_received", Callable(self, "_on_webrtc_ice_candidate_received"))
	_connect_dispatcher_signal("webrtc_ready_received", Callable(self, "_on_webrtc_ready_received"))
	_connect_dispatcher_signal("webrtc_smoke_received", Callable(self, "_on_webrtc_smoke_received"))
	_connect_dispatcher_signal("webrtc_failed_received", Callable(self, "_on_webrtc_failed_received"))
	_connect_dispatcher_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))


func _connect_network_signal(signal_name: StringName, handler: Callable) -> void:
	if network_client.has_signal(signal_name):
		network_client.connect(signal_name, handler)


func _connect_dispatcher_signal(signal_name: StringName, handler: Callable) -> void:
	if server_packet_dispatcher.has_signal(signal_name):
		server_packet_dispatcher.connect(signal_name, handler)


func _on_connected() -> void:
	_ensure_realtime_transport_session()
	if realtime_transport_session != null:
		realtime_transport_session.start()
	_send_authenticate_request_if_token_exists()
	connected.emit()


func _on_closed() -> void:
	reset_realtime_protocol_state()
	websocket_auth_authenticated = false
	websocket_auth_user_id = null
	websocket_auth_display_name = ""
	_clear_webrtc_transport()
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


func _on_debug_shape_catalog_received(packet: Dictionary) -> void:
	debug_shape_catalog_received.emit(packet)


func _on_debug_status_received(packet: Dictionary) -> void:
	debug_status_received.emit(packet)


func _on_player_pause_state_received(packet: Dictionary) -> void:
	player_pause_state_received.emit(packet)


func _on_telemetry_pong_received(packet: Dictionary) -> void:
	telemetry_pong_received.emit(packet)


func _on_webrtc_answer_received(packet: Dictionary) -> void:
	var description_type := str(packet.get("description_type", ""))
	var sdp := str(packet.get("sdp", ""))
	if realtime_transport_session != null:
		realtime_transport_session.handle_answer(description_type, sdp)
	webrtc_answer_received.emit(packet)


func _on_webrtc_ice_candidate_received(packet: Dictionary) -> void:
	if realtime_transport_session != null:
		realtime_transport_session.handle_remote_ice(str(packet.get("media", "")), int(packet.get("index", 0)), str(packet.get("name", "")))
	webrtc_ice_candidate_received.emit(packet)


func _on_webrtc_ready_received(packet: Dictionary) -> void:
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"webrtc_ready_received",
		"WebRTC ready packet received",
		packet
	)
	webrtc_ready_received.emit(packet)


func _on_webrtc_smoke_received(packet: Dictionary) -> void:
	ClientLogger.network_event(
		ClientLogger.LEVEL_INFO,
		"webrtc_smoke_received",
		"WebRTC smoke packet received",
		packet
	)
	webrtc_smoke_received.emit(packet)


func _on_webrtc_failed_received(packet: Dictionary) -> void:
	if realtime_transport_session != null:
		realtime_transport_session.handle_remote_failure()
	webrtc_failed_received.emit(packet)


func _on_unknown_packet_received(packet: Dictionary) -> void:
	unknown_packet_received.emit(packet)


func _ensure_realtime_transport_session() -> void:
	if realtime_transport_session != null:
		return
	realtime_transport_session = RealtimeTransportSession.new()
	realtime_transport_session.transport_factory = webrtc_transport_factory
	realtime_transport_session.dispatch_packet = func(packet: Dictionary) -> void:
		if server_packet_dispatcher != null:
			server_packet_dispatcher.dispatch(packet)
	realtime_transport_session.send_offer = Callable(self, "send_webrtc_offer")
	realtime_transport_session.send_ice_candidate = Callable(self, "send_webrtc_ice_candidate")
	realtime_transport_session.send_failed = Callable(self, "send_webrtc_failed")


func _clear_webrtc_transport() -> void:
	if realtime_transport_session != null:
		realtime_transport_session.close()
		realtime_transport_session = null


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


func get_realtime_packet_pipeline() -> RealtimePacketPipeline:
	return realtime_packet_pipeline

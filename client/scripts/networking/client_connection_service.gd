extends Node

const NetworkClient := preload("res://scripts/networking/network_client.gd")
const ClientPacketSender := preload("res://scripts/networking/outbound/client_packet_sender.gd")
const ServerPacketDispatcher := preload("res://scripts/networking/inbound/server_packet_dispatcher.gd")
const Constants := preload("res://scripts/constants/constants.gd")

signal connected
signal closed
signal packet_parse_failed(text: String)
signal room_snapshot_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal gameplay_state_received(packet: Dictionary)
signal player_pause_state_received(packet: Dictionary)
signal telemetry_pong_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)

var network_client: NetworkClient
var client_packet_sender: ClientPacketSender
var server_packet_dispatcher: ServerPacketDispatcher
var has_started_connection := false


func _ready() -> void:
	process_priority = Constants.NETWORK_POLL_PROCESS_PRIORITY
	network_client = NetworkClient.new()
	client_packet_sender = ClientPacketSender.new(network_client)
	server_packet_dispatcher = ServerPacketDispatcher.new()
	add_child(network_client)
	add_child(server_packet_dispatcher)
	_connect_server_packet_dispatcher_signals()
	_connect_network_client_signals()


func _process(_delta: float) -> void:
	if has_started_connection && network_client != null:
		network_client.poll()


func connect_to_server(url: String) -> Error:
	has_started_connection = true
	return network_client.connect_to_server(url)


func is_server_connected() -> bool:
	return network_client != null && network_client.is_connected_to_server()


func begin_graceful_close() -> void:
	if network_client != null:
		network_client.begin_graceful_close()


func send_start_single_player_request() -> void:
	if client_packet_sender != null:
		client_packet_sender.send_start_single_player_request()


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
	_connect_dispatcher_signal("room_snapshot_received", Callable(self, "_on_room_snapshot_received"))
	_connect_dispatcher_signal("room_state_changed", Callable(self, "_on_room_state_changed"))
	_connect_dispatcher_signal("room_error_received", Callable(self, "_on_room_error_received"))
	_connect_dispatcher_signal("gameplay_state_received", Callable(self, "_on_gameplay_state_received"))
	_connect_dispatcher_signal("player_pause_state_received", Callable(self, "_on_player_pause_state_received"))
	_connect_dispatcher_signal("telemetry_pong_received", Callable(self, "_on_telemetry_pong_received"))
	_connect_dispatcher_signal("unknown_packet_received", Callable(self, "_on_unknown_packet_received"))


func _connect_network_signal(signal_name: StringName, handler: Callable) -> void:
	if network_client.has_signal(signal_name):
		network_client.connect(signal_name, handler)


func _connect_dispatcher_signal(signal_name: StringName, handler: Callable) -> void:
	if server_packet_dispatcher.has_signal(signal_name):
		server_packet_dispatcher.connect(signal_name, handler)


func _on_connected() -> void:
	connected.emit()


func _on_closed() -> void:
	closed.emit()


func _on_packet_parse_failed(text: String) -> void:
	packet_parse_failed.emit(text)


func _on_packet_received(packet: Dictionary) -> void:
	if server_packet_dispatcher != null:
		server_packet_dispatcher.dispatch(packet)


func _on_room_snapshot_received(packet: Dictionary) -> void:
	room_snapshot_received.emit(packet)


func _on_room_state_changed(packet: Dictionary) -> void:
	room_state_changed.emit(packet)


func _on_room_error_received(packet: Dictionary) -> void:
	room_error_received.emit(packet)


func _on_gameplay_state_received(packet: Dictionary) -> void:
	gameplay_state_received.emit(packet)


func _on_player_pause_state_received(packet: Dictionary) -> void:
	player_pause_state_received.emit(packet)


func _on_telemetry_pong_received(packet: Dictionary) -> void:
	telemetry_pong_received.emit(packet)


func _on_unknown_packet_received(packet: Dictionary) -> void:
	unknown_packet_received.emit(packet)

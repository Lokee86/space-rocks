extends Node

const NetworkClient := preload("res://scripts/networking/network_client.gd")
const ServerPacketRouter := preload("res://scripts/networking/packets/server_packet_router.gd")
const ClientLogger := preload("res://scripts/logging/logger.gd")
const Packets := preload("res://scripts/networking/packets/packets.gd")
const Constants := preload("res://scripts/constants/constants.gd")

signal connected
signal closed
signal packet_parse_failed(text: String)
signal room_snapshot_received(packet: Dictionary)
signal room_state_changed(packet: Dictionary)
signal room_error_received(packet: Dictionary)
signal gameplay_state_received(packet: Dictionary)
signal unknown_packet_received(packet: Dictionary)

var network_client: NetworkClient
var has_started_connection := false


func _ready() -> void:
	process_priority = Constants.NETWORK_POLL_PROCESS_PRIORITY
	network_client = NetworkClient.new()
	add_child(network_client)
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
	if network_client != null:
		network_client.send_start_single_player_request()


func send_create_room_request() -> void:
	if network_client != null:
		network_client.send_create_room_request()


func send_join_room_request(room_code: String) -> void:
	if network_client != null:
		network_client.send_join_room_request(room_code)


func send_set_ready_request(ready: bool) -> void:
	if network_client != null:
		network_client.send_set_ready_request(ready)


func send_start_game_request() -> void:
	if network_client != null:
		network_client.send_start_game_request()


func send_input_packet(packet: Dictionary) -> void:
	if network_client != null:
		network_client.send_packet(packet)


func send_packet(packet: Dictionary) -> void:
	if network_client != null:
		network_client.send_packet(packet)


func send_respawn_request() -> void:
	if network_client != null:
		network_client.send_packet(Packets.respawn_packet())


func send_pause_player_request() -> void:
	send_packet(Packets.pause_player_packet())
	ClientLogger.network_debug("PausePlayerRequest sent")


func send_resume_player_request() -> void:
	send_packet(Packets.resume_player_packet())
	ClientLogger.network_debug("ResumePlayerRequest sent")


func send_leave_room_request() -> void:
	if network_client != null:
		network_client.send_leave_room_request()


func send_return_to_lobby_request() -> void:
	if network_client != null:
		network_client.send_packet(Packets.return_to_lobby_request_packet())
	ClientLogger.network_debug("ReturnToLobbyRequest sent")


func _connect_network_client_signals() -> void:
	_connect_network_signal("connected_to_server", Callable(self, "_on_connected"))
	_connect_network_signal("connection_closed", Callable(self, "_on_closed"))
	_connect_network_signal("packet_parse_failed", Callable(self, "_on_packet_parse_failed"))
	_connect_network_signal("packet_received", Callable(self, "_on_packet_received"))


func _connect_network_signal(signal_name: StringName, handler: Callable) -> void:
	if network_client.has_signal(signal_name):
		network_client.connect(signal_name, handler)


func _on_connected() -> void:
	connected.emit()


func _on_closed() -> void:
	closed.emit()


func _on_packet_parse_failed(text: String) -> void:
	packet_parse_failed.emit(text)


func _on_packet_received(packet: Dictionary) -> void:
	if ServerPacketRouter.is_room_snapshot(packet):
		room_snapshot_received.emit(packet)
	elif ServerPacketRouter.is_room_state_changed(packet):
		room_state_changed.emit(packet)
	elif ServerPacketRouter.is_room_error(packet):
		room_error_received.emit(packet)
	elif ServerPacketRouter.is_gameplay_state(packet):
		gameplay_state_received.emit(packet)
	else:
		unknown_packet_received.emit(packet)

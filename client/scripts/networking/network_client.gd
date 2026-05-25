extends Node
class_name NetworkClient

signal connected_to_server
signal connection_closed
signal packet_received(data: Dictionary)
signal packet_parse_failed(text: String)

const NORMAL_CLOSE_CODE := 1000
const GRACEFUL_CLOSE_TIMEOUT_SECONDS := 0.25
const Packets = preload("res://scripts/networking/packets/packets.gd")
const PacketCodec = preload("res://scripts/networking/packets/packet_codec.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var socket := WebSocketPeer.new()
var connected := false
var closed_notified := false
var closing_gracefully := false


func connect_to_server(url: String) -> Error:
	closing_gracefully = false
	closed_notified = false
	var err := socket.connect_to_url(url)
	if err != OK:
		print("connection failed")
	else:
		print("Connecting...")

	return err


func poll() -> void:
	socket.poll()

	var state := socket.get_ready_state()
	if state == WebSocketPeer.STATE_OPEN:
		if !connected:
			connected = true
			connected_to_server.emit()
	elif state == WebSocketPeer.STATE_CLOSED:
		connected = false
		if !closed_notified && !closing_gracefully:
			closed_notified = true
			connection_closed.emit()

	while socket.get_available_packet_count() > 0:
		var text := socket.get_packet().get_string_from_utf8()
		var data = PacketCodec.decode(text)
		if data == null:
			packet_parse_failed.emit(text)
			continue
		if data is Dictionary:
			packet_received.emit(data)


func send_packet(packet: Dictionary) -> void:
	if !is_connected_to_server():
		return

	socket.send_text(PacketCodec.encode(packet))


func send_create_room_request() -> void:
	send_packet(Packets.create_room_request_packet())


func send_join_room_request(room_code: String) -> void:
	send_packet(Packets.join_room_request_packet(room_code))


func send_leave_room_request() -> void:
	ClientLogger.network_debug("LeaveRoomRequest sent")
	send_packet(Packets.leave_room_request_packet())


func send_set_ready_request(is_ready: bool) -> void:
	send_packet(Packets.set_ready_request_packet(is_ready))


func send_start_game_request() -> void:
	send_packet(Packets.start_game_request_packet())


func send_start_single_player_request() -> void:
	ClientLogger.network_debug("StartSinglePlayerRequest sent")
	send_packet(Packets.start_single_player_request_packet())


func send_return_to_lobby_request() -> void:
	ClientLogger.network_debug("ReturnToLobbyRequest sent")
	send_packet(Packets.return_to_lobby_request_packet())


func close_gracefully() -> void:
	if !begin_graceful_close():
		return

	var elapsed := 0.0
	while socket.get_ready_state() != WebSocketPeer.STATE_CLOSED && elapsed < GRACEFUL_CLOSE_TIMEOUT_SECONDS:
		await get_tree().process_frame
		elapsed += get_process_delta_time()
		socket.poll()


func begin_graceful_close() -> bool:
	var state := socket.get_ready_state()
	if state != WebSocketPeer.STATE_OPEN && state != WebSocketPeer.STATE_CONNECTING:
		return false

	closing_gracefully = true
	closed_notified = true
	connected = false
	socket.close(NORMAL_CLOSE_CODE, "client closed")
	socket.poll()

	return true


func is_connected_to_server() -> bool:
	return socket.get_ready_state() == WebSocketPeer.STATE_OPEN

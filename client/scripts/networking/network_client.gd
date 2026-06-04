extends Node
class_name NetworkClient

const Constants = preload("res://scripts/generated/constants/constants.gd")

signal connected_to_server
signal connection_closed
signal packet_received(data: Dictionary)
signal packet_parse_failed(text: String)

const NORMAL_CLOSE_CODE := 1000
const GRACEFUL_CLOSE_TIMEOUT_SECONDS := 0.25
const PacketCodec = preload("res://scripts/networking/packets/packet_codec.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var socket := WebSocketPeer.new()
var connected := false
var closed_notified := false
var closing_gracefully := false


func connect_to_server(url: String) -> Error:
	closing_gracefully = false
	closed_notified = false
	socket.handshake_headers = PackedStringArray([
		"Origin: %s" % Constants.MULTIPLAYER_WS_ORIGIN
	])
	var err := socket.connect_to_url(url)
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
		var decode_result = PacketCodec.decode(text)
		if !decode_result.ok:
			ClientLogger.network_warn("Packet decode failed: %s" % decode_result.error)
			packet_parse_failed.emit(text)
			continue
		packet_received.emit(decode_result.packet)


func send_raw_packet(packet: Dictionary) -> void:
	if !is_connected_to_server():
		return

	var encode_result = PacketCodec.encode(packet)
	if !encode_result.ok:
		ClientLogger.network_warn("Packet encode failed: %s" % encode_result.error)
		return

	socket.send_text(encode_result.wire_message)


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

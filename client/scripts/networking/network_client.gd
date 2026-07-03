extends Node
class_name NetworkClient

const Constants = preload("res://scripts/generated/constants/constants.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const NetworkRuntimeMetrics = preload("res://scripts/networking/network_runtime_metrics.gd")

signal connected_to_server
signal connection_closed
signal packet_received(data: Dictionary)
signal packet_parse_failed(text: String)

const NORMAL_CLOSE_CODE := 1000
const GRACEFUL_CLOSE_TIMEOUT_SECONDS := 0.25
const PacketCodec = preload("res://scripts/networking/packets/packet_codec.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var socket = WebSocketPeer.new()
var connected := false
var closed_notified := false
var closing_gracefully := false
var runtime_metrics = NetworkRuntimeMetrics.new()


func set_socket_for_tests(socket_ref) -> void:
	socket = socket_ref


func connect_to_server(url: String) -> Error:
	closing_gracefully = false
	closed_notified = false
	socket.handshake_headers = PackedStringArray([
		"Origin: %s" % Constants.MULTIPLAYER_WS_ORIGIN
	])
	var err: Error = socket.connect_to_url(url)
	return err


func poll() -> void:
	socket.poll()

	var state: int = socket.get_ready_state()
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
		var raw_packet: PackedByteArray = socket.get_packet()
		var raw_bytes: int = raw_packet.size()
		var text: String = raw_packet.get_string_from_utf8()
		var decode_result = PacketCodec.decode(text)
		if !decode_result.ok:
			runtime_metrics.observe_decode_failure(raw_bytes)
			ClientLogger.network_event(
				ClientLogger.LEVEL_WARN,
				"packet_decode_failed",
				"Packet decode failed",
				{
					"error": decode_result.error,
					"raw_bytes": raw_bytes,
					"raw_text_length": text.length(),
				}
			)
			packet_parse_failed.emit(text)
			continue

		var packet_type := _packet_type(decode_result.packet)
		runtime_metrics.observe_inbound(raw_bytes, packet_type)
		packet_received.emit(decode_result.packet)


func send_raw_packet(packet: Dictionary) -> void:
	if !is_connected_to_server():
		return

	var packet_type := _packet_type(packet)
	var encode_result = PacketCodec.encode(packet)
	if !encode_result.ok:
		runtime_metrics.observe_encode_failure(packet_type)
		ClientLogger.network_event(
			ClientLogger.LEVEL_WARN,
			"packet_encode_failed",
			"Packet encode failed",
			{
				"error": encode_result.error,
				"packet_type": packet_type,
			}
		)
		return

	var raw_bytes: int = encode_result.wire_message.to_utf8_buffer().size()
	runtime_metrics.observe_outbound(raw_bytes, packet_type)
	socket.send_text(encode_result.wire_message)


func send_authenticate_request(token: String) -> void:
	if token.is_empty():
		return

	send_raw_packet(Packets.authenticate_request_packet(token))


func _packet_type(packet: Dictionary) -> String:
	var packet_type := str(packet.get("type", ""))
	if not packet_type.is_empty():
		return packet_type

	var compact_packet_type := str(packet.get("t", ""))
	if not compact_packet_type.is_empty():
		return compact_packet_type

	return ""


func network_metrics_snapshot() -> Dictionary:
	if runtime_metrics == null:
		return {}
	return runtime_metrics.snapshot()


func close_gracefully() -> void:
	if !begin_graceful_close():
		return

	var elapsed := 0.0
	while socket.get_ready_state() != WebSocketPeer.STATE_CLOSED && elapsed < GRACEFUL_CLOSE_TIMEOUT_SECONDS:
		await get_tree().process_frame
		elapsed += get_process_delta_time()
		socket.poll()


func begin_graceful_close() -> bool:
	var state: int = socket.get_ready_state()
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

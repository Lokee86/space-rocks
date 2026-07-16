extends Node
class_name NetworkClient

const Constants = preload("res://scripts/generated/constants/constants.gd")
const Packets = preload("res://scripts/generated/networking/packets/packets.gd")
const ObservabilityContract = preload("res://scripts/generated/observability/contract_generated.gd")
const NetworkRuntimeMetrics = preload("res://scripts/networking/network_runtime_metrics.gd")

signal connected_to_server
signal connection_closed
signal connection_closed_result(close_code: int, expected: bool)
signal packet_received(data: Dictionary)
signal packet_parse_failed(text: String)

const NORMAL_CLOSE_CODE := 1000
const GRACEFUL_CLOSE_TIMEOUT_SECONDS := 0.25
const PacketCodec = preload("res://scripts/networking/packets/packet_codec.gd")
const ClientLogger = preload("res://scripts/logging/logger.gd")

var socket = WebSocketPeer.new()
var connected := false
var closed_notified := false
var close_result_notified := false
var closing_gracefully := false
var runtime_metrics = NetworkRuntimeMetrics.new()
var _connection_trace_provider: Callable


func set_socket_for_tests(socket_ref) -> void:
	socket = socket_ref


func set_connection_trace_provider(provider: Callable) -> void:
	_connection_trace_provider = provider


func connect_to_server(url: String) -> Error:
	closing_gracefully = false
	closed_notified = false
	close_result_notified = false
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
		if !close_result_notified:
			close_result_notified = true
			connection_closed_result.emit(_close_code(), closing_gracefully)
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
			_emit_packet_failure(
				ObservabilityContract.EVENT_PACKET_DECODE_FAILED,
				decode_result.error_code,
				"invalid_wire_payload",
				"",
				raw_bytes,
				text.length()
			)
			packet_parse_failed.emit(text)
			continue

		var packet_type := _packet_type(decode_result.packet)
		runtime_metrics.observe_inbound(raw_bytes, packet_type)
		packet_received.emit(decode_result.packet)


func send_raw_packet(packet: Dictionary, trace_id: String = "") -> void:
	if !is_connected_to_server():
		return

	var packet_type := _packet_type(packet)
	var encode_result = PacketCodec.encode(packet)
	if !encode_result.ok:
		runtime_metrics.observe_encode_failure(packet_type)
		_emit_packet_failure(
			ObservabilityContract.EVENT_OUTBOUND_PACKET_ENCODE_FAILED,
			encode_result.error_code,
			"invalid_packet_shape",
			packet_type,
			-1,
			-1,
			trace_id
		)
		return

	var raw_bytes: int = encode_result.wire_message.to_utf8_buffer().size()
	runtime_metrics.observe_outbound(raw_bytes, packet_type)
	socket.send_text(encode_result.wire_message)


func send_authenticate_request(token: String, trace_id: String = "") -> void:
	if token.is_empty():
		return

	send_raw_packet(Packets.authenticate_request_packet(token, trace_id), trace_id)


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


func _emit_packet_failure(
	event_name: String,
	error_code: String,
	failure_mode: String,
	packet_type: String,
	raw_byte_count: int = -1,
	raw_text_length: int = -1,
	trace_id: String = ""
) -> void:
	var resolved_trace_id := trace_id if !trace_id.is_empty() else _connection_trace_id()
	if resolved_trace_id.is_empty():
		return
	var fields := {
		"error_code": error_code,
		"failure_mode": failure_mode,
	}
	if !packet_type.is_empty():
		fields["packet_type"] = packet_type
	if raw_byte_count >= 0:
		fields["raw_byte_count"] = raw_byte_count
	if raw_text_length >= 0:
		fields["raw_text_length"] = raw_text_length
	ClientLogger.emit_canonical(event_name, "", {"trace_id": resolved_trace_id}, fields)


func _connection_trace_id() -> String:
	if !_connection_trace_provider.is_valid():
		return ""
	return str(_connection_trace_provider.call())


func _close_code() -> int:
	if socket != null && socket.has_method("get_close_code"):
		return int(socket.get_close_code())
	return 0

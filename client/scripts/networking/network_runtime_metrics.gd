extends RefCounted
class_name NetworkRuntimeMetrics

const DEFAULT_TRANSPORT := "websocket"

var transport: String = DEFAULT_TRANSPORT
var packets_in: int = 0
var packets_out: int = 0
var bytes_in: int = 0
var bytes_out: int = 0
var last_in_packet_bytes: int = 0
var last_out_packet_bytes: int = 0
var max_in_packet_bytes: int = 0
var max_out_packet_bytes: int = 0
var decode_failures: int = 0
var encode_failures: int = 0
var send_failures: int = 0
var last_packet_type_in: String = ""
var last_packet_type_out: String = ""


func observe_inbound(raw_bytes: int, packet_type: String = "") -> void:
	var clamped_bytes: int = _clamp_bytes(raw_bytes)
	packets_in += 1
	bytes_in += clamped_bytes
	last_in_packet_bytes = clamped_bytes
	max_in_packet_bytes = max(max_in_packet_bytes, clamped_bytes)
	last_packet_type_in = packet_type


func observe_outbound(raw_bytes: int, packet_type: String = "") -> void:
	var clamped_bytes: int = _clamp_bytes(raw_bytes)
	packets_out += 1
	bytes_out += clamped_bytes
	last_out_packet_bytes = clamped_bytes
	max_out_packet_bytes = max(max_out_packet_bytes, clamped_bytes)
	last_packet_type_out = packet_type


func observe_decode_failure(raw_bytes: int) -> void:
	_clamp_bytes(raw_bytes)
	decode_failures += 1


func observe_encode_failure(packet_type: String = "") -> void:
	encode_failures += 1
	if not packet_type.is_empty():
		last_packet_type_out = packet_type


func observe_send_failure(raw_bytes: int, packet_type: String = "") -> void:
	var clamped_bytes: int = _clamp_bytes(raw_bytes)
	send_failures += 1
	last_out_packet_bytes = clamped_bytes
	max_out_packet_bytes = max(max_out_packet_bytes, clamped_bytes)
	last_packet_type_out = packet_type


func reset() -> void:
	transport = DEFAULT_TRANSPORT
	packets_in = 0
	packets_out = 0
	bytes_in = 0
	bytes_out = 0
	last_in_packet_bytes = 0
	last_out_packet_bytes = 0
	max_in_packet_bytes = 0
	max_out_packet_bytes = 0
	decode_failures = 0
	encode_failures = 0
	send_failures = 0
	last_packet_type_in = ""
	last_packet_type_out = ""


func snapshot() -> Dictionary:
	return {
		"transport": transport,
		"packets_in": packets_in,
		"packets_out": packets_out,
		"bytes_in": bytes_in,
		"bytes_out": bytes_out,
		"last_in_packet_bytes": last_in_packet_bytes,
		"last_out_packet_bytes": last_out_packet_bytes,
		"max_in_packet_bytes": max_in_packet_bytes,
		"max_out_packet_bytes": max_out_packet_bytes,
		"decode_failures": decode_failures,
		"encode_failures": encode_failures,
		"send_failures": send_failures,
		"last_packet_type_in": last_packet_type_in,
		"last_packet_type_out": last_packet_type_out,
	}


func _clamp_bytes(raw_bytes: int) -> int:
	if raw_bytes < 0:
		return 0
	return raw_bytes

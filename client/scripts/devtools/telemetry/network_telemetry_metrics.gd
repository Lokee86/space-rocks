extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")

var sequence: int = 0
var pending_pings: Dictionary = {}
var latest_rtt_ms: int = -1


func next_ping_packet() -> Dictionary:
	var sent_msec := Time.get_ticks_msec()
	var ping_sequence := sequence
	sequence += 1
	pending_pings[ping_sequence] = sent_msec

	if Packets.has_method("telemetry_ping_packet"):
		return Packets.telemetry_ping_packet(ping_sequence, sent_msec)

	var packet_constants: Dictionary = Packets.get_script_constant_map()
	var field_type := "type"
	if packet_constants.has("FIELD_TYPE"):
		field_type = Packets.FIELD_TYPE
	var type_telemetry_ping := "telemetry_ping"
	if packet_constants.has("TYPE_TELEMETRY_PING"):
		type_telemetry_ping = Packets.TYPE_TELEMETRY_PING
	var field_sequence := "sequence"
	if packet_constants.has("FIELD_SEQUENCE"):
		field_sequence = Packets.FIELD_SEQUENCE
	var field_client_sent_msec := "client_sent_msec"
	if packet_constants.has("FIELD_CLIENT_SENT_MSEC"):
		field_client_sent_msec = Packets.FIELD_CLIENT_SENT_MSEC

	return {
		field_type: type_telemetry_ping,
		field_sequence: ping_sequence,
		field_client_sent_msec: sent_msec,
	}


func apply_pong(packet: Dictionary) -> void:
	var packet_constants: Dictionary = Packets.get_script_constant_map()
	var field_sequence := "sequence"
	if packet_constants.has("FIELD_SEQUENCE"):
		field_sequence = Packets.FIELD_SEQUENCE

	var pong_sequence := int(packet.get(field_sequence, -1))
	if pong_sequence < 0:
		return
	if not pending_pings.has(pong_sequence):
		return

	var sent_msec := int(pending_pings[pong_sequence])
	latest_rtt_ms = Time.get_ticks_msec() - sent_msec
	pending_pings.erase(pong_sequence)


func snapshot() -> Dictionary:
	return {"rtt_ms": latest_rtt_ms}


func reset() -> void:
	sequence = 0
	pending_pings.clear()
	latest_rtt_ms = -1

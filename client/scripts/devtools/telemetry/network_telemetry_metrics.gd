extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")

var sequence: int = 0
var pending_pings: Dictionary = {}
var latest_rtt_ms: int = -1
var latest_server_clock_offset_ms: int = -1


func next_ping_packet() -> Dictionary:
	var sent_msec := Time.get_ticks_msec()
	var ping_sequence := sequence
	sequence += 1
	pending_pings[ping_sequence] = sent_msec

	return {
		Packets.FIELD_TYPE: Packets.TYPE_TELEMETRY_PING,
		Packets.FIELD_SEQUENCE: ping_sequence,
		Packets.FIELD_CLIENT_SENT_MSEC: sent_msec,
	}


func apply_pong(packet: Dictionary) -> void:
	var pong_sequence := int(packet.get(Packets.FIELD_SEQUENCE, -1))
	if pong_sequence < 0:
		return
	if not pending_pings.has(pong_sequence):
		return

	var sent_msec := int(pending_pings[pong_sequence])
	var local_received_msec := Time.get_ticks_msec()
	latest_rtt_ms = local_received_msec - sent_msec

	var server_received_msec := int(packet.get(Packets.FIELD_SERVER_RECEIVED_MSEC, -1))
	var server_sent_msec := int(packet.get(Packets.FIELD_SERVER_SENT_MSEC, -1))
	if server_received_msec > 0 and server_sent_msec > 0:
		var local_midpoint_msec := float(sent_msec) + (float(local_received_msec - sent_msec) / 2.0)
		var server_midpoint_msec := float(server_received_msec) + (float(server_sent_msec - server_received_msec) / 2.0)
		latest_server_clock_offset_ms = int(round(server_midpoint_msec - local_midpoint_msec))
	pending_pings.erase(pong_sequence)


func snapshot() -> Dictionary:
	return {
		"rtt_ms": latest_rtt_ms,
		"server_clock_offset_ms": latest_server_clock_offset_ms,
	}


func reset() -> void:
	sequence = 0
	pending_pings.clear()
	latest_rtt_ms = -1
	latest_server_clock_offset_ms = -1

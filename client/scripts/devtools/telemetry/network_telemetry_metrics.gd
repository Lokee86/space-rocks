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
	latest_rtt_ms = Time.get_ticks_msec() - sent_msec
	pending_pings.erase(pong_sequence)


func snapshot() -> Dictionary:
	return {"rtt_ms": latest_rtt_ms}


func reset() -> void:
	sequence = 0
	pending_pings.clear()
	latest_rtt_ms = -1

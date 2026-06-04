extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func telemetry_ping_packet(sequence: int, client_sent_msec: int) -> Dictionary:
	var packet := {}
	packet[Packets.FIELD_TYPE] = Packets.TYPE_TELEMETRY_PING
	packet[Packets.FIELD_SEQUENCE] = sequence
	packet[Packets.FIELD_CLIENT_SENT_MSEC] = client_sent_msec
	return packet


extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


static func packet_type(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_TYPE, ""))


static func is_room_snapshot(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_SNAPSHOT


static func is_room_state_changed(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_STATE_CHANGED


static func is_room_error(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_ROOM_ERROR


static func is_gameplay_state(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_STATE


static func is_player_pause_state(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_PLAYER_PAUSE_STATE


static func is_telemetry_pong(packet: Dictionary) -> bool:
	return packet_type(packet) == Packets.TYPE_TELEMETRY_PONG


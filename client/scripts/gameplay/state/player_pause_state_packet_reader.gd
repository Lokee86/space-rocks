extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func is_player_pause_state(packet: Dictionary) -> bool:
	return packet.get(Packets.FIELD_TYPE, "") == Packets.TYPE_PLAYER_PAUSE_STATE


static func read(packet: Dictionary) -> Dictionary:
	return {
		"player_id": String(packet.get(Packets.FIELD_PLAYER_ID, "")),
		"paused": bool(packet.get(Packets.FIELD_PAUSED, false)),
	}



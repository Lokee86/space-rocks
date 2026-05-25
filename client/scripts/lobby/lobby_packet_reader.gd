extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")


static func room_code(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_ROOM_CODE, ""))


static func room_state(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_ROOM_STATE, ""))


static func local_member_id(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_LOCAL_MEMBER_ID, ""))


static func owner_id(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_OWNER_ID, ""))


static func max_players(packet: Dictionary) -> int:
	return int(packet.get(Packets.FIELD_MAX_PLAYERS, 0))


static func members(packet: Dictionary) -> Array:
	var value = packet.get(Packets.FIELD_MEMBERS, [])
	if value is Array:
		return value.duplicate(true)
	return []

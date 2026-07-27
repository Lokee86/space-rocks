extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

static func room_code(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_ROOM_CODE, ""))

static func room_state(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_ROOM_STATE, ""))

static func local_player_id(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_LOCAL_PLAYER_ID, ""))

static func owner_id(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_OWNER_ID, ""))

static func max_players(packet: Dictionary) -> int:
	return int(packet.get(Packets.FIELD_MAX_PLAYERS, 0))

static func members(packet: Dictionary) -> Array:
	var value = packet.get(Packets.FIELD_MEMBERS, [])
	return value.duplicate(true) if value is Array else []

static func team_structure(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_TEAM_STRUCTURE, "ffa"))

static func team_assignment_mode(packet: Dictionary) -> String:
	return str(packet.get(Packets.FIELD_TEAM_ASSIGNMENT_MODE, ""))

static func team_count(packet: Dictionary) -> int:
	return int(packet.get(Packets.FIELD_TEAM_COUNT, 0))

static func team_assignments_locked(packet: Dictionary) -> bool:
	return bool(packet.get(Packets.FIELD_TEAM_ASSIGNMENTS_LOCKED, false))

static func member_player_id(member: Dictionary) -> String:
	return str(member.get(Packets.FIELD_PLAYER_ID, ""))

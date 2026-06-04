extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")


static func display_name(member, local_player_id: String) -> String:
	if !(member is Dictionary):
		return str(member)

	var player_id := member_player_id(member)
	var member_name := player_id
	if !local_player_id.is_empty() && player_id == local_player_id:
		return "%s (You)" % member_name
	return member_name


static func member_ready(member) -> bool:
	if member is Dictionary:
		return bool(member.get(Packets.FIELD_READY, member.get(Packets.FIELD_IS_READY, false)))
	return false


static func member_connected(member) -> bool:
	if member is Dictionary:
		return bool(member.get(Packets.FIELD_CONNECTED, member.get(Packets.FIELD_IS_CONNECTED, true)))
	return true


static func is_owner(member, owner_id: String) -> bool:
	if owner_id.is_empty() || !(member is Dictionary):
		return false
	return member_player_id(member) == owner_id


static func is_local_ready(local_player_id: String, members: Array) -> bool:
	if local_player_id.is_empty():
		return false

	for member in members:
		if member is Dictionary && member_player_id(member) == local_player_id:
			return member_ready(member)
	return false


static func member_player_id(member: Dictionary) -> String:
	return str(member.get(Packets.FIELD_PLAYER_ID, ""))


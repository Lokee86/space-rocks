extends RefCounted

const Packets := preload("res://scripts/networking/packets/packets.gd")


static func display_name(member, local_member_id: String) -> String:
	if !(member is Dictionary):
		return str(member)

	var id := member_id(member)
	var player_id := member_player_id(member)
	var member_name := player_id if !player_id.is_empty() else id
	if !local_member_id.is_empty() && id == local_member_id:
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


static func is_local_ready(local_member_id: String, members: Array) -> bool:
	if local_member_id.is_empty():
		return false

	for member in members:
		if member is Dictionary && member_id(member) == local_member_id:
			return member_ready(member)
	return false


static func member_id(member: Dictionary) -> String:
	return str(member.get(Packets.FIELD_MEMBER_ID, member.get(Packets.FIELD_ID, member.get(Packets.FIELD_PLAYER_ID, ""))))


static func member_player_id(member: Dictionary) -> String:
	return str(member.get(Packets.FIELD_PLAYER_ID, ""))

extends RefCounted


static func display_name(member, local_member_id: String) -> String:
	if !(member is Dictionary):
		return str(member)

	var id := member_id(member)
	var member_name := str(member.get("name", member.get("member_name", id)))
	if !local_member_id.is_empty() && id == local_member_id:
		return "%s (You)" % member_name
	return member_name


static func member_ready(member) -> bool:
	if member is Dictionary:
		return bool(member.get("ready", member.get("is_ready", false)))
	return false


static func member_connected(member) -> bool:
	if member is Dictionary:
		return bool(member.get("connected", member.get("is_connected", true)))
	return true


static func is_local_ready(local_member_id: String, members: Array) -> bool:
	if local_member_id.is_empty():
		return false

	for member in members:
		if member is Dictionary && member_id(member) == local_member_id:
			return member_ready(member)
	return false


static func member_id(member: Dictionary) -> String:
	return str(member.get("member_id", member.get("id", member.get("player_id", ""))))

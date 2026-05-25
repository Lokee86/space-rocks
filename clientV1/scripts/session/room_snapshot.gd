extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")


static func members(raw_members: Variant) -> Array:
	var stored_members := []
	if !(raw_members is Array):
		return stored_members

	for raw_member in raw_members:
		if raw_member is Dictionary:
			stored_members.append(raw_member.duplicate(true))
		else:
			stored_members.append(raw_member)

	return stored_members


static func ready_states(members: Array) -> Dictionary:
	var states := {}
	for member in members:
		if !(member is Dictionary):
			continue

		var member_id := str(member.get(Packets.FIELD_MEMBER_ID, member.get(Packets.FIELD_ID, ""))).strip_edges()
		if member_id == "":
			continue

		states[member_id] = bool(member.get(Packets.FIELD_READY, false))

	return states


static func all_connected_members_ready(members: Array) -> bool:
	if members.is_empty():
		return false

	for member in members:
		if !(member is Dictionary):
			return false
		if member.has(Packets.FIELD_CONNECTED) && !bool(member.get(Packets.FIELD_CONNECTED, true)):
			continue
		if !bool(member.get(Packets.FIELD_READY, false)):
			return false

	return true

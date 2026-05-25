extends RefCounted

var room_code := ""
var room_state := ""
var local_member_id := ""
var owner_id := ""
var max_players := 0
var members := []


func clear() -> void:
	room_code = ""
	room_state = ""
	local_member_id = ""
	owner_id = ""
	max_players = 0
	members = []


func apply_snapshot(
	room_code_value: String,
	room_state_value: String,
	local_member_id_value: String,
	owner_id_value: String,
	max_players_value: int,
	members_value: Array
) -> void:
	room_code = room_code_value
	room_state = room_state_value
	local_member_id = local_member_id_value
	owner_id = owner_id_value
	max_players = max_players_value
	members = members_value.duplicate(true)


func summary() -> String:
	return "room=%s state=%s members=%d/%d local=%s owner=%s" % [
		room_code,
		room_state,
		members.size(),
		max_players,
		local_member_id,
		owner_id,
	]


func is_local_owner() -> bool:
	return !local_member_id.is_empty() && local_member_id == owner_id


func all_members_ready() -> bool:
	for member in members:
		if !(member is Dictionary):
			return false
		if !bool(member.get("ready", member.get("is_ready", false))):
			return false
	return !members.is_empty()


func can_start_game() -> bool:
	return is_local_owner() && all_members_ready()

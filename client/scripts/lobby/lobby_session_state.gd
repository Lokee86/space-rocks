class_name LobbySessionState
extends RefCounted

const Packets := preload("res://scripts/generated/networking/packets/packets.gd")

var room_code := ""
var room_state := ""
var local_player_id := ""
var owner_id := ""
var max_players := 0
var members := []
var team_structure := "ffa"
var team_assignment_mode := ""
var team_count := 0
var team_assignments_locked := false


func clear() -> void:
	room_code = ""
	room_state = ""
	local_player_id = ""
	owner_id = ""
	max_players = 0
	members = []
	team_structure = "ffa"
	team_assignment_mode = ""
	team_count = 0
	team_assignments_locked = false


func apply_snapshot(
	room_code_value: String,
	room_state_value: String,
	local_player_id_value: String,
	owner_id_value: String,
	max_players_value: int,
	members_value: Array,
	team_structure_value := "ffa",
	team_assignment_mode_value := "",
	team_count_value := 0,
	team_assignments_locked_value := false
) -> void:
	room_code = room_code_value
	room_state = room_state_value
	local_player_id = local_player_id_value
	owner_id = owner_id_value
	max_players = max_players_value
	members = members_value.duplicate(true)
	team_structure = team_structure_value
	team_assignment_mode = team_assignment_mode_value
	team_count = team_count_value
	team_assignments_locked = team_assignments_locked_value


func summary() -> String:
	return "room=%s state=%s members=%d/%d local=%s owner=%s teams=%s" % [
		room_code, room_state, members.size(), max_players,
		local_player_id, owner_id, team_structure,
	]


func is_local_owner() -> bool:
	return not local_player_id.is_empty() and local_player_id == owner_id


func all_members_ready() -> bool:
	for member in members:
		if not (member is Dictionary):
			return false
		if not bool(member.get(Packets.FIELD_READY, member.get(Packets.FIELD_IS_READY, false))):
			return false
	return not members.is_empty()


func can_start_game() -> bool:
	return is_local_owner() and all_members_ready()

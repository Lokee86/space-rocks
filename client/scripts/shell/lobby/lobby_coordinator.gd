extends RefCounted

const Packets = preload("res://scripts/networking/packets.gd")
const RoomState = preload("res://scripts/session/room_state.gd")
const RoomErrors = preload("res://scripts/session/room_errors.gd")
const RoomSnapshot = preload("res://scripts/session/room_snapshot.gd")

var multiplayer_lobby
var current_room_code := ""
var current_room_state := ""
var local_room_member_id := ""
var room_members := []
var room_ready_states := {}
var room_max_players := 0
var latest_room_snapshot := {}


func configure(lobby_control) -> void:
	multiplayer_lobby = lobby_control


func set_lobby(lobby_control) -> void:
	configure(lobby_control)


func apply_room_snapshot(data: Dictionary) -> String:
	var previous_room_state := current_room_state
	current_room_code = str(data.get(Packets.FIELD_ROOM_CODE, "")).strip_edges()
	current_room_state = str(data.get(Packets.FIELD_ROOM_STATE, "")).strip_edges()
	local_room_member_id = str(data.get(Packets.FIELD_LOCAL_MEMBER_ID, "")).strip_edges()
	room_members = RoomSnapshot.members(data.get(Packets.FIELD_MEMBERS, []))
	room_ready_states = RoomSnapshot.ready_states(room_members)
	room_max_players = int(data.get(Packets.FIELD_MAX_PLAYERS, 0))
	latest_room_snapshot = {
		Packets.FIELD_ROOM_CODE: current_room_code,
		Packets.FIELD_ROOM_STATE: current_room_state,
		Packets.FIELD_LOCAL_MEMBER_ID: local_room_member_id,
		Packets.FIELD_MEMBERS: room_members.duplicate(true),
		"ready_states": room_ready_states.duplicate(true),
		Packets.FIELD_MAX_PLAYERS: room_max_players,
	}

	return previous_room_state


func apply_room_state_changed(data: Dictionary) -> void:
	current_room_code = str(data.get(Packets.FIELD_ROOM_CODE, current_room_code)).strip_edges()
	current_room_state = str(data.get(Packets.FIELD_ROOM_STATE, current_room_state)).strip_edges()
	latest_room_snapshot[Packets.FIELD_ROOM_CODE] = current_room_code
	latest_room_snapshot[Packets.FIELD_ROOM_STATE] = current_room_state


func set_status(text: String) -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return
	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status(text)


func local_member_ready() -> bool:
	if local_room_member_id == "":
		return false

	return bool(room_ready_states.get(local_room_member_id, false))


func room_state_is_lobby() -> bool:
	return RoomState.is_lobby(current_room_state)


func room_state_is_in_game() -> bool:
	return RoomState.is_in_game(current_room_state)


func room_status_text() -> String:
	var status := current_room_state
	if status == "":
		status = "Unknown"

	if room_max_players > 0:
		return "%s (%d/%d)" % [status, room_members.size(), room_max_players]

	return status


func update_room_labels() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return

	if multiplayer_lobby.has_method("set_room_code"):
		multiplayer_lobby.set_room_code(current_room_code)
	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status(room_status_text())


func update_member_rows() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return
	if !multiplayer_lobby.has_method("set_members"):
		return

	multiplayer_lobby.set_members(room_members)


func update_control_state() -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return

	var local_ready := local_member_ready()
	if multiplayer_lobby.has_method("set_local_ready"):
		multiplayer_lobby.set_local_ready(local_ready)
	if multiplayer_lobby.has_method("set_start_enabled"):
		multiplayer_lobby.set_start_enabled(
			room_state_is_lobby() && local_ready && RoomSnapshot.all_connected_members_ready(room_members)
		)

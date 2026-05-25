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
var create_room_request_pending := false
var start_single_player_request_pending := false
var pending_join_room_code := ""


func configure(lobby_control) -> void:
	multiplayer_lobby = lobby_control


func set_lobby(lobby_control) -> void:
	configure(lobby_control)


func begin_create_room_flow() -> void:
	create_room_request_pending = true
	start_single_player_request_pending = false
	pending_join_room_code = ""


func begin_join_room_flow(room_code: String) -> void:
	pending_join_room_code = room_code.strip_edges()
	create_room_request_pending = false
	start_single_player_request_pending = false


func begin_single_player_flow() -> void:
	start_single_player_request_pending = true
	create_room_request_pending = false
	pending_join_room_code = ""


func begin_create_room_connection_flow(network_connected: bool) -> Dictionary:
	begin_create_room_flow()
	var result := _connection_flow_result()
	result["status"] = "Connecting..."
	if network_connected:
		result["should_send_create_room"] = true
	else:
		result["should_connect"] = true
	return result


func begin_join_room_connection_flow(room_code: String, network_connected: bool) -> Dictionary:
	begin_join_room_flow(room_code)
	var result := _connection_flow_result()
	result["status"] = "Connecting..."
	if pending_join_room_code == "":
		result["status"] = "Enter a room code to join."
		return result
	if network_connected:
		result["should_send_join_room"] = true
	else:
		result["should_connect"] = true
	return result


func begin_single_player_connection_flow(network_connected: bool) -> Dictionary:
	begin_single_player_flow()
	var result := _connection_flow_result()
	if network_connected:
		result["should_send_single_player"] = true
	else:
		result["should_connect"] = true
	return result


func _connection_flow_result() -> Dictionary:
	return {
		"status": "",
		"should_connect": false,
		"should_send_create_room": false,
		"should_send_join_room": false,
		"should_send_single_player": false,
	}


func clear_pending_requests() -> void:
	create_room_request_pending = false
	start_single_player_request_pending = false
	pending_join_room_code = ""


func handle_create_room_connection_failed() -> void:
	create_room_request_pending = false
	set_status("Could not connect to server.")


func handle_join_room_connection_failed() -> void:
	pending_join_room_code = ""
	set_status("Could not connect to server.")


func handle_single_player_connection_failed() -> void:
	start_single_player_request_pending = false


func take_pending_create_room_request() -> bool:
	if !create_room_request_pending:
		return false

	create_room_request_pending = false
	return true


func take_pending_start_single_player_request() -> bool:
	if !start_single_player_request_pending:
		return false

	start_single_player_request_pending = false
	return true


func has_pending_join_room_code() -> bool:
	return pending_join_room_code != ""


func take_pending_join_room_code() -> String:
	var room_code := pending_join_room_code
	pending_join_room_code = ""
	return room_code


func handle_lobby_packet(data: Dictionary) -> Dictionary:
	var packet_type := str(data.get(Packets.FIELD_TYPE, ""))
	if packet_type == Packets.TYPE_ROOM_ERROR:
		return {
			"handled": true,
			"kind": "room_error",
			"message": room_error_message(data),
		}
	if packet_type == Packets.TYPE_ROOM_SNAPSHOT:
		var snapshot_result: Dictionary = apply_room_snapshot_result(data)
		update_room_labels()
		update_member_rows()
		update_control_state()
		return {
			"handled": true,
			"kind": "room_snapshot",
			"previous_room_state": snapshot_result.get("previous_room_state", ""),
		}
	if packet_type == Packets.TYPE_ROOM_STATE_CHANGED:
		apply_room_state_changed_result(data)
		update_room_labels()
		update_control_state()
		return {
			"handled": true,
			"kind": "room_state_changed",
		}

	return {
		"handled": false,
		"kind": "",
	}


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


func apply_room_snapshot_result(data: Dictionary) -> Dictionary:
	var previous_room_state := apply_room_snapshot(data)
	return _room_update_result(previous_room_state)


func apply_room_state_changed(data: Dictionary) -> void:
	current_room_code = str(data.get(Packets.FIELD_ROOM_CODE, current_room_code)).strip_edges()
	current_room_state = str(data.get(Packets.FIELD_ROOM_STATE, current_room_state)).strip_edges()
	latest_room_snapshot[Packets.FIELD_ROOM_CODE] = current_room_code
	latest_room_snapshot[Packets.FIELD_ROOM_STATE] = current_room_state


func apply_room_state_changed_result(data: Dictionary) -> Dictionary:
	var previous_room_state := current_room_state
	apply_room_state_changed(data)
	return _room_update_result(previous_room_state)


func _room_update_result(previous_room_state: String) -> Dictionary:
	return {
		"previous_room_state": previous_room_state,
		"room_state_is_lobby": room_state_is_lobby(),
		"room_state_is_in_game": room_state_is_in_game(),
	}


func set_status(text: String) -> void:
	if multiplayer_lobby == null || !is_instance_valid(multiplayer_lobby):
		return
	if multiplayer_lobby.has_method("set_status"):
		multiplayer_lobby.set_status(text)


func room_error_message(data: Dictionary) -> String:
	return RoomErrors.message_for_packet(data)


func show_room_error_status(message: String) -> void:
	set_status(message)


func local_member_ready() -> bool:
	if local_room_member_id == "":
		return false

	return bool(room_ready_states.get(local_room_member_id, false))


func should_send_ready_toggle(network_connected: bool) -> bool:
	if !network_connected:
		set_status("Not connected to server.")
		return false

	return true


func next_ready_value() -> bool:
	return !local_member_ready()


func should_send_start_game(network_connected: bool) -> bool:
	if !network_connected:
		set_status("Not connected to server.")
		return false
	if !room_state_is_lobby() || !local_member_ready():
		set_status("Ready up before starting.")
		return false

	return true


func room_code() -> String:
	return current_room_code


func room_state() -> String:
	return current_room_state


func latest_snapshot() -> Dictionary:
	return latest_room_snapshot


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

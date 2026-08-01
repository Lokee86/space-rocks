extends RefCounted

const Packets = preload("res://scripts/generated/networking/packets/packets.gd")


static func create_room_request_packet(
		trace_id := "",
		local_profile_id := "",
		team_structure := "ffa",
		team_assignment_mode := "",
		team_count := 0,
		max_players := 0,
		preset_id := "arcade_survival",
		starting_lives := 0,
		infinite_lives := false,
		target_score := 0,
		target_kills := 0
) -> Dictionary:
	return Packets.create_room_request_packet(
		trace_id,
		local_profile_id,
		team_structure,
		team_assignment_mode,
		team_count,
		max_players,
		preset_id,
		starting_lives,
		infinite_lives,
		target_score,
		target_kills
	)


static func join_room_request_packet(room_code, trace_id := "") -> Dictionary:
	return Packets.join_room_request_packet(room_code, trace_id)


static func leave_room_request_packet() -> Dictionary:
	return Packets.leave_room_request_packet()


static func set_ready_request_packet(ready) -> Dictionary:
	return Packets.set_ready_request_packet(ready)


static func set_team_assignment_request_packet(target_player_id: String, team_id: String) -> Dictionary:
	return Packets.set_team_assignment_request_packet(target_player_id, team_id)


static func start_game_request_packet() -> Dictionary:
	return Packets.start_game_request_packet()


static func add_bot_request_packet() -> Dictionary:
	return Packets.add_bot_request_packet()


static func remove_room_member_request_packet(player_id: String) -> Dictionary:
	return Packets.remove_room_member_request_packet(player_id)


static func start_single_player_request_packet(
		local_profile_id := "",
		trace_id := "",
		preset_id := "arcade_survival",
		starting_lives := 0,
		infinite_lives := false,
		target_score := 0,
		target_kills := 0
) -> Dictionary:
	return Packets.start_single_player_request_packet(
		local_profile_id,
		trace_id,
		preset_id,
		starting_lives,
		infinite_lives,
		target_score,
		target_kills
	)


static func return_to_lobby_request_packet() -> Dictionary:
	return Packets.return_to_lobby_request_packet()

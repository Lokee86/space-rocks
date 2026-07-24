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
		target_score := 0
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
		target_score
	)


static func join_room_request_packet(room_code, trace_id := "") -> Dictionary:
	return Packets.join_room_request_packet(room_code, trace_id)


static func leave_room_request_packet() -> Dictionary:
	return Packets.leave_room_request_packet()


static func set_ready_request_packet(ready) -> Dictionary:
	return Packets.set_ready_request_packet(ready)


static func set_team_assignment_request_packet(target_player_id: String, team_id: String) -> Dictionary:
	return Packets.set_team_assignment_request_packet(target_player_id, team_id)


static func loadout_options_request_packet(trace_id: String, local_profile_id: String, play_mode: String, mode_id: String) -> Dictionary:
	return Packets.loadout_options_request_packet(trace_id, local_profile_id, play_mode, mode_id)


static func set_loadout_request_packet(trace_id: String, selection: Dictionary) -> Dictionary:
	return Packets.set_loadout_request_packet(
		trace_id,
		str(selection.get("selected_owned_ship_id", "")),
		selection.get("selected_weapons_by_point", {}),
		selection.get("selected_modules_by_slot", {}),
		selection.get("starting_ammo_by_point", {})
	)


static func start_game_request_packet() -> Dictionary:
	return Packets.start_game_request_packet()


static func add_bot_request_packet() -> Dictionary:
	return Packets.add_bot_request_packet()


static func remove_room_member_request_packet(player_id: String) -> Dictionary:
	return Packets.remove_room_member_request_packet(player_id)


static func start_single_player_request_packet(local_profile_id := "", trace_id := "") -> Dictionary:
	return Packets.start_single_player_request_packet(local_profile_id, trace_id)


static func return_to_lobby_request_packet() -> Dictionary:
	return Packets.return_to_lobby_request_packet()

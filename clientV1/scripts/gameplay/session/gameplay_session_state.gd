extends RefCounted

const RoomState = preload("res://scripts/session/room_state.gd")


static func is_multiplayer_session(session_mode: String) -> bool:
	return session_mode.strip_edges().to_lower() == "multiplayer"


static func is_room_in_game(room_state: String) -> bool:
	return RoomState.is_in_game(room_state)


static func is_room_game_over(room_state: String) -> bool:
	return RoomState.is_game_over(room_state)


static func can_process_gameplay_packets(room_state: String) -> bool:
	if room_state == "":
		return true

	return is_room_in_game(room_state) || is_room_game_over(room_state)


static func is_game_over(session_mode: String, room_state: String, hud_is_game_over: bool) -> bool:
	if hud_is_game_over:
		return true

	return is_multiplayer_session(session_mode) && is_room_game_over(room_state)

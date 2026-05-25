extends RefCounted

const GameplaySessionState = preload("res://scripts/gameplay/session/gameplay_session_state.gd")


static func is_cycle_view_available(
	session_mode: String,
	room_state: String,
	hud_is_game_over: bool,
	is_spectating: bool
) -> bool:
	return (
		GameplaySessionState.is_multiplayer_session(session_mode) &&
		hud_is_game_over &&
		is_spectating &&
		!GameplaySessionState.is_room_game_over(room_state)
	)

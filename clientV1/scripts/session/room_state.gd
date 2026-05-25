extends RefCounted

const LOBBY := "Lobby"
const LOBBY_LOWER := "lobby"
const IN_GAME := "InGame"
const IN_GAME_SNAKE := "in_game"
const GAME_OVER := "GameOver"
const GAME_OVER_SNAKE := "game_over"
const GAME_OVER_COMPACT := "gameover"


static func normalize(room_state: String) -> String:
	var trimmed := room_state.strip_edges()
	if is_lobby(trimmed):
		return LOBBY
	if is_in_game(trimmed):
		return IN_GAME
	if is_game_over(trimmed):
		return GAME_OVER
	return trimmed


static func is_lobby(room_state: String) -> bool:
	var trimmed := room_state.strip_edges()
	return trimmed == LOBBY || trimmed == LOBBY_LOWER


static func is_in_game(room_state: String) -> bool:
	var trimmed := room_state.strip_edges()
	return trimmed == IN_GAME || trimmed == IN_GAME_SNAKE


static func is_game_over(room_state: String) -> bool:
	var trimmed := room_state.strip_edges()
	return trimmed == GAME_OVER || trimmed == GAME_OVER_SNAKE || trimmed == GAME_OVER_COMPACT

extends RefCounted

const ShellState := preload("res://scripts/shell/shell_state.gd")


static func from_room_state(room_state: String) -> String:
	match room_state:
		"Lobby":
			return ShellState.LOBBY
		"Starting":
			return ShellState.LOBBY
		"InGame":
			return ShellState.GAMEPLAY
		"GameOver":
			return ShellState.GAME_OVER
		"Closed":
			return ShellState.MAIN_MENU
		_:
			return ShellState.LOBBY

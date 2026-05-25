extends RefCounted

const ShellState := preload("res://scripts/shell/shell_state.gd")
const ShellConstants := preload("res://scripts/shell/constants.gd")


static func from_room_state(room_state: String) -> String:
	match room_state:
		ShellConstants.ROOM_STATE_LOBBY:
			return ShellState.LOBBY
		ShellConstants.ROOM_STATE_STARTING:
			return ShellState.LOBBY
		ShellConstants.ROOM_STATE_IN_GAME:
			return ShellState.GAMEPLAY
		ShellConstants.ROOM_STATE_GAME_OVER:
			return ShellState.GAME_OVER
		ShellConstants.ROOM_STATE_CLOSED:
			return ShellState.MAIN_MENU
		_:
			return ShellState.LOBBY

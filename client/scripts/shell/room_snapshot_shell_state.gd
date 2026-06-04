extends RefCounted

const Constants := preload("res://scripts/generated/constants/constants.gd")


static func from_room_state(room_state: String) -> String:
	match room_state:
		Constants.ROOM_STATE_LOBBY:
			return Constants.SHELL_STATE_LOBBY
		Constants.ROOM_STATE_STARTING:
			return Constants.SHELL_STATE_LOBBY
		Constants.ROOM_STATE_IN_GAME:
			return Constants.SHELL_STATE_GAMEPLAY
		Constants.ROOM_STATE_GAME_OVER:
			return Constants.SHELL_STATE_GAME_OVER
		Constants.ROOM_STATE_CLOSED:
			return Constants.SHELL_STATE_MAIN_MENU
		_:
			return Constants.SHELL_STATE_LOBBY


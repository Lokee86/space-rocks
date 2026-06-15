extends RefCounted
class_name SessionNetworkTarget

const Constants := preload("res://scripts/generated/constants/constants.gd")


static func websocket_url_for_mode(mode: String) -> String:
	match mode:
		Constants.SESSION_MODE_SINGLE_PLAYER:
			return Constants.SINGLE_PLAYER_WS_URL
		Constants.SESSION_MODE_MULTIPLAYER:
			return Constants.MULTIPLAYER_WS_URL
		_:
			return ""

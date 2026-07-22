extends RefCounted
class_name SessionNetworkTarget

const Constants := preload("res://scripts/generated/constants/constants.gd")
const LOCAL_SERVER_PORT_ENV := "SPACE_ROCKS_LOCAL_SERVER_PORT"


static func websocket_url_for_mode(mode: String) -> String:
	match mode:
		Constants.SESSION_MODE_SINGLE_PLAYER:
			return _single_player_websocket_url()
		Constants.SESSION_MODE_MULTIPLAYER:
			return Constants.MULTIPLAYER_WS_URL
		_:
			return ""


static func _single_player_websocket_url() -> String:
	var port := OS.get_environment(LOCAL_SERVER_PORT_ENV).strip_edges()
	if port.is_valid_int():
		var parsed_port := port.to_int()
		if parsed_port >= 1 && parsed_port <= 65535:
			return "ws://127.0.0.1:%d/ws" % parsed_port
	return Constants.SINGLE_PLAYER_WS_URL
